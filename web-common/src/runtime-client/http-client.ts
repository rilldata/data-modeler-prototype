import type {
  FetchWrapperOptions,
  HTTPError,
} from "@rilldata/web-common/runtime-client/fetchWrapper";
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
  config: FetchWrapperOptions,
): Promise<T> => {
  // naive request interceptors

  // set host
  const host = get(runtime).host;
  const interceptedConfig = { ...config, baseUrl: host };

  // set jwt
  const jwt = get(runtime).jwt;
  if (jwt) {
    interceptedConfig.headers = {
      ...interceptedConfig.headers,
      Authorization: `Bearer ${jwt}`,
    };
  }

  return (await httpRequestQueue.add(interceptedConfig)) as Promise<T>;
};

export default httpClient;

// This overrides Orval's generated error type. (Orval expects this to be a generic.)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export type ErrorType<Error> = HTTPError;
