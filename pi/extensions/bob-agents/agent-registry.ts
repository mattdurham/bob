/**
 * Tracks all spawned agent sessions and provides a shared task board.
 *
 * The task board replaces Claude Code's TaskCreate/TaskList/TaskGet/TaskUpdate
 * so that Bob's existing agent prompts work unchanged in pi.
 */

// ─── Agent registry ──────────────────────────────────────────────────────────

export type AgentStatus = "spawning" | "running" | "done" | "error" | "aborted";

export interface AgentRecord {
  name: string;
  role: string; // agent type from SKILL.md
  task: string;
  status: AgentStatus;
  model?: string;
  spawnedAt: number;
  finishedAt?: number;
  error?: string;
  output?: string; // final assistant text on completion
}

export class AgentRegistry {
  private agents = new Map<string, AgentRecord>();

  register(name: string, role: string, task: string, model?: string): void {
    this.agents.set(name, {
      name,
      role,
      task,
      status: "spawning",
      model,
      spawnedAt: Date.now(),
    });
  }

  update(name: string, patch: Partial<AgentRecord>): void {
    const rec = this.agents.get(name);
    if (rec) Object.assign(rec, patch);
  }

  finish(name: string, output: string): void {
    this.update(name, { status: "done", finishedAt: Date.now(), output });
  }

  fail(name: string, error: string): void {
    this.update(name, { status: "error", finishedAt: Date.now(), error });
  }

  get(name: string): AgentRecord | undefined {
    return this.agents.get(name);
  }

  /** Alias for get() — returns the full AgentRecord for a name, or undefined. */
  getEntry(name: string): AgentRecord | undefined {
    return this.agents.get(name);
  }

  getAll(): AgentRecord[] {
    return Array.from(this.agents.values());
  }

  getActive(): AgentRecord[] {
    return this.getAll().filter((a) => a.status === "spawning" || a.status === "running");
  }

  activeNames(): string[] {
    return this.getActive().map((a) => a.name);
  }

  /**
   * Generate a unique agent instance name from a role, e.g.:
   *   team-coder       (first)
   *   team-coder-2     (second)
   */
  generateName(role: string): string {
    const base = role.toLowerCase().replace(/[^a-z0-9-]/g, "-");
    if (!this.agents.has(base)) return base;
    let i = 2;
    while (this.agents.has(`${base}-${i}`)) i++;
    return `${base}-${i}`;
  }

  isTaken(name: string): boolean {
    return this.agents.has(name);
  }

  reset(): void {
    this.agents.clear();
  }
}

// ─── Task board (replaces Claude Code TaskCreate/TaskList/TaskGet/TaskUpdate) ─

export type TaskStatus = "todo" | "in_progress" | "done" | "blocked" | "failed";

export interface BoardTask {
  id: string;
  title: string;
  description: string;
  status: TaskStatus;
  owner?: string;
  tags?: string[];
  dependencies?: string[]; // task ids that must be done first
  notes?: string;
  createdAt: number;
  updatedAt: number;
}

export class TaskBoard {
  private tasks = new Map<string, BoardTask>();
  private nextId = 1;

  create(
    title: string,
    description: string,
    options?: {
      tags?: string[];
      dependencies?: string[];
      owner?: string;
    },
  ): BoardTask {
    const id = `task-${String(this.nextId++).padStart(3, "0")}`;
    const task: BoardTask = {
      id,
      title,
      description,
      status: "todo",
      owner: options?.owner,
      tags: options?.tags,
      dependencies: options?.dependencies,
      createdAt: Date.now(),
      updatedAt: Date.now(),
    };
    this.tasks.set(id, task);
    return task;
  }

  get(id: string): BoardTask | undefined {
    return this.tasks.get(id);
  }

  list(filter?: { status?: TaskStatus; owner?: string; tag?: string }): BoardTask[] {
    let result = Array.from(this.tasks.values());
    if (filter?.status) result = result.filter((t) => t.status === filter.status);
    if (filter?.owner) result = result.filter((t) => t.owner === filter.owner);
    if (filter?.tag) result = result.filter((t) => t.tags?.includes(filter.tag!));
    return result;
  }

  update(
    id: string,
    patch: {
      status?: TaskStatus;
      owner?: string;
      notes?: string;
      tags?: string[];
    },
  ): BoardTask | undefined {
    const task = this.tasks.get(id);
    if (!task) return undefined;
    Object.assign(task, { ...patch, updatedAt: Date.now() });
    return task;
  }

  reset(): void {
    this.tasks.clear();
    this.nextId = 1;
  }
}
