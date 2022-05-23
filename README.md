# Rill Developer **_(tech preview)_**

Rill Developer is a tool that makes it effortless to transform your datasets with SQL. It's not just a SQL GUI! Rill Developer follows a few guiding principles:

- _no more data analysis "side-quests"_ – helps you build intuition about your dataset through automatic profiling
- _no "run query" button required_ – responds to each keystroke by re-profiling the resulting dataset
- _works with your local datasets_ – imports and exports Parquet and CSV
- _feels good to use_ – powered by Sveltekit & DuckDB = conversation-fast, not wait-ten-seconds-for-result-set fast

It's best to show and not tell, so here's a little preview of Rill Developer:

![RillDeveloper](https://user-images.githubusercontent.com/5587788/160640657-2b68a230-9dcb-4236-a6c8-df5263c33443.gif)

## We want to hear from you if you have any questions or ideas to share.

You can [file an issue](https://github.com/rilldata/rill-developer/issues/new/choose) directly in this repository or reach us in our [Rill Discord](https://bit.ly/3unvA05) channel. Please abide by the [Rill Community Policy](https://github.com/rilldata/rill-developer/blob/main/COMMUNITY-POLICY.md).

# Prerequisites

Nodejs version 16+ installed locally: https://nodejs.org/en/download/. Check your version of Node:

```
node -v
```

Clone this repository to your local machine:

```
git clone https://github.com/rilldata/rill-developer.git
```

On Ubuntu, you'll also need to make sure you have `g++` installed in order to compile DuckDB from source during the installation steps below (please note that compiling DuckDB may take a while):

```
sudo apt install g++
```

# Install Locally

Change directories to the local Rill Developer repository

```
cd /path/to/rill-developer
```

Run npm to install dependencies and build the application.
Part of this step involves compiling DuckDB which can be time consuming to complete (it may take approximately five minutes or more, depending on your machine).
Please be patient.

```
npm install
npm run build
```

# Quick Start Example

If you are looking for a fast way to get started you can run our quick start example script. This script initializes a project, downloads an [OpenSky Network dataset](https://zenodo.org/record/6325961#.YjDFvhDMI0Q), and imports the data. The Rill Developer UI will be available at http://localhost:8080.

```
bash scripts/example-project.sh
```

If you close the example project and want to restart it, you can do so by running:

```
npm run cli --silent -- start --project ../rill-developer-example
```

# Creating Your Own Project

If you want to go beyond this example, you can also create a project using your own data.

## Initialize Your Project

Initialize your project in the Rill Developer directory.

```
npm run cli --silent -- init
```

## Import Your Data

Import datasets of interest into the Rill Developer [duckDB](https://duckdb.org/docs/sql/introduction) database to make them available. We currently support .parquet, .csv, and .tsv.

```
npm run cli --silent -- import-table /path/to/data_1.parquet
npm run cli --silent -- import-table /path/to/data_2.csv
npm run cli --silent -- import-table /path/to/data_3.tsv
```

## Start Your Project

Start the User Interface to interact with your imported tables and revisit projects you have created.

```
npm run cli --silent -- start
```

The Rill Developer UI will be available at http://localhost:8080.

# Rill Developer SQL Dialect

Rill Developer is powered by duckDB. Please visit their documentation for insight into their dialect of SQL to facilitate your queries at https://duckdb.org/docs/sql/introduction.

# Updating Rill Developer

Rill Developer will be evolving quickly! If you want an updated version, you can pull in the latest changes and rebuild the application. Once you have rebuilt the application you can restart your project to see the new experience.

```
git pull origin main
npm run build
npm run cli --silent -- start
```

# Helpful Hints

You can specify a new project folder by including the --project option.

```
npm run cli --silent -- init --project /path/to/a/new/project
npm run cli --silent -- import-table /path/to/data_1.parquet --project /path/to/a/new/project
npm run cli --silent -- start --project /path/to/a/new/project
```

By default the table name will be a sanitized version of the dataset file name. You can specify a name using the --name option.

```
npm run cli --silent -- import-table  /path/to/data_1.parquet --name my_table
```

If you have added a table to Rill Developer that you want to drop, you can do so using the --drop-table option.

```
npm run cli --silent -- drop-table my_table
```

If you have a dataset that is delimited by a character other than a comma or tab, you can use the --delimiter option. DuckDB can also attempt to automatically detect the delimiter, so it is not strictly necessary.

```
npm run cli --silent -- import-table /path/to/data_4.txt --delimiter "|"
```

You can connect to an existing duckdb database by passing --db with path to the db file.
Any updates made directly to the tables in the database will reflect in Rill Developer.
Similarly, any changes made by Rill Developer will modify the database.
Make sure to have only one connection open to the database, otherwise there will be some unexpected issues.

```
npm run cli --silent -- init --db /path/to/duckdb/file
```

You can also copy over the database so that there are no conflicts and overrides to the source.
Pass --copy along with --db to achieve this.

```
npm run cli --silent -- init --db /path/to/duckdb/file --copy
```

If you would like to see information on all the available CLI commands, you can use the help option.

```
npm run cli --silent -- --help
```

# Troubleshooting

## 404 Errors

If you have just installed the application and are trying to see the User Interface at http://localhost:8080/ but see a 404 error, it is possible that npm install is taking longer than 5 minutes to build the application and you need to wait for the build to complete. Please wait an additional 5 minutes and try again.

# Using Docker

Docker is a containerization platform that packages our application and all its dependencies together to make sure it works seamlessly in any environment. As an alternative to the Install instructions above, you can install Rill Developer using our docker container.

https://hub.docker.com/r/rilldata/rill-developer

1. Build the rill-developer using docker compose, if any changes.

   ```
   docker compose build
   ```

1. Run the rill-developer using docker compose.

   ```
   docker compose up
   ```

   Check [http://localhost:8080/](http://localhost:8080/)

   By default, it will create a project `rill-developer-example` under `./projects`
   To create a new project, update `PROJECT` in docker-compose.yml.

1. Copy over any file to import into `./projects/${PROJECT}/data/`

   ```
   docker exec -it rill-developer /bin/bash

   rill import-table ${PROJECT_BASE}/${PROJECT}/data/<fileName> \
       --project ${PROJECT_BASE}/${PROJECT}
   ```

# Legal

By downloading and using our application you are agreeing to the Rill [Terms of Service](https://www.rilldata.com/legal/tos) and [Privacy Policy](https://www.rilldata.com/legal/privacy).

# Application Developers

If you are a developer helping us build the application, please visit our [DEVELOPER-GUIDE.md](https://github.com/rilldata/rill-developer/blob/main/DEVELOPER-GUIDE.md).
