import { NotificationMessage } from "./events";
import type { Reference } from "@rilldata/web-common/features/models/utils/get-table-references";

class EventBus {
  private listeners: EventMap = new Map();

  on<Event extends T>(event: Event, callback: Listener<Event>) {
    const key = crypto.randomUUID();
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

export const eventBus = new EventBus();

export interface Events {
  notification: NotificationMessage;
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
