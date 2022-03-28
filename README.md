Thank you for trying the Rill Developer tech preview! 

We want to hear from you if you have any questions or ideas to share. You can file an issue directly in this repository or reach us in our Rill Community Slack at https://bit.ly/35ijZG4. Please abide by the [Rill Community Policy](https://github.com/rilldata/rill-developer/blob/main/COMMUNITY-POLICY.md).

# Prerequisites
Nodejs version 16+ installed locally: https://nodejs.org/en/download/. Check your version of Node:
```
node -v
```
Clone this repository to your local machine:
```
git clone https://github.com/rilldata/rill-developer.git
```

# Install Locally
Change directories to the local Rill Developer repository
```
cd /path/to/rill-developer
```
Run npm to install dependencies and build the application. This will take ~5 minutes to complete.
```
npm install
npm run build
```

# Quick Start Example
If you are looking for a fast way to get started you can run our quick start example script. This script initializes a project, downloads an OpenSky Network dataset at https://zenodo.org/record/6325961#.YjDFvhDMI0Q, and imports the data. The Rill Developer UI will be available at http://localhost:8080.
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

# Updating the Rill Developer
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
npm run cli --silent -- import-table  /path/to/data_1.parquet --name my_dataset
```
If you have a dataset that is delimited by a character other than a comma or tab, you can use the --delimiter option. DuckDB can also attempt to automatically detect the delimiter, so it is not strictly necessary.
```
npm run cli --silent -- import-table /path/to/data_4.txt --delimiter "|"
```
If you would like to see information on all of the available CLI commands, you can use the help option.
```
npm run cli --silent -- --help
```

# Legal

By downloading and using our application you are agreeing to the Rill [Terms of Service](https://www.rilldata.com/legal/tos) and [Privacy Policy](https://www.rilldata.com/legal/privacy).


# Application Developers
If you are a developer helping us build the application, please visit our [DEVELOPER-GUIDE.md](https://github.com/rilldata/rill-developer/blob/main/DEVELOPER-GUIDE.md).
