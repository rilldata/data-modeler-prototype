---
title: rill deploy
---
## rill deploy

Deploy project to Rill Cloud

```
rill deploy [flags]
```

### Flags

```
      --path string             Path to project repository (default ".")
      --org string              Org to deploy project
      --description string      Project description
      --region string           Deployment region
      --prod-db-driver string   Database driver (default "duckdb")
      --prod-db-dsn string      Database driver configuration
      --public                  Make dashboards publicly accessible
      --subpath string          Relative path to project in the repository (for monorepos)
      --prod-branch string      Git branch to deploy from (default: the default Git branch)
      --project string          Project name (default: Git repo name)
      --remote string           Remote name (defaults: first github remote)
      --api-token string        Token for authenticating with the admin API
```

### Global flags

```
  -h, --help          Print usage
      --interactive   Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill](cli.md)	 - Rill CLI

