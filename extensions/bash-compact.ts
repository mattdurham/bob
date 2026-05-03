/**
 * Compact bash tool renderer.
 *
 * Overrides the default bash tool display:
 *   - Shows the command being run
 *   - ✅ on success (exit 0) — no output shown
 *   - 🔴 on error — shows full output inline
 *   - Ctrl+E / click expands output for any result
 *
 * The agent still receives the full output regardless.
 */

import type {
	BashToolDetails,
	ExtensionAPI,
} from "@mariozechner/pi-coding-agent";
import { createBashTool } from "@mariozechner/pi-coding-agent";
import { Text } from "@mariozechner/pi-tui";

export default function (pi: ExtensionAPI) {
	const originalBash = createBashTool(process.cwd());

	pi.registerTool({
		name: "bash",
		label: "bash",
		description: originalBash.description,
		parameters: originalBash.parameters,

		async execute(toolCallId, params, signal, onUpdate) {
			return originalBash.execute(toolCallId, params, signal, onUpdate);
		},

		renderCall(args, theme) {
			const cmd =
				args.command.length > 120
					? `${args.command.slice(0, 117)}...`
					: args.command;
			let text = theme.fg("toolTitle", theme.bold("$ "));
			text += theme.fg("accent", cmd);
			if (args.timeout) text += theme.fg("dim", ` (timeout: ${args.timeout}s)`);
			return new Text(text, 0, 0);
		},

		renderResult(result, { expanded, isPartial }, theme) {
			if (isPartial) return new Text(theme.fg("dim", "running..."), 0, 0);

			const details = result.details as BashToolDetails | undefined;
			const content = result.content[0];
			const output = content?.type === "text" ? content.text : "";
			const isError = result.isError === true;

			// Extract exit code if present in output
			const exitMatch = output.match(/\nexit code: (\d+)\s*$/);
			const exitCode = exitMatch ? parseInt(exitMatch[1], 10) : isError ? 1 : 0;
			const success = exitCode === 0 && !isError;

			if (success && !expanded) {
				// Silent success — just a checkmark
				let text = theme.fg("success", "✓");
				if (details?.truncation?.truncated) {
					text += theme.fg("dim", " (output truncated)");
				}
				return new Text(text, 0, 0);
			}

			// Error or expanded — show output
			const lines = output.split("\n");
			const displayLines = expanded ? lines : lines.slice(0, 30);
			const truncated = !expanded && lines.length > 30;

			let text = success
				? theme.fg("success", "✓ (expanded)")
				: theme.fg("error", `✗ exit ${exitCode}`);

			if (details?.truncation?.truncated) {
				text += theme.fg("warning", " [output truncated]");
			}

			for (const line of displayLines) {
				text +=
					"\n" + (success ? theme.fg("dim", line) : theme.fg("error", line));
			}

			if (truncated) {
				text +=
					"\n" +
					theme.fg(
						"muted",
						`... ${lines.length - 30} more lines (ctrl+e to expand)`,
					);
			}

			return new Text(text, 0, 0);
		},
	});
}
