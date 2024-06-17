---
title: Project YAML
sidebar_label: Project YAML
sidebar_position: 40
hide_table_of_contents: true
---

The `rill.yaml` file contains metadata about your project.

## Properties

**`title`** — the name of your project which will be displayed in the upper left hand corner
**`compiler`** — the Rill project compiler version compatible with your project files (currently defaults to: `rillv1`)
**`olap_connector`** - the default OLAP engine to use in your project
**`mock_users`** — a list of mock users to test against dashboard [security policies](/manage/security). For each mock user, possible attributes include:
  - **`email`** — the mock user's email _(required)_
  - **`name`** — the mock user's name
  - **`admin`** — whether or not the mock user is an admin

## Configuring the default OLAP Engine

Rill allows you to specify the default OLAP engine to use in your project via `rill.yaml`. This setting is configurable using the `olap_connector` property (and will otherwise revert to `duckdb` if not specified). 

:::info Curious about OLAP Engines?

Please see our reference documentation on [OLAP Engines](../olap-engines/olap-engines.md).

:::
 
## Project-wide defaults

In `rill.yaml`, project-wide defaults can be specified for a resource type within a project. Unless otherwise specified, _individual resources will inherit any defaults_ that have been specified in `rill.yaml`. For available properties that can be configured, please refer to the YAML specification for each individual resource type - [sources](sources.md), [models](models.md), and [dashboards](dashboards.md)

:::note Use plurals when specifying project-wide defaults

In your `rill.yaml`, the top level property for the resource type needs to be **plural**, such as `sources`, `models`, and `dashboards`.

:::

For example, the following YAML configuration below will set a project-wide default for:
- **Sources** - Configure a [source refresh](/build/connect/source-refresh.md).
- **Models** - Automatically materialize the models as tables instead of views (the default behavior if unspecified).
- **Dashboards** - Set the [first day of the week](dashboards.md) for timeseries aggregations to be Sunday along with defining available timezones on a dashboard.

```yaml
title: My Rill Project
sources:
  refresh:
    cron: '0 * * * *'
models:
  materialize: true
dashboards:
  first_day_of_week: 7
  available_time_zones:
    - America/Los_Angeles
    - America/New_York
    - Europe/London
    - Asia/Kolkata
```

:::info Hierarchy of inheritance and property overrides

As a general rule of thumb, properties that have been specified at a more _granular_ level will supercede or override higher level properties that have been inherited. Therefore, in order of inheritance, Rill will prioritize properties in the following order:
1. Individual [source](/reference/project-files/sources.md)/[model](/reference/project-files/models.md)/[dashboard](/reference/project-files/dashboards.md) object level properties (e.g. `source.yaml` or `dashboard.yaml`)
2. [Environment](/docs/build/models/environments.md) level properties (e.g. a specific property that have been set for `dev`)
3. [Project-wide defaults](/reference/project-files/rill-yaml.md#project-wide-defaults) for a specific property and resource type

:::

## Setting variables

Primarily useful for [templating](/deploy/templating.md), variables can be set in the `rill.yaml` file directly. This allows variables to be set for your projects deployed to Rill Cloud while still being able to use different variable values locally if you prefer. 

To define a variable in `rill.yaml`, pass in the appropriate key-value pair for the variable under the `vars` key:
```yaml
vars:
  numeric_var: 10
  string_var: "string_value"
```

:::info Overriding variables locally

Variables also follow an order of precedence and can be overriden locally. By default, any variables defined will be inherited from `rill.yaml`. However, if you manually pass in a variable when starting Rill Developer locally via the CLI, this value will be used instead for the current instance of your running project:

```bash
rill start --var numeric_var=100 --var string_var="different_value"
```

:::

:::tip Setting variables through `.env`

Variables can also be set through your project's `<RILL_PROJECT_HOME>/.env` file (or using the `rill env set` CLI command), such as:
```bash
variable=xyz
```

Similar to how [connector credentials can be pushed / pulled](/build/credentials/credentials.md#pushing-and-pulling-credentials-to--from-rill-cloud) from local to cloud or vice versa, project variables set locally in Rill Developer can be pushed to Rill Cloud and/or pulled back to your local instance from your deployed project by using the `rill env push` and `rill env pull` commands respectively.

:::