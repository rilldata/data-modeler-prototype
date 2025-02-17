---
title: Explore Dashboard YAML
sidebar_label: Explore Dashboard YAML
sidebar_position: 30
hide_table_of_contents: true
---

In your Rill project directory, create a explore dashboard, `<dashboard_name>.yaml`, file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties

**`type`** — Refers to the resource type and must be `explore` _(required)_. 

**`metrics_view`** — Refers to the metrics view resource _(required)_. 

**`title`** — Refers to the display name for the dashboard [deprecated, use `display_name`] _(required)_.

**`display_name`** - Refers to the display name for the metrics view _(required)_.

**`banner`** - Refers to the custom banner displayed at the header of an explore dashboard  _(optional)_.

**`description`** - A description for the project _(optional)_.

**`dimensions`** - List of dimension names. Use `'*'` to select all dimensions (default) _(optional)_. 
  - **`regex`** - Select dimensions using a regular expression _(optional)_.
  - **`exclude`** - Select all dimensions *except* those listed here _(optional)_.

```yaml
# Example: Select a dimension
dimensions:
  - country

# Example: Select all dimensions except one
dimensions:
  exclude:
    - country

# Example: Select all dimensions that match a regex
dimensions:
  regex: "^public_.*$"
```

**`measures`** - List of measure names. Use `'*'` to select all measures (default) _(optional)_. 
  - **`regex`** - Select measures using a regular expression (see `dimensions` above for an example) _(optional)_.
  - **`exclude`** - Select all measures *except* those listed here (see `dimensions` above for an example) _(optional)_.

**`defaults`** - defines the defaults YAML struct

    - **`time_range`** — Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) (default). If not specified, defaults to the full time range of the `timeseries` column _(optional)_.


    - **`comparison_mode`** - comparison mode _(optional)_.
      - `none` - no comparison
      - `time` - time, will pick the comparison period depending on `default_time_range`
      - `dimension` - dimension comparison mode

    - **`comparison_dimension`** - for dimension mode, specify the comparison dimension by name _(optional)_.

    - **`dimensions`** - Provides the default dimensions to load on viewing the dashboard. _(optional)_.

    - **`measures`** -  Provides the default measures to load on viewing the dashboard. _(optional)_.

    **Default Example:**
    ```yaml
    defaults: #define all the defaults within here
        dimensions:
          - dim_1
          - dim_2
        measures:
          - measure_1
          - measure_2
        time_range: P1M
        comparison_mode: dimension #time, none
        comparison_dimension: filename
```

**`time_ranges`** — Overrides the list of default time range selections available in the dropdown. Note that `All Time` and `Custom` selections are always available _(optional)_.
  - **`range`** — a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) for the selection _(required)_
  - **`comparison_offsets`** — list of time comparison options for this time range selection _(optional)_. Must be one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions).
  
  **Example**:
    ```yaml
    time_ranges:
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

**`time_zones`** — Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). By adding one or more time zones will make the dashboard time zone aware and allow users to change current time zone within the dashboard _(optional)_.

**`theme`** — Refers to the default theme to apply to the dashboard. A valid theme must be defined in the project. Read this [page](./themes.md) for more detailed information about themes _(optional)_.
```yaml
theme:
  colors:
    primary: hsl(180, 100%, 50%)
    secondary: lightgreen
```

**`security`** - Defines a [security policy](/manage/security) for the dashboard _(optional)_.
  - **`access`** - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to `false` and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean _(optional)_.