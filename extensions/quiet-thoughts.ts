/**
 * quiet-thoughts — suppress assistant preamble text on turns that contain tool calls.
 *
 * When the model writes "Let me look at X..." before calling a tool, that text
 * is noise. This extension strips text content from assistant messages that also
 * contain tool calls, so the UI only shows the tool invocations + final answers.
 *
 * The model's tool calls are preserved in context so it can reference results.
 * Only turns with NO tool calls (i.e. final answers) show text.
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

export default function (pi: ExtensionAPI) {
	pi.on("message_end", async (event) => {
		if (event.message.role !== "assistant") return;

		const content = event.message.content as Array<{ type: string }>;
		if (!Array.isArray(content)) return;

		const hasToolCall = content.some((c) => c.type === "toolCall");
		if (!hasToolCall) return; // final answer — show it

		// Strip text blocks, keep tool calls
		const stripped = content.filter(
			(c) => c.type !== "text" && c.type !== "thinking",
		);
		if (stripped.length === content.length) return; // nothing to strip

		return {
			message: {
				...event.message,
				content: stripped,
			},
		};
	});
}
