---
title: rill project reconcile
---
## rill project reconcile

Send trigger to deployment

```
rill project reconcile [<project-name>] [flags]
```

### Flags

```
      --project string           Project name
      --path string              Project directory (default ".")
      --refresh                  Refresh all sources
      --refresh-source strings   Refresh specific source(s)
      --reset                    Reset and redeploy the project from scratch
```

### Global flags

```
      --api-token string   Token for authenticating with the admin API
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
      --org string         Organization Name
```

### SEE ALSO

* [rill project](project.md)	 - Manage projects

