/**
 * In-memory message bus with per-agent mailboxes and broadcast support.
 *
 * Module-level singleton so all agent sessions in the same process share
 * the same bus — including spawned subagents that load this extension.
 */

export interface BusMessage {
  id: string;
  from: string;
  to: string; // agent name or "broadcast"
  content: string;
  timestamp: number;
  read: boolean;
}

export class MessageBus {
  private mailboxes = new Map<string, BusMessage[]>();
  private listeners = new Map<string, Array<(msg: BusMessage) => void>>();

  private box(name: string): BusMessage[] {
    if (!this.mailboxes.has(name)) this.mailboxes.set(name, []);
    return this.mailboxes.get(name)!;
  }

  send(from: string, to: string, content: string): BusMessage {
    const msg: BusMessage = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      from,
      to,
      content,
      timestamp: Date.now(),
      read: false,
    };
    this.box(to).push(msg);
    this.listeners.get(to)?.forEach((cb) => cb(msg));
    return msg;
  }

  broadcast(from: string, content: string, recipients: string[]): BusMessage[] {
    return recipients
      .filter((name) => name !== from)
      .map((name) => this.send(from, name, `[broadcast from ${from}] ${content}`));
  }

  /** Returns unread messages for the given mailbox and marks them read. */
  receive(name: string): BusMessage[] {
    const unread = this.box(name).filter((m) => !m.read);
    unread.forEach((m) => (m.read = true));
    return unread;
  }

  /** Returns all messages (read and unread) for the given mailbox. */
  all(name: string): BusMessage[] {
    return [...this.box(name)];
  }

  /** Subscribe to new messages arriving in a mailbox. Returns an unsubscribe fn. */
  subscribe(name: string, callback: (msg: BusMessage) => void): () => void {
    if (!this.listeners.has(name)) this.listeners.set(name, []);
    const cbs = this.listeners.get(name)!;
    cbs.push(callback);
    return () => {
      const idx = cbs.indexOf(callback);
      if (idx >= 0) cbs.splice(idx, 1);
    };
  }

  mailboxNames(): string[] {
    return Array.from(this.mailboxes.keys());
  }

  /** Clear all state — called on session shutdown. */
  reset(): void {
    this.mailboxes.clear();
    this.listeners.clear();
  }
}
