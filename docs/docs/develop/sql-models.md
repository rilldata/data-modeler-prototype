---
title: "Model: Transform Data"
description: Transform your data using Duck SQL
sidebar_label: "Model: Transform Data"
sidebar_position: 20
---

Data models in Rill are composed of SQL `SELECT` statements that operate on source data. They allow you to join, transform, and clean data.

## SQL transformations

Data transformations in Rill Developer are powered by DuckDB and their dialect of SQL (DuckDB SQL). Please visit [DuckDB SQL documentation](https://duckdb.org/docs/sql/introduction) to learn about how to write your queries.

## Adding a data model

### Using the UI
To add a new data model using the UI, click "+" by Models in the left hand navigation pane. You can now begin typing a DuckDB SQL `SELECT` query for your model in the code editor – with keystroke-by-keystroke feedback.

### Using code
When you add a data model using the UI, a code definition will automatically be created as a `<model_name>.sql` file in the `models` folder in your Rill project.

In addition, you can create a model outside of the application and add it to Rill by placing a `<model_name>.sql` file in the `models` directory containing a DuckDB SQL `SELECT` statement. Rill will automatically detect and parse the model next time you run `rill start`.

## How to use data models

### One Big Table and dashboarding

It is powerful to be able to translate many ad hoc questions into a data framework that can answer a class of questions at scale. For example, high-level insights that are relevant to a company (how much revenue did we make last week?) are more actionable for an employee if it is relevant for their role (how did my campaign increase revenue last week?).

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable [metrics definition](./metrics-dashboard) to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.

### Intermediate processing

Models can also be cross-referenced between each other to produce the final output for your dashboard. The advantage here is more complex, intermediate data transformations can be utilized to achieve your final source for the dashboard. Example ideas for modeling:

- Lookups for id/name joins
- Unnesting and merging complex data types
- Combining multiple sources with data cleansing or transformation requirements