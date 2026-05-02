/**
 * Discovers agent definitions from:
 *   1. Bob project format:  <cwd>/agents/<name>/SKILL.md   (Bob's layout)
 *   2. Pi project format:   <cwd>/.pi/agents/<name>.md
 *   3. Pi user format:      ~/.pi/agent/agents/<name>.md
 *
 * Bob's SKILL.md uses Claude Code tool names (Read, Write, Bash, Glob, Task,
 * TaskList, etc.).  We map those to pi equivalents when building the tool list
 * for each spawned session.
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { getAgentDir, parseFrontmatter } from "@mariozechner/pi-coding-agent";

export type AgentSource = "bob-project" | "pi-project" | "pi-user";

export interface AgentDef {
  name: string;
  description: string;
  /** Raw tool names from SKILL.md frontmatter; mapped to pi built-in name allowlist via buildBuiltinTools(). */
  tools?: string[];
  model?: string;
  systemPrompt: string;
  source: AgentSource;
  filePath: string;
}

// ─── Tool name mapping: Claude Code → pi ─────────────────────────────────────

/**
 * Claude Code uses PascalCase tool names; pi uses lowercase.
 * Some CC tools (Glob) map to a different pi tool (find).
 * Coordination tools (Task, TaskList, …) are provided by this extension.
 */
const CC_TO_PI: Record<string, string> = {
  Read: "read",
  Write: "write",
  Edit: "edit",
  Bash: "bash",
  Glob: "find",
  Grep: "grep",
  LS: "ls",
  // Task and TaskXxx are provided as custom tools by the extension.
  // We still list them so callers know they were requested.
  Task: "subagent",
  TaskCreate: "TaskCreate",
  TaskList: "TaskList",
  TaskGet: "TaskGet",
  TaskUpdate: "TaskUpdate",
};

/** Normalise a tool name from a SKILL.md frontmatter list. */
function normaliseTool(name: string): string {
  return CC_TO_PI[name.trim()] ?? name.trim().toLowerCase();
}

// ─── Pi built-in tool name allowlist ─────────────────────────────────────────

/** All pi built-in tool names that can be passed to createAgentSession({ tools: string[] }). */
const PI_BUILTIN_NAMES = new Set(["read", "write", "edit", "bash", "find", "grep", "ls"]);

/** Default tool set when no explicit list is declared in SKILL.md. */
const ALL_DEFAULTS: string[] = ["read", "write", "edit", "bash", "find", "grep", "ls"];

/**
 * Build the pi built-in tool name allowlist for a given set of raw tool names
 * from a SKILL.md frontmatter.  Returns string[] suitable for passing directly
 * to createAgentSession({ tools: string[] }).
 *
 * Coordination tools (Task, TaskCreate, TaskList, TaskGet, TaskUpdate, subagent)
 * are provided by the extension as customTools and are silently filtered out here.
 */
export function buildBuiltinTools(rawTools: string[] | undefined): string[] {
  if (!rawTools || rawTools.length === 0) return ALL_DEFAULTS;

  const seen = new Set<string>();
  const result: string[] = [];

  for (const raw of rawTools) {
    const piName = normaliseTool(raw);
    if (PI_BUILTIN_NAMES.has(piName) && !seen.has(piName)) {
      seen.add(piName);
      result.push(piName);
    }
    // Non-builtin names (subagent, TaskCreate, etc.) are silently skipped —
    // they are provided as customTools by the extension.
  }

  // Always include at least read so agents can function
  return result.length > 0 ? result : ["read"];
}

// ─── Agent loading ────────────────────────────────────────────────────────────

function tryLoadSkill(filePath: string, source: AgentSource): AgentDef | undefined {
  if (!fs.existsSync(filePath)) return undefined;
  let content: string;
  try {
    content = fs.readFileSync(filePath, "utf-8");
  } catch {
    return undefined;
  }

  const { frontmatter, body } = parseFrontmatter<Record<string, string>>(content);
  if (!frontmatter.name) return undefined;

  const rawTools = frontmatter.tools
    ?.split(",")
    .map((t: string) => t.trim())
    .filter(Boolean);

  return {
    name: frontmatter.name,
    description: frontmatter.description ?? "",
    tools: rawTools,
    model: frontmatter.model,
    systemPrompt: body.trim(),
    source,
    filePath,
  };
}

function loadFromDir(dir: string, source: AgentSource): AgentDef[] {
  if (!fs.existsSync(dir)) return [];
  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return [];
  }

  const defs: AgentDef[] = [];

  for (const entry of entries) {
    const entryPath = path.join(dir, entry.name);

    if (entry.isDirectory()) {
      // Bob format: agents/<name>/SKILL.md
      const skill = tryLoadSkill(path.join(entryPath, "SKILL.md"), source);
      if (skill) defs.push(skill);
    } else if ((entry.isFile() || entry.isSymbolicLink()) && entry.name.endsWith(".md")) {
      // Pi flat format: .pi/agents/<name>.md
      const skill = tryLoadSkill(entryPath, source);
      if (skill) defs.push(skill);
    }
  }

  return defs;
}

export interface AgentDiscovery {
  agents: AgentDef[];
  /** Directories that were searched. */
  searchedDirs: string[];
}

/**
 * Discover all available agents.
 *
 * Search order (later sources override earlier ones with the same name):
 *   1. pi user-level:    ~/.pi/agent/agents/
 *   2. pi project-level: <cwd>/.pi/agents/
 *   3. Bob project:      <cwd>/agents/   ← highest priority so Bob's defs win
 */
export function discoverAgents(cwd: string): AgentDiscovery {
  const userDir = path.join(getAgentDir(), "agents");
  const piProjectDir = path.join(cwd, ".pi", "agents");
  const bobProjectDir = path.join(cwd, "agents");

  const searchedDirs = [userDir, piProjectDir, bobProjectDir];
  const byName = new Map<string, AgentDef>();

  for (const [dir, source] of [
    [userDir, "pi-user"],
    [piProjectDir, "pi-project"],
    [bobProjectDir, "bob-project"],
  ] as const) {
    for (const def of loadFromDir(dir, source)) {
      byName.set(def.name, def); // later entries override earlier
    }
  }

  return { agents: Array.from(byName.values()), searchedDirs };
}
