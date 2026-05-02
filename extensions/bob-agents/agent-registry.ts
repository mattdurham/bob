/**
 * Agent registry, team registry, task board, and message bus.
 *
 * Teams are the central organizing concept. Each team has:
 *   - Its own MessageBus   (isolated mailboxes; team lead is "orchestrator")
 *   - Its own TaskBoard    (isolated task list)
 *   - Its own AgentRegistry (tracks only its members)
 *
 * A subagent can create a team and become its team lead, forming nested hierarchies.
 * The root context (main pi session) has a default team named "__root__".
 */

import { MessageBus } from "./message-bus.js";

// ─── Shared constants ─────────────────────────────────────────────────────────

/** Rolling buffer size for live stdout capture (bytes). */
const STDOUT_BUFFER = 8_000;

/** Max lifecycle log entries per agent. */
const LOG_RING = 50;

/** Name of the root team (main pi session). */
export const ROOT_TEAM = "__root__";

// ─── Agent record ─────────────────────────────────────────────────────────────

export type AgentStatus = "spawning" | "running" | "done" | "error" | "aborted";

export interface AgentRecord {
	name: string;
	role: string;
	task: string;
	status: AgentStatus;
	model?: string;
	team: string; // team this agent belongs to
	spawnedAt: number;
	finishedAt?: number;
	error?: string;
	output?: string; // final assistant text on completion
	stdout: string; // rolling buffer of streaming deltas
	stdoutBytes: number; // total bytes received (for overflow indication)
	log: string[]; // ring buffer of lifecycle events
}

// ─── Task board ───────────────────────────────────────────────────────────────

export type TaskStatus = "todo" | "in_progress" | "done" | "blocked" | "failed";

export interface BoardTask {
	id: string;
	title: string;
	description: string;
	status: TaskStatus;
	owner?: string;
	tags?: string[];
	dependencies?: string[];
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
		options?: { tags?: string[]; dependencies?: string[]; owner?: string },
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

	list(filter?: {
		status?: TaskStatus;
		owner?: string;
		tag?: string;
	}): BoardTask[] {
		let result = Array.from(this.tasks.values());
		if (filter?.status)
			result = result.filter((t) => t.status === filter.status);
		if (filter?.owner) result = result.filter((t) => t.owner === filter.owner);
		if (filter?.tag)
			result = result.filter((t) => t.tags?.includes(filter.tag!));
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

// ─── Per-team agent registry ──────────────────────────────────────────────────

export class AgentRegistry {
	private agents = new Map<string, AgentRecord>();

	register(
		name: string,
		role: string,
		task: string,
		team: string,
		model?: string,
	): void {
		this.agents.set(name, {
			name,
			role,
			task,
			status: "spawning",
			model,
			team,
			spawnedAt: Date.now(),
			stdout: "",
			stdoutBytes: 0,
			log: [],
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
	getEntry(name: string): AgentRecord | undefined {
		return this.agents.get(name);
	}
	getAll(): AgentRecord[] {
		return Array.from(this.agents.values());
	}
	getActive(): AgentRecord[] {
		return this.getAll().filter(
			(a) => a.status === "spawning" || a.status === "running",
		);
	}
	activeNames(): string[] {
		return this.getActive().map((a) => a.name);
	}

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

	appendLog(name: string, msg: string): void {
		const rec = this.agents.get(name);
		if (!rec) return;
		rec.log.push(`${new Date().toISOString()} ${msg}`);
		if (rec.log.length > LOG_RING) rec.log.shift();
	}

	appendStdout(name: string, delta: string): void {
		const rec = this.agents.get(name);
		if (!rec) return;
		rec.stdoutBytes += delta.length;
		rec.stdout += delta;
		if (rec.stdout.length > STDOUT_BUFFER)
			rec.stdout = rec.stdout.slice(rec.stdout.length - STDOUT_BUFFER);
	}

	reset(): void {
		this.agents.clear();
	}
}

// ─── Team ─────────────────────────────────────────────────────────────────────

export interface TeamContext {
	/** Team name. */
	name: string;
	/** Agent name (or "orchestrator") that leads this team. */
	lead: string;
	/** Isolated message bus — team lead is "orchestrator" in this bus. */
	bus: MessageBus;
	/** Isolated task board for this team. */
	taskBoard: TaskBoard;
	/** Registry of agents belonging to this team. */
	registry: AgentRegistry;
	createdAt: number;
}

// ─── Global team manager ──────────────────────────────────────────────────────

/**
 * Manages all teams. The root team (__root__) represents the main pi session.
 * Any agent can create a sub-team and become its team lead.
 */
export class TeamManager {
	private teams = new Map<string, TeamContext>();

	constructor() {
		// Create root team automatically — lead is "orchestrator" (main pi session)
		this.create(ROOT_TEAM, "orchestrator");
	}

	create(name: string, lead: string): TeamContext {
		if (this.teams.has(name)) throw new Error(`Team "${name}" already exists`);
		const team: TeamContext = {
			name,
			lead,
			bus: new MessageBus(),
			taskBoard: new TaskBoard(),
			registry: new AgentRegistry(),
			createdAt: Date.now(),
		};
		this.teams.set(name, team);
		return team;
	}

	get(name: string): TeamContext | undefined {
		return this.teams.get(name);
	}
	getAll(): TeamContext[] {
		return Array.from(this.teams.values());
	}

	/** Get the root team context. */
	root(): TeamContext {
		return this.teams.get(ROOT_TEAM)!;
	}

	disband(name: string): TeamContext | undefined {
		if (name === ROOT_TEAM) throw new Error("Cannot disband root team");
		const team = this.teams.get(name);
		this.teams.delete(name);
		return team;
	}

	/** Find which team an agent belongs to. */
	teamOf(agentName: string): TeamContext | undefined {
		for (const team of this.teams.values()) {
			if (team.registry.get(agentName)) return team;
		}
		return undefined;
	}

	reset(): void {
		this.teams.clear();
		this.create(ROOT_TEAM, "orchestrator");
	}
}
