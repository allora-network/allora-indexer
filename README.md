 Allora Chain Data Pump
======================

Overview
--------

This project serves as a data pump for the Allora Chain, a blockchain built on the Cosmos network. It extracts and stores blockchain data into a PostgreSQL database, utilizing the `allorad` application for data retrieval and the Go programming language for backend operations.

Features
--------

*   Retrieval of the latest block and consensus parameters from the Allora Chain using the `allorad` application.
*   Storage of blockchain data in a PostgreSQL database for historical analysis, monitoring, and reporting.
*   Modular Go codebase designed for easy maintenance, scalability, and integration with other systems.
*   Efficient data handling and management through the `pgx` PostgreSQL driver.

Prerequisites
-------------

*   Go (version 1.21 or later) for backend development.
*   PostgreSQL (version 10 or later) for database management.
*   The `allorad` command-line application for interacting with the Allora Chain.

Setup and Configuration
-----------------------

Ensure the `allorad` application is configured correctly to connect to the Allora Chain node. The PostgreSQL database should be set up with the necessary tables and permissions for data storage. Modify the Go application's configuration to point to the correct database and node endpoints.

Running the Data Pump
---------------------

To start the data pump, execute the compiled Go application. It will begin to fetch data from the Allora Chain node and populate the PostgreSQL database with the latest blockchain information.

```bash
go run . --node=https://rpc.network:443 -cliApp=allorad --conn=postgres://default:password@localhost:5432/catalog
```