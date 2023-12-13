// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file rill/runtime/v1/expression.proto (package rill.runtime.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Value } from "@bufbuild/protobuf";

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
// Retrieve enum metadata with: proto3.getEnumType(Operation)
proto3.util.setEnumType(Operation, "rill.runtime.v1.Operation", [
  { no: 0, name: "OPERATION_UNSPECIFIED" },
  { no: 1, name: "OPERATION_EQ" },
  { no: 2, name: "OPERATION_NEQ" },
  { no: 3, name: "OPERATION_LT" },
  { no: 4, name: "OPERATION_LTE" },
  { no: 5, name: "OPERATION_GT" },
  { no: 6, name: "OPERATION_GTE" },
  { no: 7, name: "OPERATION_OR" },
  { no: 8, name: "OPERATION_AND" },
  { no: 9, name: "OPERATION_IN" },
  { no: 10, name: "OPERATION_NIN" },
  { no: 11, name: "OPERATION_LIKE" },
  { no: 12, name: "OPERATION_NLIKE" },
]);

/**
 * @generated from message rill.runtime.v1.Expression
 */
export class Expression extends Message<Expression> {
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
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Expression>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "rill.runtime.v1.Expression";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ident", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "expression" },
    { no: 2, name: "val", kind: "message", T: Value, oneof: "expression" },
    { no: 3, name: "cond", kind: "message", T: Condition, oneof: "expression" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Expression {
    return new Expression().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Expression {
    return new Expression().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Expression {
    return new Expression().fromJsonString(jsonString, options);
  }

  static equals(a: Expression | PlainMessage<Expression> | undefined, b: Expression | PlainMessage<Expression> | undefined): boolean {
    return proto3.util.equals(Expression, a, b);
  }
}

/**
 * @generated from message rill.runtime.v1.Condition
 */
export class Condition extends Message<Condition> {
  /**
   * @generated from field: rill.runtime.v1.Operation op = 1;
   */
  op = Operation.UNSPECIFIED;

  /**
   * @generated from field: repeated rill.runtime.v1.Expression exprs = 2;
   */
  exprs: Expression[] = [];

  constructor(data?: PartialMessage<Condition>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "rill.runtime.v1.Condition";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "op", kind: "enum", T: proto3.getEnumType(Operation) },
    { no: 2, name: "exprs", kind: "message", T: Expression, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Condition {
    return new Condition().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Condition {
    return new Condition().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Condition {
    return new Condition().fromJsonString(jsonString, options);
  }

  static equals(a: Condition | PlainMessage<Condition> | undefined, b: Condition | PlainMessage<Condition> | undefined): boolean {
    return proto3.util.equals(Condition, a, b);
  }
}

