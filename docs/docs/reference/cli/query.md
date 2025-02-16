---
note: GENERATED. DO NOT EDIT.
title: rill query
---
## rill query

Query a resolver within a project

### Synopsis

Query a resolver within a project.

You can query a resolver by providing a SQL query, a resolver name, or a connector name.

Example Usage:

Query a resolver by providing a SQL query:
rill query my-project --sql "SELECT * FROM my-table"
rill query --sql "SELECT * FROM my-table" --limit 10


```
rill query [<project>] [flags]
```

### Flags

```
      --args stringToString         Explicit resolver args (only with --resolver) (default [])
      --connector string            Connector to execute against. Defaults to the OLAP connector.
      --limit string                The maximum number of rows to print (default: 100) (default "100")
      --local                       Target local runtime instead of Rill Cloud
      --path string                 Project directory (default ".")
      --project string              Project name
      --properties stringToString   Explicit resolver properties (only with --resolver) (default [])
      --resolver string             Explicit resolver (cannot be combined with --sql)
      --sql string                  A SELECT query to execute
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill](cli.md)	 - A CLI for Rill

