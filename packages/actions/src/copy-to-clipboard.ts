import { eventBus } from "@rilldata/events";

/**
 * The Clipboard API is only available in secure contexts.
 * So, a self-hosted Rill Developer instance served over HTTP (not HTTPS) will not have access to the Clipboard API.
 * See: https://developer.mozilla.org/en-US/docs/Web/API/Clipboard
 */
export function isClipboardApiSupported(): boolean {
  return !!navigator.clipboard;
}

export function copyToClipboard(value: string, message?: string) {
  navigator.clipboard.writeText(value).catch(console.error);
  eventBus.emit("notification", {
    message: message || `copied ${value} to clipboard`,
  });
}
