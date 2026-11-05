/**
 * Engram memory extension for pi.
 *
 * On session start: runs `engram context` and injects recent memories into
 * the system prompt so the model has relevant context automatically.
 *
 * Registers tools:
 *   mem_search  — search memories by query
 *   mem_save    — save a new memory (title + content)
 *   mem_context — reload context memories into a message
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { Type } from "typebox";
import { execSync } from "node:child_process";

function runEngram(args: string): string {
	try {
		return execSync(`engram ${args}`, {
			encoding: "utf8",
			timeout: 10_000,
			stdio: ["pipe", "pipe", "pipe"],
		})
			.split("\n")
			.filter(
				(l) =>
					!l.startsWith("Update available") &&
					!l.startsWith("To update:") &&
					!l.includes("brew update"),
			)
			.join("\n")
			.trim();
	} catch (e: unknown) {
		const err = e as { stdout?: string; stderr?: string; message?: string };
		const out = (err.stdout ?? "")
			.split("\n")
			.filter(
				(l) =>
					!l.startsWith("Update available") &&
					!l.startsWith("To update:") &&
					!l.includes("brew update"),
			)
			.join("\n")
			.trim();
		return out || err.message || "engram error";
	}
}

export default function (pi: ExtensionAPI) {
	// ── Auto-inject context on session start ───────────────────────────────────

	pi.on("before_agent_start", async (event) => {
		const context = runEngram("context");
		if (!context || context === "No previous session memories found.") return;

		const injection = `\n\n--- Memory from Previous Sessions (engram) ---\n${context}\n---`;
		return {
			systemPrompt: event.systemPrompt + injection,
		};
	});

	// ── mem_search ─────────────────────────────────────────────────────────────

	pi.registerTool({
		name: "mem_search",
		label: "Memory Search",
		description: "Search engram memory for relevant past context.",
		parameters: Type.Object({
			query: Type.String({ description: "Search query" }),
			limit: Type.Optional(
				Type.Number({ description: "Max results (default: 5)" }),
			),
		}),
		async execute(_id, params) {
			const limitFlag = params.limit
				? ` --limit ${params.limit}`
				: " --limit 5";
			const result = runEngram(
				`search ${JSON.stringify(params.query)}${limitFlag}`,
			);
			return {
				content: [
					{ type: "text" as const, text: result || "No memories found." },
				],
				details: {},
			};
		},
	});

	// ── mem_save ───────────────────────────────────────────────────────────────

	pi.registerTool({
		name: "mem_save",
		label: "Memory Save",
		description:
			"Save a memory to engram. Use for decisions, findings, architecture changes, or anything worth remembering across sessions.",
		parameters: Type.Object({
			title: Type.String({ description: "Short title for the memory" }),
			content: Type.String({ description: "Full memory content" }),
		}),
		async execute(_id, params) {
			const result = runEngram(
				`save ${JSON.stringify(params.title)} ${JSON.stringify(params.content)}`,
			);
			return {
				content: [{ type: "text" as const, text: result }],
				details: {},
			};
		},
	});

	// ── mem_context ────────────────────────────────────────────────────────────

	pi.registerTool({
		name: "mem_context",
		label: "Memory Context",
		description:
			"Load recent memory context from engram for the current project.",
		parameters: Type.Object({
			project: Type.Optional(
				Type.String({ description: "Project name (default: current)" }),
			),
		}),
		async execute(_id, params) {
			const projectFlag = params.project
				? ` ${JSON.stringify(params.project)}`
				: "";
			const result = runEngram(`context${projectFlag}`);
			return {
				content: [
					{ type: "text" as const, text: result || "No memories found." },
				],
				details: {},
			};
		},
	});

	// ── mem_stats ──────────────────────────────────────────────────────────────

	pi.registerTool({
		name: "mem_stats",
		label: "Memory Stats",
		description: "Show engram memory statistics.",
		parameters: Type.Object({}),
		async execute() {
			const result = runEngram("stats");
			return {
				content: [{ type: "text" as const, text: result }],
				details: {},
			};
		},
	});
}
