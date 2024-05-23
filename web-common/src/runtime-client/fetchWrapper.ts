export type FetchWrapperOptions = {
  baseUrl?: string;
  url: string;
  method: string;
  headers?: HeadersInit;
  params?: Record<string, unknown>;
  data?: any;
  signal?: AbortSignal;
};

export interface HTTPError {
  response: {
    status: number;
    data: {
      message: string;
    };
  };
  message: string;
}

export async function fetchWrapper({
  url,
  method,
  headers,
  data,
  params,
  signal,
}: FetchWrapperOptions) {
  if (signal && signal.aborted) return Promise.reject(new Error("Aborted"));

  headers ??= { "Content-Type": "application/json" };

  if (url.endsWith("default/resource")) {
    console.log("send", url, params);
  }
  url = encodeURI(url);

  if (params) {
    const paramParts: string[] = [];
    for (const p in params) {
      paramParts.push(`${p}=${encodeURIComponent(params[p] as string)}`);
    }
    if (paramParts.length) {
      url = `${url}?${paramParts.join("&")}`;
    }
  }

  const resp = await fetch(url, {
    method,
    ...(data ? { body: serializeBody(data) } : {}),
    headers,
    signal,
  });

  const json = await resp.json();

  if (resp.ok) return json;

  // Return runtime errors in the same form as the Axios client had previously
  if (json?.code && json?.message) {
    return Promise.reject({
      response: {
        status: resp.status,
        data: json,
      },
    });
  } else {
    // Fallback error handling
    const err = new Error();
    (err as any).response = json;
    return Promise.reject(err);
  }
}

function serializeBody(body: BodyInit | Record<string, unknown>): BodyInit {
  return body instanceof FormData ? body : JSON.stringify(body);
}
