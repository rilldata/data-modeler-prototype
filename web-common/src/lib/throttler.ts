export class Throttler {
  private timer: ReturnType<typeof setTimeout> | undefined;
  private callback: () => void | Promise<void>;

  public constructor(private readonly timeout: number) {}

  public throttle(callback: () => void | Promise<void>) {
    this.cancel();
    this.callback = callback;

    this.timer = setTimeout(() => {
      this.timer = undefined;
      this.callback();
    }, this.timeout);
  }

  public isThrottling() {
    return this.timer !== undefined;
  }

  public cancel() {
    if (!this.timer) return;

    clearTimeout(this.timer);
    this.timer = undefined;
  }
}
