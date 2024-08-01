// @generated by protoc-gen-es v2.0.0 with parameter "target=ts"
// @generated from file rill/runtime/v1/expression.proto (package rill.runtime.v1, syntax proto3)
/* eslint-disable */

import type { GenEnum, GenFile, GenMessage } from "@bufbuild/protobuf/codegenv1";
import { enumDesc, fileDesc, messageDesc } from "@bufbuild/protobuf/codegenv1";
import type { Value } from "@bufbuild/protobuf/wkt";
import { file_google_protobuf_struct } from "@bufbuild/protobuf/wkt";
import { file_validate_validate } from "../../../validate/validate_pb";
import type { Message } from "@bufbuild/protobuf";

/**
 * Describes the file rill/runtime/v1/expression.proto.
 */
export const file_rill_runtime_v1_expression: GenFile = /*@__PURE__*/
  fileDesc("CiByaWxsL3J1bnRpbWUvdjEvZXhwcmVzc2lvbi5wcm90bxIPcmlsbC5ydW50aW1lLnYxIq0BCgpFeHByZXNzaW9uEg8KBWlkZW50GAEgASgJSAASJQoDdmFsGAIgASgLMhYuZ29vZ2xlLnByb3RvYnVmLlZhbHVlSAASKgoEY29uZBgDIAEoCzIaLnJpbGwucnVudGltZS52MS5Db25kaXRpb25IABItCghzdWJxdWVyeRgEIAEoCzIZLnJpbGwucnVudGltZS52MS5TdWJxdWVyeUgAQgwKCmV4cHJlc3Npb24iaQoJQ29uZGl0aW9uEjAKAm9wGAEgASgOMhoucmlsbC5ydW50aW1lLnYxLk9wZXJhdGlvbkII+kIFggECEAESKgoFZXhwcnMYAiADKAsyGy5yaWxsLnJ1bnRpbWUudjEuRXhwcmVzc2lvbiKIAQoIU3VicXVlcnkSEQoJZGltZW5zaW9uGAEgASgJEhAKCG1lYXN1cmVzGAIgAygJEioKBXdoZXJlGAMgASgLMhsucmlsbC5ydW50aW1lLnYxLkV4cHJlc3Npb24SKwoGaGF2aW5nGAQgASgLMhsucmlsbC5ydW50aW1lLnYxLkV4cHJlc3Npb24qiAIKCU9wZXJhdGlvbhIZChVPUEVSQVRJT05fVU5TUEVDSUZJRUQQABIQCgxPUEVSQVRJT05fRVEQARIRCg1PUEVSQVRJT05fTkVREAISEAoMT1BFUkFUSU9OX0xUEAMSEQoNT1BFUkFUSU9OX0xURRAEEhAKDE9QRVJBVElPTl9HVBAFEhEKDU9QRVJBVElPTl9HVEUQBhIQCgxPUEVSQVRJT05fT1IQBxIRCg1PUEVSQVRJT05fQU5EEAgSEAoMT1BFUkFUSU9OX0lOEAkSEQoNT1BFUkFUSU9OX05JThAKEhIKDk9QRVJBVElPTl9MSUtFEAsSEwoPT1BFUkFUSU9OX05MSUtFEAxCuAEKE2NvbS5yaWxsLnJ1bnRpbWUudjFCD0V4cHJlc3Npb25Qcm90b1ABWjJnaXRodWIuY29tL3JpbGxkYXRhL3JpbGwvcmlsbC9ydW50aW1lL3YxO3J1bnRpbWV2MaICA1JSWKoCD1JpbGwuUnVudGltZS5WMcoCD1JpbGxcUnVudGltZVxWMeICG1JpbGxcUnVudGltZVxWMVxHUEJNZXRhZGF0YeoCEVJpbGw6OlJ1bnRpbWU6OlYxYgZwcm90bzM", [file_google_protobuf_struct, file_validate_validate]);

/**
 * @generated from message rill.runtime.v1.Expression
 */
export type Expression = Message<"rill.runtime.v1.Expression"> & {
  /**
   * @generated from oneof rill.runtime.v1.Expression.expression
   */
  expression: {
    /**
     * @generated from field: string ident = 1;
     */
    value: string;
    case: "ident";
  } | {
    /**
     * @generated from field: google.protobuf.Value val = 2;
     */
    value: Value;
    case: "val";
  } | {
    /**
     * @generated from field: rill.runtime.v1.Condition cond = 3;
     */
    value: Condition;
    case: "cond";
  } | {
    /**
     * @generated from field: rill.runtime.v1.Subquery subquery = 4;
     */
    value: Subquery;
    case: "subquery";
  } | { case: undefined; value?: undefined };
};

/**
 * Describes the message rill.runtime.v1.Expression.
 * Use `create(ExpressionSchema)` to create a new message.
 */
export const ExpressionSchema: GenMessage<Expression> = /*@__PURE__*/
  messageDesc(file_rill_runtime_v1_expression, 0);

/**
 * @generated from message rill.runtime.v1.Condition
 */
export type Condition = Message<"rill.runtime.v1.Condition"> & {
  /**
   * @generated from field: rill.runtime.v1.Operation op = 1;
   */
  op: Operation;

  /**
   * @generated from field: repeated rill.runtime.v1.Expression exprs = 2;
   */
  exprs: Expression[];
};

/**
 * Describes the message rill.runtime.v1.Condition.
 * Use `create(ConditionSchema)` to create a new message.
 */
export const ConditionSchema: GenMessage<Condition> = /*@__PURE__*/
  messageDesc(file_rill_runtime_v1_expression, 1);

/**
 * @generated from message rill.runtime.v1.Subquery
 */
export type Subquery = Message<"rill.runtime.v1.Subquery"> & {
  /**
   * @generated from field: string dimension = 1;
   */
  dimension: string;

  /**
   * @generated from field: repeated string measures = 2;
   */
  measures: string[];

  /**
   * @generated from field: rill.runtime.v1.Expression where = 3;
   */
  where?: Expression;

  /**
   * @generated from field: rill.runtime.v1.Expression having = 4;
   */
  having?: Expression;
};

/**
 * Describes the message rill.runtime.v1.Subquery.
 * Use `create(SubquerySchema)` to create a new message.
 */
export const SubquerySchema: GenMessage<Subquery> = /*@__PURE__*/
  messageDesc(file_rill_runtime_v1_expression, 2);

/**
 * @generated from enum rill.runtime.v1.Operation
 */
export enum Operation {
  /**
   * @generated from enum value: OPERATION_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: OPERATION_EQ = 1;
   */
  EQ = 1,

  /**
   * @generated from enum value: OPERATION_NEQ = 2;
   */
  NEQ = 2,

  /**
   * @generated from enum value: OPERATION_LT = 3;
   */
  LT = 3,

  /**
   * @generated from enum value: OPERATION_LTE = 4;
   */
  LTE = 4,

  /**
   * @generated from enum value: OPERATION_GT = 5;
   */
  GT = 5,

  /**
   * @generated from enum value: OPERATION_GTE = 6;
   */
  GTE = 6,

  /**
   * @generated from enum value: OPERATION_OR = 7;
   */
  OR = 7,

  /**
   * @generated from enum value: OPERATION_AND = 8;
   */
  AND = 8,

  /**
   * @generated from enum value: OPERATION_IN = 9;
   */
  IN = 9,

  /**
   * @generated from enum value: OPERATION_NIN = 10;
   */
  NIN = 10,

  /**
   * @generated from enum value: OPERATION_LIKE = 11;
   */
  LIKE = 11,

  /**
   * @generated from enum value: OPERATION_NLIKE = 12;
   */
  NLIKE = 12,
}

/**
 * Describes the enum rill.runtime.v1.Operation.
 */
export const OperationSchema: GenEnum<Operation> = /*@__PURE__*/
  enumDesc(file_rill_runtime_v1_expression, 0);

