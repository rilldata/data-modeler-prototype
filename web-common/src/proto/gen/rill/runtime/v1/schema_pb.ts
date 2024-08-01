// @generated by protoc-gen-es v2.0.0 with parameter "target=ts"
// @generated from file rill/runtime/v1/schema.proto (package rill.runtime.v1, syntax proto3)
/* eslint-disable */

import type { GenEnum, GenFile, GenMessage } from "@bufbuild/protobuf/codegenv1";
import { enumDesc, fileDesc, messageDesc } from "@bufbuild/protobuf/codegenv1";
import type { Message } from "@bufbuild/protobuf";

/**
 * Describes the file rill/runtime/v1/schema.proto.
 */
export const file_rill_runtime_v1_schema: GenFile = /*@__PURE__*/
  fileDesc("ChxyaWxsL3J1bnRpbWUvdjEvc2NoZW1hLnByb3RvEg9yaWxsLnJ1bnRpbWUudjEinwUKBFR5cGUSKAoEY29kZRgBIAEoDjIaLnJpbGwucnVudGltZS52MS5UeXBlLkNvZGUSEAoIbnVsbGFibGUYAiABKAgSMQoSYXJyYXlfZWxlbWVudF90eXBlGAMgASgLMhUucmlsbC5ydW50aW1lLnYxLlR5cGUSMAoLc3RydWN0X3R5cGUYBCABKAsyGy5yaWxsLnJ1bnRpbWUudjEuU3RydWN0VHlwZRIqCghtYXBfdHlwZRgFIAEoCzIYLnJpbGwucnVudGltZS52MS5NYXBUeXBlIskDCgRDb2RlEhQKEENPREVfVU5TUEVDSUZJRUQQABINCglDT0RFX0JPT0wQARINCglDT0RFX0lOVDgQAhIOCgpDT0RFX0lOVDE2EAMSDgoKQ09ERV9JTlQzMhAEEg4KCkNPREVfSU5UNjQQBRIPCgtDT0RFX0lOVDEyOBAGEg8KC0NPREVfSU5UMjU2EBkSDgoKQ09ERV9VSU5UOBAHEg8KC0NPREVfVUlOVDE2EAgSDwoLQ09ERV9VSU5UMzIQCRIPCgtDT0RFX1VJTlQ2NBAKEhAKDENPREVfVUlOVDEyOBALEhAKDENPREVfVUlOVDI1NhAaEhAKDENPREVfRkxPQVQzMhAMEhAKDENPREVfRkxPQVQ2NBANEhIKDkNPREVfVElNRVNUQU1QEA4SDQoJQ09ERV9EQVRFEA8SDQoJQ09ERV9USU1FEBASDwoLQ09ERV9TVFJJTkcQERIOCgpDT0RFX0JZVEVTEBISDgoKQ09ERV9BUlJBWRATEg8KC0NPREVfU1RSVUNUEBQSDAoIQ09ERV9NQVAQFRIQCgxDT0RFX0RFQ0lNQUwQFhINCglDT0RFX0pTT04QFxINCglDT0RFX1VVSUQQGCJ7CgpTdHJ1Y3RUeXBlEjEKBmZpZWxkcxgBIAMoCzIhLnJpbGwucnVudGltZS52MS5TdHJ1Y3RUeXBlLkZpZWxkGjoKBUZpZWxkEgwKBG5hbWUYASABKAkSIwoEdHlwZRgCIAEoCzIVLnJpbGwucnVudGltZS52MS5UeXBlIl0KB01hcFR5cGUSJwoIa2V5X3R5cGUYASABKAsyFS5yaWxsLnJ1bnRpbWUudjEuVHlwZRIpCgp2YWx1ZV90eXBlGAIgASgLMhUucmlsbC5ydW50aW1lLnYxLlR5cGVCtAEKE2NvbS5yaWxsLnJ1bnRpbWUudjFCC1NjaGVtYVByb3RvUAFaMmdpdGh1Yi5jb20vcmlsbGRhdGEvcmlsbC9yaWxsL3J1bnRpbWUvdjE7cnVudGltZXYxogIDUlJYqgIPUmlsbC5SdW50aW1lLlYxygIPUmlsbFxSdW50aW1lXFYx4gIbUmlsbFxSdW50aW1lXFYxXEdQQk1ldGFkYXRh6gIRUmlsbDo6UnVudGltZTo6VjFiBnByb3RvMw");

/**
 * Type represents a data type in a schema
 *
 * @generated from message rill.runtime.v1.Type
 */
export type Type = Message<"rill.runtime.v1.Type"> & {
  /**
   * Code designates the type
   *
   * @generated from field: rill.runtime.v1.Type.Code code = 1;
   */
  code: Type_Code;

  /**
   * Nullable indicates whether null values are possible
   *
   * @generated from field: bool nullable = 2;
   */
  nullable: boolean;

  /**
   * If code is CODE_ARRAY, array_element_type specifies the type of the array elements
   *
   * @generated from field: rill.runtime.v1.Type array_element_type = 3;
   */
  arrayElementType?: Type;

  /**
   * If code is CODE_STRUCT, struct_type specifies the type of the struct's fields
   *
   * @generated from field: rill.runtime.v1.StructType struct_type = 4;
   */
  structType?: StructType;

  /**
   * If code is CODE_MAP, map_type specifies the map's key and value types
   *
   * @generated from field: rill.runtime.v1.MapType map_type = 5;
   */
  mapType?: MapType;
};

/**
 * Describes the message rill.runtime.v1.Type.
 * Use `create(TypeSchema)` to create a new message.
 */
export const TypeSchema: GenMessage<Type> = /*@__PURE__*/
  messageDesc(file_rill_runtime_v1_schema, 0);

/**
 * Code enumerates all the types that can be represented in a schema
 *
 * @generated from enum rill.runtime.v1.Type.Code
 */
export enum Type_Code {
  /**
   * @generated from enum value: CODE_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: CODE_BOOL = 1;
   */
  BOOL = 1,

  /**
   * @generated from enum value: CODE_INT8 = 2;
   */
  INT8 = 2,

  /**
   * @generated from enum value: CODE_INT16 = 3;
   */
  INT16 = 3,

  /**
   * @generated from enum value: CODE_INT32 = 4;
   */
  INT32 = 4,

  /**
   * @generated from enum value: CODE_INT64 = 5;
   */
  INT64 = 5,

  /**
   * @generated from enum value: CODE_INT128 = 6;
   */
  INT128 = 6,

  /**
   * @generated from enum value: CODE_INT256 = 25;
   */
  INT256 = 25,

  /**
   * @generated from enum value: CODE_UINT8 = 7;
   */
  UINT8 = 7,

  /**
   * @generated from enum value: CODE_UINT16 = 8;
   */
  UINT16 = 8,

  /**
   * @generated from enum value: CODE_UINT32 = 9;
   */
  UINT32 = 9,

  /**
   * @generated from enum value: CODE_UINT64 = 10;
   */
  UINT64 = 10,

  /**
   * @generated from enum value: CODE_UINT128 = 11;
   */
  UINT128 = 11,

  /**
   * @generated from enum value: CODE_UINT256 = 26;
   */
  UINT256 = 26,

  /**
   * @generated from enum value: CODE_FLOAT32 = 12;
   */
  FLOAT32 = 12,

  /**
   * @generated from enum value: CODE_FLOAT64 = 13;
   */
  FLOAT64 = 13,

  /**
   * @generated from enum value: CODE_TIMESTAMP = 14;
   */
  TIMESTAMP = 14,

  /**
   * @generated from enum value: CODE_DATE = 15;
   */
  DATE = 15,

  /**
   * @generated from enum value: CODE_TIME = 16;
   */
  TIME = 16,

  /**
   * @generated from enum value: CODE_STRING = 17;
   */
  STRING = 17,

  /**
   * @generated from enum value: CODE_BYTES = 18;
   */
  BYTES = 18,

  /**
   * @generated from enum value: CODE_ARRAY = 19;
   */
  ARRAY = 19,

  /**
   * @generated from enum value: CODE_STRUCT = 20;
   */
  STRUCT = 20,

  /**
   * @generated from enum value: CODE_MAP = 21;
   */
  MAP = 21,

  /**
   * @generated from enum value: CODE_DECIMAL = 22;
   */
  DECIMAL = 22,

  /**
   * @generated from enum value: CODE_JSON = 23;
   */
  JSON = 23,

  /**
   * @generated from enum value: CODE_UUID = 24;
   */
  UUID = 24,
}

/**
 * Describes the enum rill.runtime.v1.Type.Code.
 */
export const Type_CodeSchema: GenEnum<Type_Code> = /*@__PURE__*/
  enumDesc(file_rill_runtime_v1_schema, 0, 0);

/**
 * StructType is a type composed of ordered, named and typed sub-fields
 *
 * @generated from message rill.runtime.v1.StructType
 */
export type StructType = Message<"rill.runtime.v1.StructType"> & {
  /**
   * @generated from field: repeated rill.runtime.v1.StructType.Field fields = 1;
   */
  fields: StructType_Field[];
};

/**
 * Describes the message rill.runtime.v1.StructType.
 * Use `create(StructTypeSchema)` to create a new message.
 */
export const StructTypeSchema: GenMessage<StructType> = /*@__PURE__*/
  messageDesc(file_rill_runtime_v1_schema, 1);

/**
 * @generated from message rill.runtime.v1.StructType.Field
 */
export type StructType_Field = Message<"rill.runtime.v1.StructType.Field"> & {
  /**
   * @generated from field: string name = 1;
   */
  name: string;

  /**
   * @generated from field: rill.runtime.v1.Type type = 2;
   */
  type?: Type;
};

/**
 * Describes the message rill.runtime.v1.StructType.Field.
 * Use `create(StructType_FieldSchema)` to create a new message.
 */
export const StructType_FieldSchema: GenMessage<StructType_Field> = /*@__PURE__*/
  messageDesc(file_rill_runtime_v1_schema, 1, 0);

/**
 * MapType is a complex type for mapping keys to values
 *
 * @generated from message rill.runtime.v1.MapType
 */
export type MapType = Message<"rill.runtime.v1.MapType"> & {
  /**
   * @generated from field: rill.runtime.v1.Type key_type = 1;
   */
  keyType?: Type;

  /**
   * @generated from field: rill.runtime.v1.Type value_type = 2;
   */
  valueType?: Type;
};

/**
 * Describes the message rill.runtime.v1.MapType.
 * Use `create(MapTypeSchema)` to create a new message.
 */
export const MapTypeSchema: GenMessage<MapType> = /*@__PURE__*/
  messageDesc(file_rill_runtime_v1_schema, 2);

