import { type UserFacingError } from "./error-store";

export function createUserFacingError(
  status: number | null,
  message: string,
): UserFacingError {
  // Handle network errors
  if (message === "Network Error") {
    return {
      statusCode: null,
      header: "Network Error",
      body: "It seems we're having trouble reaching our servers. Check your connection or try again later.",
    };
  }

  // Handle some application errors
  if (status === 400 && message === "driver: not found") {
    return {
      statusCode: status,
      header: "Project deployment not found",
      body: "This is potentially a temporary state if the project has just been reset.",
    };
  } else if (status === 401 && message === "auth token is expired") {
    return {
      statusCode: 401,
      header: "Oops! This link has expired",
      body: "It looks like this link is no longer active. Please reach out to the sender to request a new link.",
      fatal: true,
    };
  } else if (status === 403) {
    return {
      statusCode: status,
      header: "Access denied",
      body: "You don't have access to this page. Please check that you have the correct permissions.",
    };
  } else if (message === "org not found") {
    return {
      statusCode: status,
      header: "Organization not found",
      body: "The organization you requested could not be found. Please check that you have provided a valid organization name.",
    };
  } else if (message === "project not found") {
    return {
      statusCode: status,
      header: "Project not found",
      body: "The project you requested could not be found. Please check that you have provided a valid project name.",
    };
  } else if (status === 404 && message === "resource not found") {
    return {
      statusCode: 404,
      header: "Resource not found",
      body: "This resource may have been deleted, renamed, or is temporarily unavailable.",
    };
  } else if (status === 404) {
    return {
      statusCode: 404,
      header: "Sorry, we can't find this page!",
      body: "The page you're looking for might have been removed, had its name changed, or is temporarily unavailable.",
    };
  }

  // Fallback for all other errors (including 5xx errors)
  return {
    statusCode: status,
    header: "Sorry, something went wrong!",
    body: "Try refreshing the page, and reach out to us if the problem persists.",
    detail: message,
  };
}
