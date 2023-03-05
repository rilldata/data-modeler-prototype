import type { FetchWrapperOptions } from "@rilldata/web-local/lib/util/fetchWrapper";
import { get } from "svelte/store";
import { HttpRequestQueue } from "./http-request-queue/HttpRequestQueue";
import { runtime } from "./runtime-store";

/**
 * Runtime base URL
 *  Local
 *    In dev & prod: http://localhost:9009
 *  Cloud
 *    In dev: http://localhost:9009
 *    In prod: https://{region}.runtime.rilldata.com
 */

export const httpRequestQueue = new HttpRequestQueue();

export const httpClient = async <T>(
  config: FetchWrapperOptions
): Promise<T> => {
  // naive request interceptor
  const host = get(runtime).host;
  const interceptedConfig = { ...config, baseUrl: host };

  return (await httpRequestQueue.add(interceptedConfig)) as Promise<T>;
};

export default httpClient;
