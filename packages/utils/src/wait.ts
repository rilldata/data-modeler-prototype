export function asyncWait(time: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, time));
}

export async function waitUntil(
  checkFunc: () => boolean,
  timeout = 30000,
  interval = 250,
): Promise<boolean> {
  const startTime = Date.now();

  while (!checkFunc() && (timeout === -1 || Date.now() - startTime < timeout)) {
    await asyncWait(interval);
  }

  return checkFunc();
}

export async function asyncWaitUntil(
  checkFunc: () => Promise<boolean>,
  timeout = 30000,
  interval = 250,
): Promise<boolean> {
  const startTime = Date.now();

  let check = false;
  while (!check && (timeout === -1 || Date.now() - startTime < timeout)) {
    check = await checkFunc();
    if (check) {
      break;
    }

    await asyncWait(interval);
  }

  return check;
}
