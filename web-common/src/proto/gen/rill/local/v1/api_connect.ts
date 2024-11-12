// @generated by protoc-gen-connect-es v1.4.0 with parameter "target=ts"
// @generated from file rill/local/v1/api.proto (package rill.local.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { DeployProjectRequest, DeployProjectResponse, GetCurrentProjectRequest, GetCurrentProjectResponse, GetCurrentUserRequest, GetCurrentUserResponse, GetMetadataRequest, GetMetadataResponse, GetUserOrgMetadataRequest, GetUserOrgMetadataResponse, GetVersionRequest, GetVersionResponse, PingRequest, PingResponse, PushToGithubRequest, PushToGithubResponse, RedeployProjectRequest, RedeployProjectResponse } from "./api_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service rill.local.v1.LocalService
 */
export const LocalService = {
  typeName: "rill.local.v1.LocalService",
  methods: {
    /**
     * Ping returns the current time.
     *
     * @generated from rpc rill.local.v1.LocalService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
    },
    /**
     * GetMetadata returns information about the local Rill instance.
     *
     * @generated from rpc rill.local.v1.LocalService.GetMetadata
     */
    getMetadata: {
      name: "GetMetadata",
      I: GetMetadataRequest,
      O: GetMetadataResponse,
      kind: MethodKind.Unary,
    },
    /**
     * GetVersion returns details about the current and latest available Rill versions.
     *
     * @generated from rpc rill.local.v1.LocalService.GetVersion
     */
    getVersion: {
      name: "GetVersion",
      I: GetVersionRequest,
      O: GetVersionResponse,
      kind: MethodKind.Unary,
    },
    /**
     * PushToGithub create a Git repo from local project and pushed to users git account.
     *
     * @generated from rpc rill.local.v1.LocalService.PushToGithub
     */
    pushToGithub: {
      name: "PushToGithub",
      I: PushToGithubRequest,
      O: PushToGithubResponse,
      kind: MethodKind.Unary,
    },
    /**
     * DeployProject deploys the local project to the Rill cloud.
     *
     * @generated from rpc rill.local.v1.LocalService.DeployProject
     */
    deployProject: {
      name: "DeployProject",
      I: DeployProjectRequest,
      O: DeployProjectResponse,
      kind: MethodKind.Unary,
    },
    /**
     * RedeployProject updates a deployed project.
     *
     * @generated from rpc rill.local.v1.LocalService.RedeployProject
     */
    redeployProject: {
      name: "RedeployProject",
      I: RedeployProjectRequest,
      O: RedeployProjectResponse,
      kind: MethodKind.Unary,
    },
    /**
     * GetCurrentUser returns the locally logged in user
     *
     * @generated from rpc rill.local.v1.LocalService.GetCurrentUser
     */
    getCurrentUser: {
      name: "GetCurrentUser",
      I: GetCurrentUserRequest,
      O: GetCurrentUserResponse,
      kind: MethodKind.Unary,
    },
    /**
     * GetCurrentProject returns the rill cloud project connected to the local project
     *
     * @generated from rpc rill.local.v1.LocalService.GetCurrentProject
     */
    getCurrentProject: {
      name: "GetCurrentProject",
      I: GetCurrentProjectRequest,
      O: GetCurrentProjectResponse,
      kind: MethodKind.Unary,
    },
    /**
     * GetUserOrgMetadata returns metadata about the current user's orgs.
     *
     * @generated from rpc rill.local.v1.LocalService.GetUserOrgMetadata
     */
    getUserOrgMetadata: {
      name: "GetUserOrgMetadata",
      I: GetUserOrgMetadataRequest,
      O: GetUserOrgMetadataResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

