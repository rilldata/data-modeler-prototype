# web-cloud

This folder contains the cloud frontend implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). 

## Running in development

1. Run the Go backend in `server-cloud` (see its `README` for instructions)
2. Run `npm install`
3. Run `npm run dev`

There's currently no UI to create orgs or projects. You can add some directly using `curl`:
```
# Add an organization
curl -X POST http://localhost:8080/v1/organizations -H 'Content-Type: application/json' -d '{"name":"foo", "description":"org foo"}'

# Add a project
curl -X POST http://localhost:8080/v1/organizations/foo/projects -H 'Content-Type: application/json' -d '{"name":"bar", "description":"project bar"}'
```

## Generating the client

We use [Orval](https://orval.dev) to generate a client for interacting with cloud backend server (in `server-cloud`). The client is generated in `web-cloud/src/client/gen/` and based on the OpenAPI schema in `server-cloud/api/openapi.yaml`. Orval is configured to generate a client that uses [@sveltestack/svelte-query](https://sveltequery.vercel.app).

You have to manually re-generate the client when the OpenAPI spec changes. We could automate this step, but for now, we're going to avoid cross-language magic.

To re-generate the client, run:

```script
npm run generate:client -w web-cloud
```
