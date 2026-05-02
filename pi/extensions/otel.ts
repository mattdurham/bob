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
  traceId: string;       // 32 lowercase hex chars
  spanId: string;        // 16 lowercase hex chars
  parentSpanId?: string; // 16 lowercase hex chars
  name: string;
  kind: number;          // 1 = INTERNAL
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

function newTraceId(): string { return crypto.randomBytes(16).toString("hex"); }
function newSpanId(): string  { return crypto.randomBytes(8).toString("hex");  }

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
      if (input.agent)      out.push(attr("subagent.agent",   String(input.agent)));
      if (input.background) out.push(attr("subagent.mode",    "background"));
      else if (input.tasks) out.push(attr("subagent.mode",    "parallel"));
      else if (input.chain) out.push(attr("subagent.mode",    "chain"));
      else                  out.push(attr("subagent.mode",    "foreground"));
      // Include task description, truncated
      if (input.task)       out.push(attr("subagent.task",    String(input.task).slice(0, 500)));
      break;

    case "agent_wait":
      if (Array.isArray(input.agents))
        out.push(attr("agent_wait.agents", (input.agents as string[]).join(",")));
      if (input.timeout_ms)
        out.push(attr("agent_wait.timeout_ms", String(input.timeout_ms)));
      break;

    case "agent_status":
      break; // no params worth recording

    case "TaskCreate":
      if (input.title)  out.push(attr("task.title",  String(input.title)));
      if (input.owner)  out.push(attr("task.owner",  String(input.owner)));
      break;

    case "TaskUpdate":
      if (input.id)     out.push(attr("task.id",     String(input.id)));
      if (input.status) out.push(attr("task.status", String(input.status)));
      if (input.owner)  out.push(attr("task.owner",  String(input.owner)));
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
          attributes: [attr("service.name", "pi-bob-agents")],
        },
        scopeSpans: [
          {
            scope: { name: "pi-bob-agents", version: "1.0.0" },
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
  const user     = process.env.OTEL_USER?.trim();
  const token    = process.env.OTEL_TOKEN?.trim();

  // All three must be set — if any missing, disable silently
  if (!endpoint || !user || !token) return;

  const tracesUrl =
    endpoint.endsWith("/v1/traces")
      ? endpoint
      : `${endpoint.replace(/\/$/, "")}/v1/traces`;

  const authHeader =
    "Basic " + Buffer.from(`${user}:${token}`).toString("base64");

  // ─── Mutable state ─────────────────────────────────────────────────────────

  const buffer: OtelSpan[] = [];
  let timer: ReturnType<typeof setInterval> | null = null;

  let rootTraceId  = "";
  let sessionSpan: OtelSpan | null = null;
  let agentSpan:   OtelSpan | null = null;

  // ─── Batching ───────────────────────────────────────────────────────────────

  function enqueue(span: OtelSpan): void {
    buffer.push(span);
    if (buffer.length >= 100) drainAndFlush();
  }

  function drainAndFlush(): void {
    if (buffer.length === 0) return;
    const batch = buffer.splice(0);
    void flush(tracesUrl, authHeader, batch);
  }

  // ─── Session lifecycle ──────────────────────────────────────────────────────

  pi.on("session_start", async (_event, ctx) => {
    rootTraceId  = newTraceId();
    const now    = nowNano();
    sessionSpan  = {
      traceId:          rootTraceId,
      spanId:           newSpanId(),
      name:             "pi.session",
      kind:             1,
      startTimeUnixNano: now,
      endTimeUnixNano:   now, // updated on shutdown
      attributes:       [attr("session.cwd", ctx.cwd)],
      events:           [],
      status:           { code: 0 },
    };

    timer = setInterval(drainAndFlush, 10_000);
  });

  pi.on("session_shutdown", async () => {
    if (timer) { clearInterval(timer); timer = null; }

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
    // Close any dangling agent span
    if (agentSpan) {
      agentSpan.endTimeUnixNano = nowNano();
      enqueue(agentSpan);
    }

    const now = nowNano();
    agentSpan = {
      traceId:           rootTraceId,
      spanId:            newSpanId(),
      parentSpanId:      sessionSpan?.spanId,
      name:              "pi.agent.loop",
      kind:              1,
      startTimeUnixNano: now,
      endTimeUnixNano:   now, // updated on agent_end
      attributes:        [],
      events: [
        {
          name:           "user.message",
          timeUnixNano:   now,
          attributes:     [attr("message.content", event.prompt.slice(0, 2000))],
        },
      ],
      status: { code: 0 },
    };
  });

  pi.on("agent_end", async (event) => {
    if (!agentSpan) return;

    // Record the final assistant message
    const messages = event.messages as Array<{ role: string; content: unknown }>;
    for (let i = messages.length - 1; i >= 0; i--) {
      const msg = messages[i];
      if (msg.role === "assistant") {
        const text = extractText(msg.content).trim();
        if (text) {
          agentSpan.events.push({
            name:         "assistant.message",
            timeUnixNano: nowNano(),
            attributes:   [attr("message.content", text.slice(0, 2000))],
          });
        }
        break;
      }
    }

    agentSpan.endTimeUnixNano = nowNano();
    enqueue(agentSpan);
    agentSpan = null;
  });

  // ─── Tool calls ─────────────────────────────────────────────────────────────

  pi.on("tool_call", async (event) => {
    // Skip file-operation tools entirely
    if (SKIP_TOOLS.has(event.toolName)) return;
    if (!agentSpan) return;

    const input =
      "input" in event
        ? (event.input as Record<string, unknown>)
        : {};

    agentSpan.events.push({
      name:         "tool.call",
      timeUnixNano: nowNano(),
      attributes:   toolAttrs(event.toolName, input),
    });
  });
}
