import { type Events, eventBus } from "@rilldata/events";

type Handler = (e: MouseEvent) => void | Promise<void>;
type Modifier = "command" | "shift" | "shift-command" | "click";
type Params = Partial<Record<Modifier, Handler>>;

export function modified(params: Params) {
  return async (e: MouseEvent) => {
    const { ctrlKey, shiftKey, metaKey } = e;

    let handler: Handler | null = null;
    let modifier: Modifier | null = null;

    if ((ctrlKey || metaKey) && shiftKey && params["shift-command"]) {
      e.preventDefault();
      handler = params["shift-command"];
      modifier = "shift-command";
    } else if ((ctrlKey || metaKey) && params.command) {
      e.preventDefault();
      handler = params.command;
      modifier = "command";
    } else if (shiftKey && params.shift) {
      e.preventDefault();
      handler = params.shift;
      modifier = "shift";
    } else if (params.click) {
      handler = params.click;
    }

    const event = (modifier ? `${modifier}-click` : "click") as keyof Events;
    eventBus.emit(event, null);

    if (handler) {
      await handler(e);
    }
  };
}
