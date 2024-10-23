import { Throttler, asyncWait } from "@rilldata/utils";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/fetch-streaming-wrapper";
import type {
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client/index";
import { get, writable } from "svelte/store";
import { runtime } from "./runtime-store";

const MAX_RETRIES = 5;
const BACKOFF_DELAY = 1000;
const RETRY_COUNT_DELAY = 500;
const RECONNECT_CALLBACK_DELAY = 150;

type WatchResponse =
  | V1WatchFilesResponse
  | V1WatchResourcesResponse
  | V1WatchLogsResponse;

type StreamingFetchResponse<Res extends WatchResponse> = {
  result?: Res;
  error?: { code: number; message: string };
};

type EventMap<T> = {
  response: T;
  reconnect: void;
};

type Listeners<T> = Map<keyof EventMap<T>, Callback<T, keyof EventMap<T>>[]>;

type Callback<T, K extends keyof EventMap<T>> = (
  eventData: EventMap<T>[K],
) => void | Promise<void>;

export class WatchRequestClient<Res extends WatchResponse> {
  private url: string | undefined;
  private controller: AbortController | undefined;
  private stream: AsyncGenerator<StreamingFetchResponse<Res>> | undefined;
  private outOfFocusThrottler = new Throttler(120000, 30000);
  public retryAttempts = writable(0);
  private reconnectTimeout: ReturnType<typeof setTimeout> | undefined;
  private retryTimeout: ReturnType<typeof setTimeout> | undefined;
  private listeners: Listeners<Res> = new Map([
    ["response", []],
    ["reconnect", []],
  ]);
  public closed = writable(false);

  public on<K extends keyof EventMap<Res>>(
    event: K,
    listener: Callback<Res, K>,
  ) {
    this.listeners.get(event)?.push(listener);
  }

  public heartbeat = () => {
    if (get(this.closed)) {
      this.reconnect().catch(console.error);
    }
    this.throttle();
  };

  public watch(url: string) {
    this.cancel();
    this.url = url;

    this.listen().catch(console.error);

    // Start throttling after the first connection
    this.throttle();
  }

  public close = () => {
    this.cancel();
    this.closed.set(true);
  };

  public throttle(prioritize: boolean = false) {
    this.outOfFocusThrottler.throttle(this.close, prioritize);
  }

  private async reconnect() {
    clearTimeout(this.reconnectTimeout);

    if (this.outOfFocusThrottler.isThrottling()) {
      this.outOfFocusThrottler.cancel();
    }

    // The stream was not cancelled, so don't reconnect
    if (this.controller && !this.controller.signal.aborted) return;

    const currentAttempts = get(this.retryAttempts);

    if (currentAttempts >= MAX_RETRIES) throw new Error("Max retries exceeded");

    if (currentAttempts > 0) {
      const delay = BACKOFF_DELAY * 2 ** currentAttempts;
      await asyncWait(delay);
    }

    this.retryAttempts.update((n) => n + 1);
    this.listen(true).catch(console.error);
  }

  private cancel() {
    this.controller?.abort();
    this.stream = this.controller = undefined;
  }

  private async listen(reconnect = false) {
    clearTimeout(this.reconnectTimeout);

    if (!this.url) throw new Error("URL not set");

    this.controller = new AbortController();
    this.stream = this.getFetchStream(this.url, this.controller);

    try {
      this.retryTimeout = setTimeout(() => {
        this.retryAttempts.set(0);
      }, RETRY_COUNT_DELAY);

      if (reconnect) {
        this.reconnectTimeout = setTimeout(() => {
          this.listeners.get("reconnect")?.forEach((cb) => void cb());
        }, RECONNECT_CALLBACK_DELAY);
      }

      this.closed.set(false);
      for await (const res of this.stream) {
        if (this.controller?.signal.aborted) break;
        if (res.error) throw new Error(res.error.message);

        if (res.result)
          this.listeners.get("response")?.forEach((cb) => void cb(res.result));
      }
    } catch {
      clearTimeout(this.retryTimeout);

      this.cancel();
      if (get(this.closed)) return;

      this.reconnect().catch((e) => {
        throw new Error(e);
      });
    }
  }

  private getFetchStream(url: string, controller: AbortController) {
    const headers = { "Content-Type": "application/json" };
    const jwt = get(runtime).jwt;
    if (jwt) {
      headers["Authorization"] = `Bearer ${jwt.token}`;
    }

    return streamingFetchWrapper<StreamingFetchResponse<Res>>(
      url,
      "GET",
      undefined,
      headers,
      controller.signal,
    );
  }
}
