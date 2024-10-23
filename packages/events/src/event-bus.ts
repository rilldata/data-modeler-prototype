import type { Reference } from "@rilldata/web-common/features/models/utils/get-table-references";
import type { BannerMessage, NotificationMessage } from "./events";

class EventBus {
  private listeners: EventMap = new Map();

  on<Event extends T>(event: Event, callback: Listener<Event>) {
    const key = generateUUID();
    const eventMap = this.listeners.get(event);

    if (!eventMap) {
      this.listeners.set(
        event,
        new Map<string, Listener<T>>([[key, callback]]),
      );
    } else {
      eventMap.set(key, callback);
    }

    const unsubscribe = () => this.listeners.get(event)?.delete(key);

    return unsubscribe;
  }

  emit<Event extends T>(event: Event, payload: Events[Event]) {
    const listeners = this.listeners.get(event);

    listeners?.forEach((cb) => {
      cb(payload);
    });
  }
}

function generateUUID(): string {
  // Generate random numbers for the UUID
  const randomNumbers: number[] = new Array(16)
    .fill(0)
    .map(() => Math.floor(Math.random() * 256));

  // Set the version and variant bits
  randomNumbers[6] = (randomNumbers[6] & 0x0f) | 0x40; // Version 4
  randomNumbers[8] = (randomNumbers[8] & 0x3f) | 0x80; // Variant 10

  // Convert to hexadecimal and format as a UUID
  const hexDigits: string = randomNumbers
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
  return `${hexDigits.slice(0, 8)}-${hexDigits.slice(8, 12)}-${hexDigits.slice(12, 16)}-${hexDigits.slice(16, 20)}-${hexDigits.slice(20, 32)}`;
}

export const eventBus = new EventBus();

export interface Events {
  notification: NotificationMessage;
  banner: BannerMessage | null;
  highlightSelection: Reference[];
  "shift-click": null;
  "command-click": null;
  click: null;
  "shift-command-click": null;
}

type T = keyof Events;

type Listener<EventType extends T> = (e: Events[EventType]) => void;

type EventMap = Map<T, Listeners>;

type Listeners = Map<string, Listener<T>>;
