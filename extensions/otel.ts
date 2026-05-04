/**
 * OpenTelemetry trace exporter for pi sessions.
 *
 * Enabled only when all three env vars are set:
 *   OTEL_ENDPOINT  — OTLP HTTP base URL (e.g. https://otlp-gateway-prod-us-east-0.grafana.net/otlp)
 *   OTEL_USER      — Basic-auth username
 *   OTEL_TOKEN     — Basic-auth password / token
 *
 * What is recorded:
 *   - Session span  (root)
 *   - Agent-loop spans (child of session, one per user prompt)
 *   - user.message events  — prompt text, truncated to 2000 chars
 *   - assistant.message events — final reply text, truncated to 2000 chars
 *   - tool.call events — name + non-sensitive params for custom/coordination tools
 *   - bash tool name only (no command content)
 *
 * What is NOT recorded:
 *   - read / write / edit / find / grep / ls tool calls (file content)
 *   - bash command strings
 *   - file paths or file contents
 *
 * Batching: flush when buffer reaches 100 spans, or every 10 s.
 */

import * as crypto from "node:crypto";
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

// ─── Shared trace context (read by bob-agents to propagate to subagents) ────────

declare global {
  // biome-ignore lint: cross-extension sharing via globalThis
  var __bobOtelCtx: { traceId: string; spanId: string } | undefined;
}

// ─── OTLP types (minimal, JSON encoding) ─────────────────────────────────────

interface Attr {
	key: string;
	value: { stringValue: string };
}

interface OtelEvent {
	name: string;
	timeUnixNano: string;
	attributes: Attr[];
}

interface OtelSpan {
	traceId: string; // 32 lowercase hex chars
	spanId: string; // 16 lowercase hex chars
	parentSpanId?: string; // 16 lowercase hex chars
	name: string;
	kind: number; // 1 = INTERNAL
	startTimeUnixNano: string;
	endTimeUnixNano: string;
	attributes: Attr[];
	events: OtelEvent[];
	status: { code: number };
}

// ─── Tool filtering ───────────────────────────────────────────────────────────

/** Built-in file-operation tools — skip entirely (content is sensitive). */
const SKIP_TOOLS = new Set(["read", "write", "edit", "find", "grep", "ls"]);

// ─── Helpers ──────────────────────────────────────────────────────────────────

function attr(key: string, value: string): Attr {
	return { key, value: { stringValue: value } };
}

function nowNano(): string {
	// Date.now() is ms; append 6 zeros for nanoseconds (ms precision is fine for tracing)
	return Date.now().toString() + "000000";
}

function newTraceId(): string {
	return crypto.randomBytes(16).toString("hex");
}
function newSpanId(): string {
	return crypto.randomBytes(8).toString("hex");
}

function extractText(content: unknown): string {
	if (typeof content === "string") return content;
	if (!Array.isArray(content)) return "";
	return (content as Array<{ type?: string; text?: string }>)
		.filter((b) => b?.type === "text" && typeof b.text === "string")
		.map((b) => b.text as string)
		.join("");
}

/** Build sanitised attributes for a tool call event. */
function toolAttrs(toolName: string, input: Record<string, unknown>): Attr[] {
	const out: Attr[] = [attr("tool.name", toolName)];

	switch (toolName) {
		// Coordination tools — record intent, not content
		case "subagent":
			if (input.agent) out.push(attr("subagent.agent", String(input.agent)));
			if (input.background) out.push(attr("subagent.mode", "background"));
			else if (input.tasks) out.push(attr("subagent.mode", "parallel"));
			else if (input.chain) out.push(attr("subagent.mode", "chain"));
			else out.push(attr("subagent.mode", "foreground"));
			// Include task description, truncated
			if (input.task)
				out.push(attr("subagent.task", String(input.task).slice(0, 500)));
			break;

		case "agent_wait":
			if (Array.isArray(input.agents))
				out.push(
					attr("agent_wait.agents", (input.agents as string[]).join(",")),
				);
			if (input.timeout_ms)
				out.push(attr("agent_wait.timeout_ms", String(input.timeout_ms)));
			break;

		case "agent_status":
			break; // no params worth recording

		case "TaskCreate":
			if (input.title) out.push(attr("task.title", String(input.title)));
			if (input.owner) out.push(attr("task.owner", String(input.owner)));
			break;

		case "TaskUpdate":
			if (input.id) out.push(attr("task.id", String(input.id)));
			if (input.status) out.push(attr("task.status", String(input.status)));
			if (input.owner) out.push(attr("task.owner", String(input.owner)));
			break;

		case "TaskGet":
		case "TaskList":
			break; // read-only, no mutation to record

		case "mailbox_send_as":
		case "mailbox_send":
			if (input.to) out.push(attr("mailbox.to", String(input.to)));
			// message content omitted (may contain task details / sensitive data)
			break;

		case "mailbox_broadcast":
			out.push(attr("mailbox.broadcast", "true"));
			break;

		case "mailbox_read":
			break; // read-only

		case "bash":
			// Record that bash ran — command string omitted (may expose secrets / file content)
			out.push(attr("bash.command", "(omitted)"));
			break;

		default:
			// Unknown custom tool — record name only
			break;
	}

	return out;
}

/** POST a batch of spans to the OTLP/HTTP traces endpoint. */
async function flush(
	url: string,
	authHeader: string,
	spans: OtelSpan[],
): Promise<void> {
	if (spans.length === 0) return;
	const body = JSON.stringify({
		resourceSpans: [
			{
				resource: {
					attributes: [attr("service.name", "bob")],
				},
				scopeSpans: [
					{
						scope: { name: "bob", version: "1.0.0" },
						spans,
					},
				],
			},
		],
	});

	try {
		await fetch(url, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				Authorization: authHeader,
			},
			body,
			signal: AbortSignal.timeout(10_000),
		});
	} catch {
		// Telemetry failures must never surface to the user
	}
}

// ─── Extension ────────────────────────────────────────────────────────────────

export default function (pi: ExtensionAPI) {
	const endpoint = process.env.OTEL_ENDPOINT?.trim();
	const user = process.env.OTEL_USER?.trim();
	const token = process.env.OTEL_TOKEN?.trim();

	// All three must be set — if any missing, disable silently
	if (!endpoint || !user || !token) return;

	const tracesUrl = endpoint.endsWith("/v1/traces")
		? endpoint
		: `${endpoint.replace(/\/$/, "")}/v1/traces`;

	const authHeader =
		"Basic " + Buffer.from(`${user}:${token}`).toString("base64");

	// ─── Child vs root mode ────────────────────────────────────────────────────
	// If __bobOtelCtx is set we are running inside a spawned subagent.
	// Use the parent's traceId and parent spanId — no session span, no timer.
	const parentCtx = globalThis.__bobOtelCtx;
	const isChild = !!parentCtx;

	// ─── Mutable state ─────────────────────────────────────────────────────────

	const buffer: OtelSpan[] = [];
	let timer: ReturnType<typeof setInterval> | null = null;

	let rootTraceId = isChild ? parentCtx!.traceId : "";
	let sessionSpan: OtelSpan | null = null;
	let agentSpan: OtelSpan | null = null;
	let currentModel = "unknown";
	// In child mode the parent span is the bob-agents subagent tool span
	const childParentSpanId = isChild ? parentCtx!.spanId : undefined;

	// ─── Batching ───────────────────────────────────────────────────────────────

	function enqueue(span: OtelSpan): void {
		buffer.push(span);
		if (buffer.length >= 100) drainAndFlush();
	}

	function drainAndFlush(): void {
		if (buffer.length === 0) return;
		const batch = buffer.splice(0);
		const debug = process.env.OTEL_DEBUG === "1";
	if (debug) {
		for (const s of batch) {
			process.stderr.write(`[otel] span: ${s.name} attrs=${s.attributes.map(a => a.key).join(",")} events=${s.events.length}\n`);
		}
	}
	void flush(tracesUrl, authHeader, batch);
	}

	// ─── Session lifecycle ──────────────────────────────────────────────────────

	pi.on("session_start", async (_event, ctx) => {
		// Start flush timer once — /reload fires session_start again
		if (!timer) timer = setInterval(drainAndFlush, 10_000);
		// Child: inherit parent trace, no session span needed
		if (isChild) return;
		// Root: only initialise once so /reload keeps spans in the same trace
		if (sessionSpan) return;
		rootTraceId = newTraceId();
		const now = nowNano();
		sessionSpan = {
			traceId: rootTraceId,
			spanId: newSpanId(),
			name: "pi.session",
			kind: 1,
			startTimeUnixNano: now,
			endTimeUnixNano: now, // updated on shutdown
			attributes: [attr("session.cwd", ctx.cwd)],
			events: [],
			status: { code: 0 },
		};
	});

	pi.on("session_shutdown", async () => {
		if (timer) {
			clearInterval(timer);
			timer = null;
		}

		// Close any open agent span (shouldn't happen normally)
		if (agentSpan) {
			agentSpan.endTimeUnixNano = nowNano();
			enqueue(agentSpan);
			agentSpan = null;
		}

		if (sessionSpan) {
			sessionSpan.endTimeUnixNano = nowNano();
			enqueue(sessionSpan);
			sessionSpan = null;
		}

		drainAndFlush();
	});

	// ─── Agent loop ─────────────────────────────────────────────────────────────

	pi.on("before_agent_start", async (event) => {
		if (process.env.OTEL_DEBUG === "1") process.stderr.write(`[otel] before_agent_start prompt_len=${event.prompt.length}\n`);
		// Close any dangling agent span
		if (agentSpan) {
			agentSpan.endTimeUnixNano = nowNano();
			enqueue(agentSpan);
		}

		const now = nowNano();
		const agentSpanId = newSpanId();
		globalThis.__bobOtelCtx = { traceId: rootTraceId, spanId: agentSpanId };
		agentSpan = {
			traceId: rootTraceId,
			spanId: agentSpanId,
			parentSpanId: isChild ? childParentSpanId : sessionSpan?.spanId,
			name: "pi.agent.loop",
			kind: 1,
			startTimeUnixNano: now,
			endTimeUnixNano: now, // updated on agent_end
			attributes: [],
			events: [
				{
					name: "user.message",
					timeUnixNano: now,
					attributes: [attr("message.content", event.prompt.slice(0, 2000))],
				},
			],
			status: { code: 0 },
		};
	});

	pi.on("model_select", async (event) => {
		if (process.env.OTEL_DEBUG === "1") process.stderr.write(`[otel] model_select: ${event.model?.id ?? event.model?.name ?? 'unknown'}\n`);
		currentModel = event.model?.id ?? event.model?.name ?? "unknown";
		if (sessionSpan) {
			// Update model on session span so the last-used model is always recorded
			const existing = sessionSpan.attributes.findIndex(
				(a) => a.key === "model",
			);
			if (existing >= 0)
				sessionSpan.attributes[existing] = attr("model", currentModel);
			else sessionSpan.attributes.push(attr("model", currentModel));
		}
	});

	pi.on("agent_end", async (event) => {
		if (process.env.OTEL_DEBUG === "1") process.stderr.write(`[otel] agent_end model=${currentModel} msgs=${event.messages.length} span_events=${agentSpan?.events.length ?? 0}\n`);
		if (!agentSpan) return;

		// Record model on the span
		agentSpan.attributes.push(attr("model", currentModel));

		// Sum token usage across all assistant messages in this loop
		type MsgWithUsage = {
			role: string;
			content: unknown;
			usage?: {
				input?: number;
				output?: number;
				cacheRead?: number;
				cacheWrite?: number;
				totalTokens?: number;
				cost?: { input?: number; output?: number; total?: number };
			};
		};
		const messages = event.messages as MsgWithUsage[];
		let tokensIn = 0,
			tokensOut = 0,
			cacheRead = 0,
			cacheWrite = 0;
		let costIn = 0,
			costOut = 0;
		let lastText = "";

		for (let i = messages.length - 1; i >= 0; i--) {
			const msg = messages[i];
			if (msg.role === "assistant") {
				if (msg.usage) {
					tokensIn += msg.usage.input ?? 0;
					tokensOut += msg.usage.output ?? 0;
					cacheRead += msg.usage.cacheRead ?? 0;
					cacheWrite += msg.usage.cacheWrite ?? 0;
					costIn += msg.usage.cost?.input ?? 0;
					costOut += msg.usage.cost?.output ?? 0;
				}
				if (!lastText) lastText = extractText(msg.content).trim();
			}
		}

		// Token / cost attributes
		if (tokensIn || tokensOut) {
			agentSpan.attributes.push(
				attr("tokens.input", String(tokensIn)),
				attr("tokens.output", String(tokensOut)),
				attr("tokens.total", String(tokensIn + tokensOut)),
			);
		}
		if (cacheRead || cacheWrite) {
			agentSpan.attributes.push(
				attr("tokens.cache_read", String(cacheRead)),
				attr("tokens.cache_write", String(cacheWrite)),
			);
		}
		if (costIn || costOut) {
			agentSpan.attributes.push(
				attr("cost.input", costIn.toFixed(6)),
				attr("cost.output", costOut.toFixed(6)),
				attr("cost.total", (costIn + costOut).toFixed(6)),
			);
		}

		// Final assistant reply
		if (lastText) {
			agentSpan.events.push({
				name: "assistant.message",
				timeUnixNano: nowNano(),
				attributes: [attr("message.content", lastText.slice(0, 2000))],
			});
		}

		agentSpan.endTimeUnixNano = nowNano();
		enqueue(agentSpan);
		agentSpan = null;
		globalThis.__bobOtelCtx = undefined;
	});

	// ─── Tool calls ─────────────────────────────────────────────────────────────

	pi.on("tool_call", async (event) => {
		// Skip file-operation tools entirely
		if (SKIP_TOOLS.has(event.toolName)) return;
		if (!agentSpan) return;

		const input =
			"input" in event ? (event.input as Record<string, unknown>) : {};

		agentSpan.events.push({
			name: "tool.call",
			timeUnixNano: nowNano(),
			attributes: toolAttrs(event.toolName, input),
		});
	});
}
