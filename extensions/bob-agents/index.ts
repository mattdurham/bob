/**
 * Bob Agents Extension
 *
 * Provides in-process subagent spawning with a shared in-memory message bus
 * and task board, so Bob's existing agent definitions (agents/*\/SKILL.md)
 * work in pi without modification.
 *
 * Registered tools (available to the orchestrating LLM):
 *   subagent        — spawn agents (single / parallel / chain)
 *   agent_status    — list running and finished agents (with live stdout tail)
 *   agent_output    — get full stdout buffer for a specific agent
 *   mailbox_read    — read the orchestrator's mailbox
 *   mailbox_send_as — send a message to an agent from the orchestrator
 *   mailbox_broadcast — broadcast a message to all active agents
 *   TaskCreate      — create a task on the shared board
 *   TaskList        — list tasks (optionally filtered)
 *   TaskGet         — get a task by id
 *   TaskUpdate      — update task status / owner / notes
 *
 * Custom tools injected into every spawned agent (bound to that agent's name):
 *   mailbox_receive — read own unread messages
 *   mailbox_send    — send a message to another agent or "orchestrator"
 *   mailbox_broadcast — broadcast from own name to all active agents
 *   agent_status    — same shared view of all agents
 *   TaskCreate / TaskList / TaskGet / TaskUpdate — same shared task board
 *
 * Module-level singletons mean every agent session — including recursively
 * spawned sub-agents — shares the same bus, registry, and task board.
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { StringEnum } from "@mariozechner/pi-ai";
import {
  AuthStorage,
  DefaultResourceLoader,
  ModelRegistry,
  SessionManager,
  SettingsManager,
  createAgentSession,
  defineTool,
} from "@mariozechner/pi-coding-agent";
import { type ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { Type } from "typebox";
import { TeamManager, ROOT_TEAM, type TeamContext } from "./agent-registry.js";
import { type AgentDef, buildBuiltinTools, discoverAgents, getAgentDir } from "./agent-loader.js";
import { registerTeamCommands } from "./team-ui.js";
import otelExtension from "../otel.js";

// ─── Module-level singletons (shared across all sessions in this process) ─────

const teams = new TeamManager();

// Tracks live AgentSession objects so we can abort on shutdown
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const liveSessions = new Map<string, any>();

const authStorage = AuthStorage.create();
const modelRegistry = new ModelRegistry(authStorage);

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getFinalOutput(session: { messages: { role: string; content: { type: string; text?: string }[] }[] }): string {
  for (let i = session.messages.length - 1; i >= 0; i--) {
    const msg = session.messages[i];
    if (msg.role === "assistant") {
      for (const part of msg.content) {
        if (part.type === "text" && part.text) return part.text;
      }
    }
  }
  return "";
}

function resolveModel(modelId: string) {
  // Use synchronous find() — getAvailable() does async API calls and can hang
  try {
    return modelRegistry.find(undefined, modelId) ?? undefined;
  } catch {
    return undefined;
  }
}

function formatAgentStatus(all: { name: string; role: string; status: string; spawnedAt: number; finishedAt?: number; model?: string; stdout: string; stdoutBytes: number; team: string }[]): string {
  if (all.length === 0) return "No agents have been spawned yet.";
  return all
    .map((a) => {
      const elapsed = a.finishedAt
        ? `${((a.finishedAt - a.spawnedAt) / 1000).toFixed(1)}s`
        : `${((Date.now() - a.spawnedAt) / 1000).toFixed(1)}s`;
      const statusIcon = { spawning: "⏳", running: "▶", done: "✓", error: "✗", aborted: "⊘" }[a.status] ?? "?";
      const tail = a.status === "running" && a.stdout
        ? `\n    ${a.stdout.split("\n").slice(-3).join("\n    ").slice(0, 200)}`
        : "";
      const overflow = a.stdoutBytes > a.stdout.length ? ` [+${a.stdoutBytes - a.stdout.length}b truncated]` : "";
      return `${statusIcon} ${a.name} [${a.role}] ${a.status} ${elapsed}${a.model ? ` (${a.model})` : ""}${overflow}${tail}`;
    })
    .join("\n");
}

// ─── Task board tool definitions (shared between orchestrator and agents) ─────

function makeTaskTools(teamCtx: TeamContext) {
  const { taskBoard } = teamCtx;
  const TaskCreate = defineTool({
    name: "TaskCreate",
    label: "Task Create",
    description: "Create a task on the shared task board",
    parameters: Type.Object({
      title: Type.String({ description: "Short task title" }),
      description: Type.String({ description: "Full task description" }),
      owner: Type.Optional(Type.String({ description: "Agent name to assign the task to" })),
      tags: Type.Optional(Type.Array(Type.String(), { description: "Tags for filtering" })),
      dependencies: Type.Optional(Type.Array(Type.String(), { description: "Task IDs that must complete first" })),
    }),
    async execute(_id, params) {
      const task = taskBoard.create(params.title, params.description, {
        owner: params.owner,
        tags: params.tags,
        dependencies: params.dependencies,
      });
      return {
        content: [{ type: "text" as const, text: `Created task ${task.id}: ${task.title}` }],
        details: { task },
      };
    },
  });

  const TaskList = defineTool({
    name: "TaskList",
    label: "Task List",
    description: "List tasks on the shared board. Optionally filter by status or owner.",
    parameters: Type.Object({
      status: Type.Optional(
        StringEnum(["todo", "in_progress", "done", "blocked", "failed"] as const, {
          description: "Filter by status",
        }),
      ),
      owner: Type.Optional(Type.String({ description: "Filter by owner agent name" })),
      tag: Type.Optional(Type.String({ description: "Filter by tag" })),
    }),
    async execute(_id, params) {
      const tasks = taskBoard.list({
        status: params.status as Parameters<typeof taskBoard.list>[0]["status"],
        owner: params.owner,
        tag: params.tag,
      });
      if (tasks.length === 0) return { content: [{ type: "text" as const, text: "No tasks found." }], details: { tasks: [] } };
      const lines = tasks.map(
        (t) =>
          `[${t.id}] ${t.status.padEnd(12)} ${t.owner ? `@${t.owner} ` : ""}${t.title}`,
      );
      return {
        content: [{ type: "text" as const, text: lines.join("\n") }],
        details: { tasks },
      };
    },
  });

  const TaskGet = defineTool({
    name: "TaskGet",
    label: "Task Get",
    description: "Get full details of a task by its ID",
    parameters: Type.Object({
      id: Type.String({ description: "Task ID (e.g. task-001)" }),
    }),
    async execute(_id, params) {
      const task = taskBoard.get(params.id);
      if (!task) {
        return {
          content: [{ type: "text" as const, text: `Task ${params.id} not found.` }],
          details: {},
          isError: true,
        };
      }
      const lines = [
        `ID:          ${task.id}`,
        `Title:       ${task.title}`,
        `Status:      ${task.status}`,
        `Owner:       ${task.owner ?? "(unassigned)"}`,
        `Tags:        ${task.tags?.join(", ") ?? "none"}`,
        `Depends on:  ${task.dependencies?.join(", ") ?? "none"}`,
        ``,
        task.description,
        ...(task.notes ? [`\nNotes:\n${task.notes}`] : []),
      ];
      return {
        content: [{ type: "text" as const, text: lines.join("\n") }],
        details: { task },
      };
    },
  });

  const TaskUpdate = defineTool({
    name: "TaskUpdate",
    label: "Task Update",
    description: "Update a task's status, owner, or notes",
    parameters: Type.Object({
      id: Type.String({ description: "Task ID to update" }),
      status: Type.Optional(
        StringEnum(["todo", "in_progress", "done", "blocked", "failed"] as const, {
          description: "New status",
        }),
      ),
      owner: Type.Optional(Type.String({ description: "Assign to this agent name" })),
      notes: Type.Optional(Type.String({ description: "Progress notes or findings" })),
    }),
    async execute(_id, params) {
      const task = taskBoard.update(params.id, {
        status: params.status as Parameters<typeof taskBoard.update>[1]["status"],
        owner: params.owner,
        notes: params.notes,
      });
      if (!task) {
        return {
          content: [{ type: "text" as const, text: `Task ${params.id} not found.` }],
          details: {},
          isError: true,
        };
      }
      return {
        content: [{ type: "text" as const, text: `Updated ${task.id}: status=${task.status}${task.owner ? ` owner=${task.owner}` : ""}` }],
        details: { task },
      };
    },
  });

  return [TaskCreate, TaskList, TaskGet, TaskUpdate];
}

// ─── Agent-bound tools (injected per spawned session) ─────────────────────────

function makeBoundTools(agentName: string, teamCtx: TeamContext) {
  const { bus, registry } = teamCtx;
  const AgentStatus = defineTool({
    name: "agent_status",
    label: "Agent Status",
    description: "List all agents and their current status",
    parameters: Type.Object({}),
    async execute() {
      const allAgents = teams.getAll().flatMap((t) => t.registry.getAll());
      return {
        content: [{ type: "text" as const, text: formatAgentStatus(allAgents) }],
        details: { agents: allAgents },
      };
    },
  });

  const MailboxReceive = defineTool({
    name: "mailbox_receive",
    label: "Mailbox Receive",
    description: "Read unread messages in your mailbox",
    parameters: Type.Object({}),
    async execute() {
      const msgs = bus.receive(agentName);
      if (msgs.length === 0) {
        return { content: [{ type: "text" as const, text: "No new messages." }], details: { messages: [] } };
      }
      const formatted = msgs
        .map((m) => `From: ${m.from} [${new Date(m.timestamp).toISOString()}]\n${m.content}`)
        .join("\n---\n");
      return {
        content: [{ type: "text" as const, text: formatted }],
        details: { messages: msgs },
      };
    },
  });

  const MailboxSend = defineTool({
    name: "mailbox_send",
    label: "Mailbox Send",
    description: 'Send a message to another agent or to "orchestrator"',
    parameters: Type.Object({
      to: Type.String({ description: 'Recipient agent name or "orchestrator"' }),
      content: Type.String({ description: "Message content" }),
    }),
    async execute(_id, params) {
      const msg = bus.send(agentName, params.to, params.content);
      return {
        content: [{ type: "text" as const, text: `Message sent to ${params.to} (id: ${msg.id})` }],
        details: { messageId: msg.id },
      };
    },
  });

  const MailboxBroadcast = defineTool({
    name: "mailbox_broadcast",
    label: "Mailbox Broadcast",
    description: "Broadcast a message to all active agents",
    parameters: Type.Object({
      content: Type.String({ description: "Message to broadcast" }),
    }),
    async execute(_id, params) {
      const recipients = registry.activeNames().filter((n) => n !== agentName);
      bus.broadcast(agentName, params.content, recipients);
      return {
        content: [{ type: "text" as const, text: `Broadcast sent to ${recipients.length} agents: ${recipients.join(", ")}` }],
        details: { recipients },
      };
    },
  });

  const AgentTeam = defineTool({
    name: "agent_team",
    label: "Agent Team",
    description: "List your teammates — agents in the same team, their roles, and current status.",
    parameters: Type.Object({}),
    async execute() {
      const members = teamCtx.registry.getAll().filter((a) => a.name !== agentName);
      if (members.length === 0) {
        return {
          content: [{ type: "text" as const, text: `You are the only member of team "${teamCtx.name}". Team lead: ${teamCtx.lead}.` }],
          details: { team: teamCtx.name, lead: teamCtx.lead, teammates: [] },
        };
      }
      const lines = [
        `Team: ${teamCtx.name} | Lead: ${teamCtx.lead}`,
        ...members.map((m) => {
          const elapsed = m.finishedAt
            ? `${((m.finishedAt - m.spawnedAt) / 1000).toFixed(1)}s`
            : `${((Date.now() - m.spawnedAt) / 1000).toFixed(1)}s`;
          const icon = { spawning: "⏳", running: "▶", done: "✓", error: "✗", aborted: "⊘" }[m.status] ?? "?";
          return `${icon} ${m.name} [${m.role}] ${m.status} ${elapsed}`;
        }),
      ];
      return {
        content: [{ type: "text" as const, text: lines.join("\n") }],
        details: { team: teamCtx.name, lead: teamCtx.lead, teammates: members },
      };
    },
  });

  return [AgentStatus, AgentTeam, MailboxReceive, MailboxSend, MailboxBroadcast, ...makeTaskTools(teamCtx)];
}

// ─── Core spawn logic ─────────────────────────────────────────────────────────

interface SpawnResult {
  name: string;
  output: string;
  status: "done" | "error" | "aborted";
  error?: string;
}

async function spawnAgent(
  agentDef: AgentDef,
  instanceName: string,
  task: string,
  cwd: string,
  teamName: string,
  signal?: AbortSignal,
  onUpdate?: (text: string) => void,
  parentModel?: import("@mariozechner/pi-ai").Model<any>,
): Promise<SpawnResult> {
  const teamCtx = teams.get(teamName) ?? teams.root();
  const { registry } = teamCtx;
  // Register (or update existing placeholder from parallel pre-reservation)
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " spawnAgent start " + instanceName + "\n");
  if (registry.isTaken(instanceName)) {
    registry.update(instanceName, { role: agentDef.name, task, model: agentDef.model, status: "spawning" });
  } else {
    registry.register(instanceName, agentDef.name, task, teamName, agentDef.model);
  }
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " spawnAgent registered\n");

  // Build system prompt with injected identity so the agent knows its name
  // Also load AGENTS.md / CLAUDE.md from agentDir and cwd for user guidance (e.g. conciseness rules)
  const agentsMdCandidates = [
    path.join(getAgentDir(), "AGENTS.md"),
    path.join(getAgentDir(), "CLAUDE.md"),
    path.join(cwd, "AGENTS.md"),
    path.join(cwd, "CLAUDE.md"),
  ];
  const agentsMdSections: string[] = [];
  for (const filePath of agentsMdCandidates) {
    try {
      const fileContent = fs.readFileSync(filePath, "utf-8").trim();
      if (fileContent) {
        const label = path.basename(filePath);
        agentsMdSections.push(`\n\n--- ${label} ---\n${fileContent}`);
      }
    } catch { /* file not found — skip */ }
  }

  const identityBlock = [
    `\n\n---`,
    `Your agent name is: **${instanceName}**`,
    `Your team: **${teamName}** | Team lead: **${teamCtx.lead}**`,
    `You are running inside a pi agent session.`,
    ``,
    `Communication tools:`,
    `- mailbox_receive: check for incoming messages (including status requests from the team lead)`,
    `- mailbox_send to="orchestrator": reply to the team lead — use this whenever you receive a steer/status request`,
    `- mailbox_broadcast: reach all active agents`,
    `- agent_status: see who else is running`,
    ``,
    `When you receive a steering message (an interruption from the team lead), respond immediately`,
    `with a brief status update via mailbox_send to="orchestrator" before continuing your work.`,
    ``,
    `Task coordination: TaskCreate / TaskList / TaskGet / TaskUpdate`,
  ].join("\n");

  const systemPrompt = agentDef.systemPrompt + identityBlock + agentsMdSections.join("");

  // Resolve model
  let model: Awaited<ReturnType<typeof resolveModel>> = undefined;
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " spawnAgent entered " + instanceName + "\n");
  registry.appendLog(instanceName, "[bob-agents] spawnAgent: resolving model...");
  if (agentDef.model) {
    model = resolveModel(agentDef.model);
    registry.appendLog(instanceName, "[bob-agents] spawnAgent: model resolved");
  }

  registry.appendLog(instanceName, "[bob-agents] spawnAgent: building resource loader...");
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " building loader model=" + (model ? (model as any).id ?? "set" : "none") + "\n");

  // Inject team-scoped tools via extensionFactories so they register via pi.registerTool()
  // and are guaranteed active. This avoids the customTools activation bug and bob-agents
  // conflict when the parent extension also loads in child sessions.
  const agentDir = getAgentDir();
  const capturedTeamCtx = teamCtx;
  const capturedInstanceName = instanceName;
  const agentExtensionFactory = (childPi: import("@mariozechner/pi-coding-agent").ExtensionAPI) => {
    const boundTools = makeBoundTools(capturedInstanceName, capturedTeamCtx);
    for (const tool of boundTools) {
      childPi.registerTool(tool);
    }
  };

  // If OTel is configured, inject the otel extension into the child session so
  // it emits traces into the parent's trace (reads __bobOtelCtx for child mode).
  const otelEnabled = !!(process.env.OTEL_ENDPOINT && process.env.OTEL_USER && process.env.OTEL_TOKEN);
  const childExtensionFactories = otelEnabled ? [otelExtension] : [];

  const loader = new DefaultResourceLoader({
    cwd,
    agentDir,
    noExtensions: true, // block path-based extensions (bob-agents, pi-lens etc.)
    noSkills: true,
    noPromptTemplates: true,
    noThemes: true,
    noContextFiles: true,
    systemPromptOverride: () => systemPrompt,
    appendSystemPromptOverride: () => [],
    extensionFactories: [agentExtensionFactory, ...childExtensionFactories],
  });
  await loader.reload();
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " loader ready, calling createAgentSession\n");

  const { session } = await createAgentSession({
    cwd,
    agentDir,
    sessionManager: SessionManager.inMemory(),
    settingsManager: SettingsManager.create(cwd, agentDir),
    resourceLoader: loader,
    ...(model ? { model } : {}),
    modelRegistry,
  });

  const { extensionsResult: extRes } = await Promise.resolve({ extensionsResult: loader.getExtensions() });
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " ext errors: " + JSON.stringify(extRes.errors) + "\n");
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " ext count: " + extRes.extensions.length + " tools: " + extRes.extensions.flatMap((e: any) => [...e.tools.keys()]).join(",") + "\n");
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " session created, active: " + session.getActiveToolNames().join(",") + "\n");

  // bindExtensions fires session_start so extension factories initialize properly
  await session.bindExtensions({});
  fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " after bindExtensions, active: " + session.getActiveToolNames().join(",") + "\n");

  registry.appendLog(instanceName, "[bob-agents] spawnAgent: session ready, prompting...");

  liveSessions.set(instanceName, session);
  registry.update(instanceName, { status: "running" });

  // Propagate abort
  const abortHandler = () => {
    session.abort().catch(() => {});
  };
  if (signal?.aborted) {
    session.abort().catch(() => {});
  } else {
    signal?.addEventListener("abort", abortHandler, { once: true });
  }

  // Stream text deltas for onUpdate
  const unsub = session.subscribe((event: { type: string; assistantMessageEvent?: { type: string; delta?: string } }) => {
    if (event.type === "message_update" && event.assistantMessageEvent?.type === "text_delta") {
      const delta = event.assistantMessageEvent.delta ?? "";
      registry.appendStdout(instanceName, delta);
      onUpdate?.(delta);
    }
  });

  try {
    await session.prompt(task);
    const output = getFinalOutput(session);
    registry.finish(instanceName, output);
    return { name: instanceName, output, status: "done" };
  } catch (err: unknown) {
    if (signal?.aborted) {
      registry.update(instanceName, { status: "aborted", finishedAt: Date.now() });
      return { name: instanceName, output: "", status: "aborted", error: "Aborted by user" };
    }
    const error = err instanceof Error ? err.message : String(err);
    registry.fail(instanceName, error);
    return { name: instanceName, output: "", status: "error", error };
  } finally {
    unsub();
    signal?.removeEventListener("abort", abortHandler);
    liveSessions.delete(instanceName);
    session.dispose();
  }
}

// ─── Extension factory ────────────────────────────────────────────────────────

export default function (pi: ExtensionAPI) {
  // ── subagent ──────────────────────────────────────────────────────────────

  pi.registerTool({
    name: "subagent",
    label: "Subagent",
    description: [
      "Delegate tasks to Bob agents loaded from agents/*/SKILL.md.",
      "Three modes:",
      "  single:   { agent, task }",
      "  parallel: { tasks: [{agent, task}, …] }  — all run concurrently",
      "  chain:    { chain: [{agent, task}, …] }  — sequential, {previous} passes prior output",
      "Agents share an in-memory message bus (mailbox_receive/send/broadcast) and task board (TaskCreate/List/Get/Update).",
    ].join(" "),
    parameters: Type.Object({
      agent: Type.Optional(Type.String({ description: "Agent name for single mode" })),
      task: Type.Optional(Type.String({ description: "Task description for single mode" })),
      background: Type.Optional(Type.Boolean({ description: "Single mode only: spawn agent and return immediately without waiting for completion. Use agent_status or agent_wait to track progress." })),
      team: Type.Optional(Type.String({ description: "Team name to spawn agent into. Agents share the team's isolated bus and task board. Defaults to root team." })),
      tasks: Type.Optional(
        Type.Array(
          Type.Object({
            agent: Type.String({ description: "Agent name" }),
            task: Type.String({ description: "Task description" }),
          }),
          { description: "Parallel tasks" },
        ),
      ),
      chain: Type.Optional(
        Type.Array(
          Type.Object({
            agent: Type.String({ description: "Agent name" }),
            task: Type.String({ description: "Task, may include {previous} placeholder" }),
          }),
          { description: "Sequential chain" },
        ),
      ),
    }),

    async execute(_toolCallId, params, signal, onUpdate, ctx) {
      fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " execute called cwd=" + ctx.cwd + " agent=" + (params.agent ?? "?") + "\n");
      fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " calling discoverAgents\n");
      const { agents } = discoverAgents(ctx.cwd);
      fs.appendFileSync("/tmp/bob-agents-debug.log", new Date().toISOString() + " discoverAgents done agents=" + agents.length + "\n");

      const findAgent = (name: string): AgentDef | undefined =>
        agents.find((a) => a.name === name);

      const modeCount =
        Number(Boolean(params.agent && params.task)) +
        Number((params.tasks?.length ?? 0) > 0) +
        Number((params.chain?.length ?? 0) > 0);

      if (modeCount !== 1) {
        const names = agents.map((a) => `${a.name} (${a.source})`).join(", ") || "none";
        return {
          content: [{ type: "text" as const, text: `Provide exactly one mode (single/parallel/chain). Available agents: ${names}` }],
          details: {},
          isError: true,
        };
      }

      // ── Single ─────────────────────────────────────────────────────────────
      if (params.agent && params.task) {
        const def = findAgent(params.agent);
        if (!def) {
          const names = agents.map((a) => a.name).join(", ");
          return {
            content: [{ type: "text" as const, text: `Unknown agent "${params.agent}". Available: ${names}` }],
            details: {},
            isError: true,
          };
        }
        const teamCtxPar = teams.get(params.team ?? ROOT_TEAM) ?? teams.root();
        const name = teamCtxPar.registry.generateName(params.agent);

        // Background mode: fire-and-forget, return immediately
        if (params.background) {
          const p = spawnAgent(def, name, params.task, ctx.cwd, params.team ?? ROOT_TEAM, signal, undefined, ctx.model ?? undefined);
          p.catch(() => {}); // errors are visible via agent_status
          return {
            content: [{ type: "text" as const, text: `Agent ${name} spawned in background. Use agent_status to monitor or agent_wait to block until done.` }],
            details: { name, status: "spawning" },
          };
        }

        // Foreground mode (default): block until complete
        let accumulated = "";
        const result = await spawnAgent(def, name, params.task, ctx.cwd, params.team ?? ROOT_TEAM, signal, (delta) => {
          accumulated += delta;
          onUpdate?.({
            content: [{ type: "text" as const, text: `[${name}] ${accumulated.slice(-300)}` }],
            details: { name, status: "running" },
          });
        }, ctx.model ?? undefined);
        return {
          content: [{ type: "text" as const, text: result.output || result.error || "(no output)" }],
          details: { name, status: result.status },
          isError: result.status === "error",
        };
      }

      // ── Parallel ───────────────────────────────────────────────────────────
      if (params.tasks && params.tasks.length > 0) {
        const MAX_PARALLEL = 8;
        if (params.tasks.length > MAX_PARALLEL) {
          return {
            content: [{ type: "text" as const, text: `Too many parallel tasks (max ${MAX_PARALLEL})` }],
            details: {},
            isError: true,
          };
        }

        // Pre-validate all agents before spawning anything
        for (const t of params.tasks) {
          if (!findAgent(t.agent)) {
            const names = agents.map((a) => a.name).join(", ");
            return {
              content: [{ type: "text" as const, text: `Unknown agent "${t.agent}". Available: ${names}` }],
              details: {},
              isError: true,
            };
          }
        }

        // Assign unique instance names up front so task board ownership is stable.
        // We register placeholders immediately so generateName won't reuse the same
        // name for two concurrent agents of the same role.
        const instanceNames: string[] = [];
        for (const t of params.tasks) {
          const name = teamCtxPar.registry.generateName(t.agent);
          instanceNames.push(name);
          // Register as spawning so the name is reserved; spawnAgent will re-register properly.
          teamCtxPar.registry.register(name, t.agent, t.task, params.team ?? ROOT_TEAM);
        }

        // Emit initial parallel status
        const statusText = () => {
          const active = teamCtxPar.registry.getActive();
          const done = teamCtxPar.registry.getAll().filter((a) => instanceNames.includes(a.name) && (a.status === "done" || a.status === "error" || a.status === "aborted"));
          return `Parallel: ${done.length}/${instanceNames.length} done — ${active.map((a) => a.name).join(", ") || "all finished"}`;
        };

        const results = await Promise.all(
          params.tasks.map((t, i) => {
            const def = findAgent(t.agent)!;
            const name = instanceNames[i];
            // Undo the placeholder so spawnAgent can register cleanly
            teamCtxPar.registry.update(name, { status: "spawning" });
            let acc = "";
            return spawnAgent(def, name, t.task, ctx.cwd, params.team ?? ROOT_TEAM, signal, (delta) => {
              acc += delta;
              onUpdate?.({
                content: [{ type: "text" as const, text: statusText() }],
                details: { mode: "parallel", agents: teamCtxPar.registry.getAll().filter((a) => instanceNames.includes(a.name)) },
              });
            }, ctx.model ?? undefined);
          }),
        );

        const succeeded = results.filter((r) => r.status === "done").length;
        const summaries = results
          .map((r) => `[${r.name}] ${r.status === "done" ? "✓" : "✗"} ${(r.output || r.error || "(no output)").slice(0, 120)}`)
          .join("\n\n");

        return {
          content: [{ type: "text" as const, text: `Parallel: ${succeeded}/${results.length} succeeded\n\n${summaries}` }],
          details: { mode: "parallel", results },
        };
      }

      // ── Chain ──────────────────────────────────────────────────────────────
      if (params.chain && params.chain.length > 0) {
        const results: SpawnResult[] = [];
        let previous = "";

        for (let i = 0; i < params.chain.length; i++) {
          const step = params.chain[i];
          const def = findAgent(step.agent);
          if (!def) {
            const names = agents.map((a) => a.name).join(", ");
            return {
              content: [{ type: "text" as const, text: `Unknown agent "${step.agent}" at step ${i + 1}. Available: ${names}` }],
              details: {},
              isError: true,
            };
          }
          const chainTeamCtx = teams.get(params.team ?? ROOT_TEAM) ?? teams.root();
          const name = chainTeamCtx.registry.generateName(step.agent);
          const task = step.task.replace(/\{previous\}/g, previous);
          let acc = "";
          const result = await spawnAgent(def, name, task, ctx.cwd, params.team ?? ROOT_TEAM, signal, (delta) => {
            acc += delta;
            onUpdate?.({
              content: [{ type: "text" as const, text: `[chain step ${i + 1}/${params.chain!.length}: ${name}] ${acc.slice(-200)}` }],
              details: { mode: "chain", step: i + 1, total: params.chain!.length },
            });
          }, ctx.model ?? undefined);
          results.push(result);
          if (result.status !== "done") {
            return {
              content: [{ type: "text" as const, text: `Chain stopped at step ${i + 1} (${name}): ${result.error ?? "unknown error"}` }],
              details: { mode: "chain", results },
              isError: true,
            };
          }
          previous = result.output;
        }

        const finalResult = results[results.length - 1];
        return {
          content: [{ type: "text" as const, text: finalResult?.output || "(no output)" }],
          details: { mode: "chain", results },
        };
      }

      return {
        content: [{ type: "text" as const, text: "No mode matched." }],
        details: {},
        isError: true,
      };
    },
  });


  // ── agent_steer ───────────────────────────────────────────────────────────

  pi.registerTool({
    name: "agent_steer",
    label: "Agent Steer",
    description: "Send a steering message to a running agent. Delivered after its current tool call finishes. The agent is instructed to reply via mailbox_send to the team lead so you can read the response without waiting for full completion.",
    parameters: Type.Object({
      agent: Type.String({ description: "Agent name to steer" }),
      message: Type.String({ description: "Message to deliver — e.g. 'brief status update', 'what are you currently working on?'" }),
      team: Type.Optional(Type.String({ description: "Team name to read reply from. Defaults to root team." })),
    }),
    async execute(_id, params) {
      const liveSession = liveSessions.get(params.agent);
      if (!liveSession) {
        const rec = teams.getAll().flatMap(t => t.registry.getAll()).find(a => a.name === params.agent);
        if (!rec) return { content: [{ type: "text" as const, text: `Agent ${params.agent} not found.` }], details: {}, isError: true };
        if (rec.status === "done" || rec.status === "error" || rec.status === "aborted") {
          return { content: [{ type: "text" as const, text: `Agent ${params.agent} is ${rec.status} — cannot steer.` }], details: {}, isError: true };
        }
        return { content: [{ type: "text" as const, text: `Agent ${params.agent} has no live session.` }], details: {}, isError: true };
      }

      // Find the team lead for the reply instruction
      const teamCtx = params.team ? (teams.get(params.team) ?? teams.root()) : teams.teamOf(params.agent) ?? teams.root();
      const lead = teamCtx.lead;

      // Inject the steer with explicit reply instruction
      const steerText = `[Steering message from ${lead}]: ${params.message}\n\nReply now via mailbox_send(to="${lead}", content="<your status>") before continuing your work.`;
      await liveSession.steer(steerText);

      return {
        content: [{ type: "text" as const, text: `Steered ${params.agent}. Reply will arrive in mailbox "${lead}" — use mailbox_read(team: "${teamCtx.name}") to check.` }],
        details: { agent: params.agent, team: teamCtx.name, lead },
      };
    },
  });

  // ── agent_wait ────────────────────────────────────────────────────────────

  pi.registerTool({
    name: "agent_wait",
    label: "Agent Wait",
    description: "Block until one or more named background agents finish (done/error/aborted). Returns their final status and output.",
    parameters: Type.Object({
      agents: Type.Array(Type.String(), { description: "Agent names to wait for" }),
      timeout_ms: Type.Optional(Type.Number({ description: "Max wait in milliseconds (default: 300000)" })),
    }),
    async execute(_id, params, signal) {
      const timeout = params.timeout_ms ?? 300_000;
      const deadline = Date.now() + timeout;
      const TERMINAL = new Set(["done", "error", "aborted"]);

      while (Date.now() < deadline) {
        if (signal?.aborted) break;
        const all = params.agents.map((n) => teams.getAll().flatMap(t => t.registry.getAll()).find(a => a.name === n));
        if (all.every((a) => a && TERMINAL.has(a.status))) break;
        await new Promise((r) => setTimeout(r, 500));
      }

      const lines = params.agents.map((n) => {
        const a = teams.getAll().flatMap(t => t.registry.getAll()).find(ag => ag.name === n);
        if (!a) return `${n}: not found`;
        const excerpt = (a.output || a.error || "").slice(0, 300);
        return `${a.status.padEnd(8)} ${n}${excerpt ? `: ${excerpt}` : ""}`;
      });

      return {
        content: [{ type: "text" as const, text: lines.join("\n") }],
        details: { agents: params.agents.map((n) => teams.getAll().flatMap(t => t.registry.getAll()).find(a => a.name === n)) },
      };
    },
  });

  // ── agent_status ───────────────────────────────────────────────────────────

  pi.registerTool({
    name: "agent_status",
    label: "Agent Status",
    description: "List all spawned agents and their current status",
    parameters: Type.Object({}),
    async execute() {
      const allAgents = teams.getAll().flatMap((t) => t.registry.getAll());
      return {
        content: [{ type: "text" as const, text: formatAgentStatus(allAgents) }],
        details: { agents: allAgents },
      };
    },
  });

  // ── agent_output ──────────────────────────────────────────────────────────

  pi.registerTool({
    name: "agent_output",
    label: "Agent Output",
    description: "Get the live stdout buffer and lifecycle log for a specific agent (last 8KB of streaming output + last 50 log entries).",
    parameters: Type.Object({
      agent: Type.String({ description: "Agent name" }),
      tail: Type.Optional(Type.Number({ description: "Number of stdout lines to tail (default: all buffered)" })),
    }),
    async execute(_id, params) {
      const rec = teams.getAll().flatMap(t => t.registry.getAll()).find(a => a.name === params.agent);
      if (!rec) {
        return { content: [{ type: "text" as const, text: `Agent ${params.agent} not found.` }], details: {}, isError: true };
      }
      const elapsed = rec.finishedAt
        ? ((rec.finishedAt - rec.spawnedAt) / 1000).toFixed(1)
        : ((Date.now() - rec.spawnedAt) / 1000).toFixed(1);
      const status = `Agent: ${rec.name} [${rec.role}] ${rec.status} ${elapsed}s${rec.team ? ` team=${rec.team}` : ""}`;

      // Lifecycle log
      const logSection = rec.log.length > 0
        ? `\n--- Lifecycle Log ---\n${rec.log.join("\n")}`
        : "\n--- Lifecycle Log ---\n(empty)";

      // Stdout
      const lines = rec.stdout.split("\n");
      const tail = params.tail ?? lines.length;
      const output = lines.slice(-tail).join("\n");
      const overflow = rec.stdoutBytes > rec.stdout.length
        ? `[... +${rec.stdoutBytes - rec.stdout.length} bytes truncated ...]\n`
        : "";
      const stdoutSection = `\n--- Stdout ---\n${overflow}${output || "(no output yet)"}`;

      return {
        content: [{ type: "text" as const, text: `${status}${logSection}${stdoutSection}` }],
        details: { name: rec.name, status: rec.status, log: rec.log, stdoutBytes: rec.stdoutBytes, stdout: rec.stdout },
      };
    },
  });

  // ── Orchestrator mailbox tools ─────────────────────────────────────────────

  pi.registerTool({
    name: "mailbox_read",
    label: "Mailbox Read",
    description: 'Read messages sent to the orchestrator mailbox (or any named mailbox). Pass agent name to read an agent\'s mailbox.',
    parameters: Type.Object({
      agent: Type.Optional(Type.String({ description: 'Read this agent\'s mailbox instead of "orchestrator"' })),
      all: Type.Optional(Type.Boolean({ description: "Include already-read messages (default: false)" })),
      team: Type.Optional(Type.String({ description: "Read from this team\'s bus instead of root. Defaults to root team." })),
    }),
    async execute(_id, params) {
      const target = params.agent ?? "orchestrator";
      const teamCtx = params.team ? (teams.get(params.team) ?? teams.root()) : teams.root();
      const msgs = params.all ? teamCtx.bus.all(target) : teamCtx.bus.receive(target);
      if (msgs.length === 0) {
        return {
          content: [{ type: "text" as const, text: `No ${params.all ? "" : "unread "}messages in mailbox "${target}".` }],
          details: { messages: [] },
        };
      }
      const formatted = msgs
        .map(
          (m) =>
            `From: ${m.from} [${new Date(m.timestamp).toISOString()}]${m.read && params.all ? " (read)" : ""}\n${m.content}`,
        )
        .join("\n---\n");
      return {
        content: [{ type: "text" as const, text: formatted }],
        details: { messages: msgs },
      };
    },
  });

  pi.registerTool({
    name: "mailbox_send_as",
    label: "Mailbox Send",
    description: 'Send a message from "orchestrator" to a specific agent',
    parameters: Type.Object({
      to: Type.String({ description: "Recipient agent name" }),
      content: Type.String({ description: "Message content" }),
      team: Type.Optional(Type.String({ description: "Send on this team\'s bus. Defaults to root team." })),
    }),
    async execute(_id, params) {
      const teamCtx = params.team ? (teams.get(params.team) ?? teams.root()) : teams.root();
      const msg = teamCtx.bus.send("orchestrator", params.to, params.content);
      return {
        content: [{ type: "text" as const, text: `Message sent to ${params.to} (id: ${msg.id})` }],
        details: { messageId: msg.id },
      };
    },
  });

  pi.registerTool({
    name: "mailbox_broadcast",
    label: "Mailbox Broadcast",
    description: "Broadcast a message from the orchestrator to all active agents",
    parameters: Type.Object({
      content: Type.String({ description: "Message to broadcast" }),
    }),
    async execute(_id, params) {
      const recipients = teams.root().registry.activeNames();
      teams.root().bus.broadcast("orchestrator", params.content, recipients);
      return {
        content: [{ type: "text" as const, text: `Broadcast sent to ${recipients.length} agents: ${recipients.join(", ")}` }],
        details: { recipients },
      };
    },
  });

  // ── Task board tools (orchestrator side) ───────────────────────────────────

  for (const tool of makeTaskTools(teams.root())) {
    pi.registerTool(tool);
  }


  // ── Team management tools ──────────────────────────────────────────────────

  pi.registerTool({
    name: "TeamCreate",
    label: "Team Create",
    description: "Create a named team with its own isolated message bus, task board, and agent registry. The caller becomes the team lead (use your agent name, or \"orchestrator\" for the main session).",
    parameters: Type.Object({
      name: Type.String({ description: "Unique team name" }),
      lead: Type.Optional(Type.String({ description: 'Team lead agent name. Default: "orchestrator"' })),
    }),
    async execute(_id, params) {
      try {
        const team = teams.create(params.name, params.lead ?? "orchestrator");
        return {
          content: [{ type: "text" as const, text: `Team "${team.name}" created. Lead: ${team.lead}. Spawn agents with subagent(team: "${team.name}", ...)` }],
          details: { team: { name: team.name, lead: team.lead } },
        };
      } catch (err) {
        return { content: [{ type: "text" as const, text: String(err) }], details: {}, isError: true };
      }
    },
  });

  pi.registerTool({
    name: "TeamStatus",
    label: "Team Status",
    description: "Show status of a team: lead, members, task board summary, and recent messages.",
    parameters: Type.Object({
      name: Type.String({ description: "Team name" }),
    }),
    async execute(_id, params) {
      const team = teams.get(params.name);
      if (!team) return { content: [{ type: "text" as const, text: `Team "${params.name}" not found.` }], details: {}, isError: true };
      const members = team.registry.getAll();
      const tasks = team.taskBoard.list();
      const unread = team.bus.receive(team.lead);
      const lines = [
        `Team: ${team.name} | Lead: ${team.lead} | Members: ${members.length}`,
        ...members.map((m) => {
          const elapsed = m.finishedAt ? `${((m.finishedAt - m.spawnedAt) / 1000).toFixed(1)}s` : `${((Date.now() - m.spawnedAt) / 1000).toFixed(1)}s`;
          const icon = { spawning: "⏳", running: "▶", done: "✓", error: "✗", aborted: "⊘" }[m.status] ?? "?";
          return `  ${icon} ${m.name} [${m.role}] ${m.status} ${elapsed}`;
        }),
        `Tasks: ${tasks.length} total, ${tasks.filter(t => t.status === "todo").length} todo, ${tasks.filter(t => t.status === "in_progress").length} in_progress, ${tasks.filter(t => t.status === "done").length} done`,
        unread.length > 0 ? `📬 ${unread.length} unread message(s) for team lead` : "",
      ].filter(Boolean);
      return {
        content: [{ type: "text" as const, text: lines.join("\n") }],
        details: { name: team.name, lead: team.lead, members, tasks },
      };
    },
  });

  pi.registerTool({
    name: "TeamDisband",
    label: "Team Disband",
    description: "Disband a team. Does not abort running agents — call agent_wait first if needed.",
    parameters: Type.Object({
      name: Type.String({ description: "Team name to disband" }),
    }),
    async execute(_id, params) {
      try {
        const team = teams.disband(params.name);
        if (!team) return { content: [{ type: "text" as const, text: `Team "${params.name}" not found.` }], details: {}, isError: true };
        return { content: [{ type: "text" as const, text: `Team "${params.name}" disbanded.` }], details: {} };
      } catch (err) {
        return { content: [{ type: "text" as const, text: String(err) }], details: {}, isError: true };
      }
    },
  });

  // ── /agents command ────────────────────────────────────────────────────────

  pi.registerCommand("agents", {
    description: "Show all spawned agents, their status, and recent mailbox activity",
    handler: async (_args, ctx) => {
      const all = teams.getAll().flatMap(t => t.registry.getAll());
      if (all.length === 0) {
        ctx.ui.notify("No agents spawned yet.", "info");
        return;
      }
      ctx.ui.notify(formatAgentStatus(all), "info");

      // Show unread orchestrator messages
      const unread = teams.root().bus.receive("orchestrator");
      if (unread.length > 0) {
        ctx.ui.notify(`📬 ${unread.length} new message(s) in orchestrator mailbox`, "info");
      }
    },
  });

  // ── Team UI commands (/team-status, /team-tasks) ────────────────────────

  registerTeamCommands(pi, teams);

  // ── Context file loading (AGENTS.md / CLAUDE.md) ─────────────────────────

  pi.on("before_agent_start", async (event, ctx) => {
    // Prefer the cwd baked into systemPromptOptions (most accurate for the session)
    type SysPromptOpts = { cwd?: string; contextFiles?: { path: string }[] };
    const opts = event.systemPromptOptions as SysPromptOpts | undefined;
    const cwd = opts?.cwd ?? ctx.cwd;

    // Files pi already loaded natively — skip to avoid duplication
    const alreadyLoaded = new Set<string>((opts?.contextFiles ?? []).map((f) => f.path));

    const agentDir = getAgentDir();
    const candidates = [
      // User-global (loaded first — project context appends on top)
      path.join(agentDir, "AGENTS.md"),
      path.join(agentDir, "CLAUDE.md"),
      // Project .pi/
      path.join(cwd, ".pi", "AGENTS.md"),
      path.join(cwd, ".pi", "CLAUDE.md"),
      // Project root
      path.join(cwd, "AGENTS.md"),
      path.join(cwd, "CLAUDE.md"),
    ];

    const sections: string[] = [];
    for (const filePath of candidates) {
      if (alreadyLoaded.has(filePath)) continue;
      try {
        const content = fs.readFileSync(filePath, "utf-8").trim();
        if (content) {
          const label = path.relative(cwd, filePath);
          sections.push(`\n\n--- ${label} ---\n${content}`);
        }
      } catch {
        // File doesn't exist or is unreadable — skip silently
      }
    }

    if (sections.length === 0) return;

    return {
      systemPrompt: event.systemPrompt + sections.join(""),
    };
  });

  // ── Cleanup on session shutdown ────────────────────────────────────────────

  pi.on("session_shutdown", async () => {
    // Only abort live sessions spawned by THIS session's extension instance.
    // Do NOT reset shared singletons (bus, registry, taskBoard) — they are
    // process-lifetime objects shared across all sessions. Resetting them on
    // child session shutdown would wipe the parent orchestrator's state.
    const aborts = Array.from(liveSessions.values()).map((s: any) => s.abort().catch(() => {}));
    await Promise.all(aborts);
    liveSessions.clear();
  });
}
