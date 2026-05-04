/**
 * Team UI commands for bob-agents.
 *
 * /team-status  — overlay showing all teams, their members and agent status
 * /team-tasks   — overlay showing all tasks across all teams in a tree view
 *
 * Escape closes either overlay.
 */

import type { ExtensionAPI, Theme } from "@mariozechner/pi-coding-agent";
import { matchesKey, Text } from "@mariozechner/pi-tui";
import type { TeamManager } from "./agent-registry.js";

// ─── Team Status overlay ──────────────────────────────────────────────────────

class TeamStatusOverlay {
	readonly width = 80;

	constructor(
		private teams: TeamManager,
		private theme: Theme,
		private done: (v: void) => void,
	) {}

	handleInput(data: string): void {
		if (
			matchesKey(data, "escape") ||
			matchesKey(data, "return") ||
			matchesKey(data, "q")
		) {
			this.done();
		}
	}

	render(width: number): string[] {
		const t = this.theme;
		const lines: string[] = [];
		const w = Math.min(width, this.width);

		lines.push(
			t.fg("toolTitle", t.bold(" Team Status")) +
				t.fg("dim", "  (Esc to close)"),
		);
		lines.push(t.fg("dim", "─".repeat(w - 2)));

		for (const team of this.teams.getAll()) {
			const memberCount = team.registry.getAll().length;
			lines.push(
				t.fg("accent", ` ▸ ${team.name}`) +
					t.fg("dim", `  lead: ${team.lead}  members: ${memberCount}`),
			);

			const members = team.registry.getAll();
			if (members.length === 0) {
				lines.push(t.fg("dim", "     (no members)"));
			}
			for (const m of members) {
				const elapsed = m.finishedAt
					? `${((m.finishedAt - m.spawnedAt) / 1000).toFixed(1)}s`
					: `${((Date.now() - m.spawnedAt) / 1000).toFixed(1)}s`;
				const icon =
					{ spawning: "⏳", running: "▶", done: "✓", error: "✗", aborted: "⊘" }[
						m.status
					] ?? "?";
				const statusColor =
					m.status === "done"
						? "success"
						: m.status === "error"
							? "error"
							: m.status === "running"
								? "accent"
								: "dim";
				const tail =
					m.status === "running" && m.stdout
						? t.fg(
								"dim",
								"  " +
									m.stdout
										.split("\n")
										.filter((l) => l.trim())
										.slice(-1)[0]
										?.slice(0, 40) ?? "",
							)
						: "";
				lines.push(
					`   ${icon} ` +
						t.fg(statusColor, m.name) +
						t.fg("dim", ` [${m.role}] ${elapsed}`) +
						tail,
				);
			}

			// unread messages for lead
			const unread = team.bus.all(team.lead).filter((m) => !m.read);
			if (unread.length > 0) {
				lines.push(
					t.fg(
						"warning",
						`   📬 ${unread.length} unread message(s) for ${team.lead}`,
					),
				);
			}

			lines.push("");
		}

		if (this.teams.getAll().length === 0) {
			lines.push(t.fg("dim", " No teams."));
		}

		return lines.map((l) => l.slice(0, w));
	}
}

// ─── Team Tasks overlay ───────────────────────────────────────────────────────

class TeamTasksOverlay {
	readonly width = 90;

	constructor(
		private teams: TeamManager,
		private theme: Theme,
		private done: (v: void) => void,
	) {}

	handleInput(data: string): void {
		if (
			matchesKey(data, "escape") ||
			matchesKey(data, "return") ||
			matchesKey(data, "q")
		) {
			this.done();
		}
	}

	render(width: number): string[] {
		const t = this.theme;
		const lines: string[] = [];
		const w = Math.min(width, this.width);

		lines.push(
			t.fg("toolTitle", t.bold(" Team Tasks")) +
				t.fg("dim", "  (Esc to close)"),
		);
		lines.push(t.fg("dim", "─".repeat(w - 2)));

		const statusIcon: Record<string, string> = {
			todo: "○",
			in_progress: "▶",
			done: "✓",
			blocked: "⊘",
			failed: "✗",
		};
		const statusColor: Record<string, string> = {
			todo: "dim",
			in_progress: "accent",
			done: "success",
			blocked: "warning",
			failed: "error",
		};

		for (const team of this.teams.getAll()) {
			const tasks = team.taskBoard.list();
			const active = tasks.filter(
				(t) => t.status !== "done" && t.status !== "failed",
			);
			lines.push(
				t.fg("accent", ` ▸ ${team.name}`) +
					t.fg("dim", `  ${tasks.length} tasks (${active.length} active)`),
			);

			if (tasks.length === 0) {
				lines.push(t.fg("dim", "     (no tasks)"));
			}

			for (const task of tasks) {
				const icon = statusIcon[task.status] ?? "?";
				const color = statusColor[task.status] ?? "dim";
				const owner = task.owner ? t.fg("dim", ` @${task.owner}`) : "";
				const deps = task.dependencies?.length
					? t.fg("dim", ` [deps: ${task.dependencies.join(",")}]`)
					: "";
				const title = task.title.slice(0, 45);
				lines.push(
					`   ${icon} ` +
						t.fg(color, `${task.id}`) +
						t.fg("dim", ` ${title}`) +
						owner +
						deps,
				);
				if (task.notes) {
					lines.push(t.fg("dim", `        ${task.notes.slice(0, 60)}`));
				}
			}

			lines.push("");
		}

		if (this.teams.getAll().length === 0) {
			lines.push(t.fg("dim", " No teams."));
		}

		return lines.map((l) => l.slice(0, w));
	}
}

// ─── Register commands ────────────────────────────────────────────────────────

export function registerTeamCommands(
	pi: ExtensionAPI,
	teams: TeamManager,
): void {
	pi.registerCommand("team-status", {
		description: "Show all teams, members, and agent status in a popup",
		handler: async (_args, ctx) => {
			if (!ctx.hasUI) {
				const lines: string[] = [];
				for (const team of teams.getAll()) {
					lines.push(`Team: ${team.name} (lead: ${team.lead})`);
					for (const m of team.registry.getAll()) {
						const elapsed = m.finishedAt
							? `${((m.finishedAt - m.spawnedAt) / 1000).toFixed(1)}s`
							: `${((Date.now() - m.spawnedAt) / 1000).toFixed(1)}s`;
						lines.push(
							`  ${m.status.padEnd(9)} ${m.name} [${m.role}] ${elapsed}`,
						);
					}
				}
				ctx.ui.notify(lines.join("\n") || "No teams.", "info");
				return;
			}
			await ctx.ui.custom<void>(
				(_tui, theme, _keys, done) => new TeamStatusOverlay(teams, theme, done),
				{ overlay: true },
			);
		},
	});

	pi.registerCommand("team-tasks", {
		description: "Show all team task boards in a popup tree view",
		handler: async (_args, ctx) => {
			if (!ctx.hasUI) {
				const lines: string[] = [];
				for (const team of teams.getAll()) {
					lines.push(`Team: ${team.name}`);
					for (const task of team.taskBoard.list()) {
						lines.push(
							`  [${task.status}] ${task.id}: ${task.title}${task.owner ? ` @${task.owner}` : ""}`,
						);
					}
				}
				ctx.ui.notify(lines.join("\n") || "No tasks.", "info");
				return;
			}
			await ctx.ui.custom<void>(
				(_tui, theme, _keys, done) => new TeamTasksOverlay(teams, theme, done),
				{ overlay: true },
			);
		},
	});
}
