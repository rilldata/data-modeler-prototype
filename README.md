# Rill Developer **_(tech preview)_**
Rill Developer is a tool that makes it effortless to transform your datasets with SQL and create powerful, opinionated dashboards. Rill Developer follows a few guiding principles:

- *feels good to use* – powered by Sveltekit & DuckDB = conversation-fast, not wait-ten-seconds-for-result-set fast
- *works with your local datasets* – imports and exports Parquet and CSV
- *no more data analysis "side-quests"* – helps you build intuition about your dataset through automatic profiling
- *no "run query" button required* – responds to each keystroke by re-profiling the resulting dataset
- *radically simple dashboards* - thoughtful, opinionated defaults to help you quickly derive insights from your data


It's best to show and not tell, so here's a little preview of Rill Developer:

![Kapture 2022-07-21 at 15 34 45](https://user-images.githubusercontent.com/5587788/180313544-a7a03fcc-ea85-4b51-8ec3-ad5ed5043571.gif)

### We want to hear from you if you have any questions or ideas to share

You can [file an issue](https://github.com/rilldata/rill-developer/issues/new/choose) directly in this repository or reach us in our [Rill discord](https://bit.ly/3unvA05) channel. Please abide by the [rill community policy](https://github.com/rilldata/rill-developer/blob/main/COMMUNITY-POLICY.md).

## Pick an install option:
Our binary is the fastest way to install Rill Developer:
- [binary](https://docs.rilldata.com/install/binary) : download the most recent version's binary.
  - [macos-arm64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.8/macos-arm64/rill)
  - [macos-x64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.8/macos-x64/rill)
  - [linux-x64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.8/linux-x64/rill)
  <!-- - [win-x64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.8/win-x64/rill.exe) -->
- `cd` to your download directory and get started with the Rill [CLI](https://docs.rilldata.com/cli).

You can also install Rill Developer using npm or docker:
- [npm](https://docs.rilldata.com/install/npm) : run `npm install -g @rilldata/rill`
- [docker](https://docs.rilldata.com/install/docker) : download our [docker container](https://hub.docker.com/r/rilldata/rill-developer)

## Quick start a new project

You can create and augment your own projects in Rill Developer using the [CLI](https://docs.rilldata.com/cli). Every project starts by initializing the experience. Once initialized, you can ingest data into the project and start the UI.

```
rill init
rill import-source /path/to/data_1.parquet
rill start

```

or try our example:

```
rill init-example
```
<!-- (Please note that the command `rill init-example` is temporarily unavailable on Windows.) -->

## More information
See our [documentation](https://docs.rilldata.com) for more information.

## Legal
By downloading and using our application you are agreeing to the [Rill Terms of Service](https://www.rilldata.com/legal/tos) and [Privacy Policy](https://www.rilldata.com/legal/privacy).
