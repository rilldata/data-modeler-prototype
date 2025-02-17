import type { AxiosError } from "axios";
import type { RpcStatus } from "../../client";

// Temporary function to get the correct error message.
// We get AxiosError<RpcStatus> but type is RpcStatus
// TODO: fix the root types in orval
export function getRpcErrorMessage(
  error: RpcStatus | undefined,
): string | undefined {
  const mappedError = error as unknown as AxiosError<RpcStatus> | undefined;
  return mappedError?.response?.data?.message;
}
