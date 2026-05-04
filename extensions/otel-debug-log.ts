/**
 * Tiny helper imported by flush() in otel.ts when debug logging is needed.
 * Kept separate so string literals never get mangled by patch scripts.
 */
import * as fs from "node:fs";

const LOG = "/tmp/otel_debug.log";
const NL = "\n";

export function logFlush(spans: { name: string; traceId: string; parentSpanId?: string; attributes: { key: string }[]; events: unknown[] }[]): void {
  try {
    const parts: string[] = [new Date().toISOString() + " flush " + String(spans.length) + " spans"];
    for (const s of spans) {
      parts.push(
        "  " + s.name +
        " traceId=" + s.traceId.slice(0, 8) +
        " parent=" + (s.parentSpanId ?? "none") +
        " attrs=[" + s.attributes.map((a) => a.key).join(",") + "]" +
        " events=" + String(s.events.length),
      );
    }
    fs.appendFileSync(LOG, parts.join(NL) + NL);
  } catch { /* ignore */ }
}

export function logHttp(status: number, statusText: string): void {
  try { fs.appendFileSync(LOG, new Date().toISOString() + " HTTP " + String(status) + " " + statusText + NL); } catch { /* ignore */ }
}

export function logError(label: string, detail: string): void {
  try { fs.appendFileSync(LOG, new Date().toISOString() + " " + label + ": " + detail + NL); } catch { /* ignore */ }
}
