# Coding task for {company_name} interview

The task is to implement an application that
- receives a .CSV file in a designated folder
- saves it to the data storage
- provides an HTTP REST API that allows querying the data by a GET request for a specific ID

## Description

The .CSV file contains three columns:
- id - a UUID field
- price - a floating point number with high precision
- expiration_date - a timestamp with a timezone

The .CSV file can potentially be large and contain billions of rows.

Also, the HTTP REST API can receive a high load (millions of requests per minute).


## Implementation details

### How to run

#### Prerequisites

In order to run this app, you'll need Go 1.20 and a MySQL instance.

Or you can use the provided docker-compose configuration that handles everything (you'll need to have Docker installed on your machine).

#### Run without Docker

The application consists of two executables:
- HTTP API server that accepts HTTP requests
- File processor

Also, there is an additional tool to generate test data.

In order to build all available executables, you'll need to run this command:
```bash
$ make build
```

Configuration files for both apps are located at [configs directory](./configs):
- API server - [prices_app.yaml](./configs/prices_app.yaml)
- File processor - [files_app.yaml](./configs/files_app.yaml)

To run the API server, you'll need to run this command:
```bash
$ make run-prices
```

To run the File processor, you'll need to run this command:
```bash
$ make run-files
```

### Run with Docker