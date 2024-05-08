/**
 * Generated by orval v6.12.0 🍺
 * Do not edit manually.
 * rill/admin/v1/ai.proto
 * OpenAPI spec version: version not set
 */
import { createMutation } from "@tanstack/svelte-query";
import type {
  CreateMutationOptions,
  MutationFunction,
} from "@tanstack/svelte-query";
import type {
  V1CompleteResponse,
  RpcStatus,
  V1CompleteRequest,
} from "../index.schemas";
import { httpClient } from "../../http-client";

type AwaitedInput<T> = PromiseLike<T> | T;

type Awaited<O> = O extends AwaitedInput<infer T> ? T : never;

/**
 * @summary Complete sends the messages of a chat to the AI and asks it to generate a new message.
 */
export const aIServiceComplete = (v1CompleteRequest: V1CompleteRequest) => {
  return httpClient<V1CompleteResponse>({
    url: `/v1/ai/complete`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: v1CompleteRequest,
  });
};

export type AIServiceCompleteMutationResult = NonNullable<
  Awaited<ReturnType<typeof aIServiceComplete>>
>;
export type AIServiceCompleteMutationBody = V1CompleteRequest;
export type AIServiceCompleteMutationError = RpcStatus;

export const createAIServiceComplete = <
  TError = RpcStatus,
  TContext = unknown,
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof aIServiceComplete>>,
    TError,
    { data: V1CompleteRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof aIServiceComplete>>,
    { data: V1CompleteRequest }
  > = (props) => {
    const { data } = props ?? {};

    return aIServiceComplete(data);
  };

  return createMutation<
    Awaited<ReturnType<typeof aIServiceComplete>>,
    TError,
    { data: V1CompleteRequest },
    TContext
  >(mutationFn, mutationOptions);
};
