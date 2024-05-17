---
title: Dashboard YAML
sidebar_label: Dashboard YAML
sidebar_position: 30
hide_table_of_contents: true
---

In your Rill project directory, create a `<dashboard_name>.yaml` file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

:::tip Did you know?

Files that are *nested at any level* under your native `dashboards` directory will be assumed to be metric definitions, i.e. dashboards (unless **otherwise** specified by the `type` property).

:::

## Properties

**`type`** - Refers to the resource type and must be `metrics_view` _(required)_. 

**`model`** — Refers to the **model** powering the dashboard with no path specified; should only be used for [Rill models](/build/models/models.md) _(either **model** or **table** is required)_.

**`table`** - Refers to the **table** powering the dashboard with no path specified; should be used instead of `model` for dashboards created directly from [sources](/build/connect/connect.md) and/or from [external OLAP tables](build/olap/olap.md#external-olap-tables) _(either **table** or **model** is required)_.

**`title`** — Refers to the display name for the dashboard _(required)_.

**`timeseries`** — Refers to the timestamp column from your model that will underlie x-axis data in the line charts. If not specified, the line charts will not appear _(optional)_.

**`default_time_range`** — Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) (default). If not specified, defaults to the full time range of the `timeseries` column _(optional)_.

**`smallest_time_grain`** — Refers to the smallest time granularity the user is allowed to view in the dashboard. The valid values are: `millisecond`, `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year` _(optional)_.

**`first_day_of_week`** — Refers to the first day of the week for time grain aggregation (for example, Sunday instead of Monday). The valid values are 1 through 7 where Monday=`1` and Sunday=`7` _(optional)_.

**`first_month_of_year`** — Refers to the first month of the year for time grain aggregation. The valid values are 1 through 12 where January=`1` and December=`12` _(optional)_.

**`available_time_zones`** — Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). By adding one or more time zones will make the dashboard time zone aware and allow users to change current time zone within the dashboard _(optional)_.

**`default_theme`** — Refers to the default theme to apply to the dashboard. A valid theme must be defined in the project. Read this [page](./themes.md) for more detailed information about themes _(optional)_.

**`default_comparison`** - Defines which should be the default comparison mode. Default to `none` _(optional)_.
  - **`mode`** - comparison mode
    - `none` - no comparison
    - `time` - time, will pick the comparison period depending on `default_time_range`
    - `dimension` - dimension comparison mode
  - **`dimension`** - for dimension mode, specify the comparison dimension by name

**`dimensions`** — Relates to exploring segments or [dimensions](/build/dashboards/dashboards.md#dimensions) of your data and filtering the dashboard _(required)_.
  - **`column`** — a categorical column _(required)_ 
  - **`expression`** a non-aggregate expression such as `string_split(domain, '.')`. One of `column` and `expression` is required but cannot have both at the same time _(required)_
  - **`name`** — a stable identifier for the dimension _(optional)_
  - **`label`** — a label for your dashboard dimension _(optional)_ 
  - **`description`** — a freeform text description of the dimension for your dashboard _(optional)_
  - **`unnest`** - if true, allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match _(optional)_
  - **`ignore`** — hides the dimension _(optional)_ 

**`measures`** — Used to define the numeric [aggregates](/build/dashboards/dashboards.md#measures) of columns from your data model  _(required)_.
  - **`expression`** — a combination of operators and functions for aggregations _(required)_ 
  - **`name`** — a stable identifier for the measure _(required)_
  - **`label`** — a label for your dashboard measure _(optional)_ 
  - **`description`** — a freeform text description of the dimension for your dashboard _(optional)_ 
  - **`ignore`** — hides the measure _(optional)_ 
  - **`valid_percent_of_total`** — a boolean indicating whether percent-of-total values should be rendered for this measure _(optional)_ 
  - **`format_d3`** — controls the formatting of this measure in the dashboard using a [d3-format string](https://d3js.org/d3-format). If an invalid format string is supplied, measures will be formatted with `format_preset: humanize` (described below). Measures <u>cannot</u> have both `format_preset` and `format_d3` entries. _(optional; if neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset)_
    - **Example**: to show a measure using fixed point formatting with 2 digits after the decimal point, your measure specification would include: `format_d3: ".2f"`.
    - **Example**: to show a measure using grouped thousands with two significant digits, your measure specification would include: `format_d3: ",.2r"`.
  - **`format_preset`** — controls the formatting of this measure in the dashboard according to option specified below. Measures <u>cannot</u> have both `format_preset` and `format_d3` entries. _(optional; if neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset)_
    - `humanize` — round off numbers in an opinionated way to thousands (K), millions (M), billions (B), etc.
    - `none` — raw output
    - `currency_usd` —  output rounded to 2 decimal points prepended with a dollar sign: `$`
    - `currency_eur` —  output rounded to 2 decimal points prepended with a euro symbol: `€`
    - `percentage` — output transformed from a rate to a percentage appended with a percentage sign
    - `interval_ms` — time intervals given in milliseconds are transformed into human readable time units like hours (h), days (d), years (y), etc.

**`available_time_ranges`** — Overrides the list of default time range selections available in the dropdown. Note that `All Time` and `Custom` selections are always available _(optional)_.
  - **`range`** — a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) for the selection _(required)_
  - **`comparison_offsets`** — list of time comparison options for this time range selection _(optional)_. Must be one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions).
  - **Example**:
    ```yaml
    available_time_ranges:
    - PT15M // Simplified syntax to specify only the range
    - PT1H
    - PT6H
    - P7D
    - range: P5D // Advanced syntax to specify comparison_offsets as well
      comparison_offsets:
        - rill-PP
        - rill-PW
    - P4W
    - rill-TD // Today
    - rill-WTD // Week-To-date
    ```

**`default_dimensions`** - Provides the list of dimensions that should be visible on the dashboard upon initial render by default _(optional)_.

**`default_measures`** - Provides the list of measures that should be visible on the dashboard upon initial render by default _(optional)_.

**`security`** - Defines a [security policy](/manage/security) for the dashboard _(optional)_.
  - **`access`** - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to `false` and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean _(optional)_.
  - **`row_filter`** - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a `WHERE` clause _(optional)_.
  - **`exclude`** - List of dimension or measure names to exclude from the dashboard. If `exclude` is defined all other dimensions and measures are included _(optional)_.
    - **`if`** - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_.
    - **`names`** - List of fields to exclude. Should match the `name` of one of the dashboard's dimensions or measures _(required)_.
  - **`include`** - List of dimension or measure names to include in the dashboard. If `include` is defined all other dimensions and measures are excluded _(optional)_.
    - **`if`** - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_.
    - **`names`** - List of fields to include. Should match the `name` of one of the dashboard's dimensions or measures _(required)_.
