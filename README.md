# Coding task for {company_name} interview

The task is to implement an application that receives a .CSV file in a designated folder, saves it to the data storage,
and provides an HTTP REST API that allows querying the data by a GET request for a specific ID.

## Description

The .CSV file contains three columns:
- id - a UUID field
- price - a floating point number with high precision
- expiration_date - a timestamp with a timezone

The .CSV file can potentially be large and contain billions of rows.

Also, the HTTP REST API can receive a high load (millions of requests per minute). 
