export enum AuthStep {
  Base = 0,
  // SSO = 1,
  Login = 2,
  SignUp = 3,
  // Thanks = 4,
  Loading = 5,
}

type InternalOptions = {
  protocol: string;
  response_type: string;
  prompt: string;
  scope: string;
  _csrf: string;
  leeway: number;
};

export type Config = {
  auth0Domain: string;
  clientID: string;
  auth0Tenant: string;
  authorizationServer: {
    issuer: string;
  };
  callbackURL: string;
  internalOptions: InternalOptions;
  extraParams?: { screen_hint?: string };
};
