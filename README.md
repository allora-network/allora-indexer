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

The application will attempt to catch up from the last block found in the database.
However, it is possible to parse only specific events by passing the flag: `--blocks=123,456,567`. It will override the default catch-up mechanism - it will only attempt to process those blocks and add them to the database. The use of this flag is expected in testing or backfilling.



Flags and Usage
---------------

| Flag | Default Value | Description |
|------|---------------|-------------|
| `--WORKERS_NUM` | 1 | Number of workers to process blocks concurrently |
| `--NODE` | "https://allora-rpc.testnet-1.testnet.allora.network/" | Node address for the blockchain |
| `--CLIAPP` | "allorad" | CLI app to execute commands |
| `--CONNECTION` | "postgres://pump:pump@localhost:5433/pump" | Database connection string |
| `--AWS_ACCESS_KEY` | "" | AWS access key for S3 access |
| `--AWS_SECURITY_KEY` | "" | AWS security key for S3 access |
| `--S3_BUCKET_NAME` | "allora-testnet-1-indexer-backups" | AWS S3 bucket name for backups |
| `--S3_FILE_KEY` | "latest" | AWS S3 file key for the backup file in the bucket. Use "latest" to automatically fetch the most recent backup. |
| `--MODE` | "full" | Operation mode: 'full' for full update, 'dump' to load a dump and exit, 'empty' to create an empty DB and exit |
| `--RESTORE_PARALLEL_JOBS` | 4 | Number of parallel jobs (workers) to restore the dump |
| `--EXIT_APP` | false | Exit when the last block is processed. If false, will keep processing new blocks. |

## Modes

- **full**: Performs a full update of the blockchain data. It will restore from S3 if the database is empty, then process all blocks up to the latest.
- **dump**: Simply overwrites the database by loading a dump from S3 and then exits.
- **empty**: Creates an empty database and exits.

## Examples

Note: add to the examples below the `AWS_ACCESS_KEY`, `AWS_SECRET_KEY`, `S3_BUCKET_NAME` and `S3_FILE_KEY` where S3 dumps are used, 
and `CONNECTION` (of the type `postgres://default:password@db:5432/app`) where the database is hosted.

1. Run in full mode with 8 workers, after having restored the latest dump from S3:
   ```
   ./blockchain-pump --MODE=full --WORKERS_NUM=8
   ```

2. Restore the latest dump from S3:
   ```
   ./blockchain-pump --MODE=dump --S3_FILE_KEY=latest
   ```

3. Create an empty database (to examine schema and generate data, useful for testing purposes):
   ```
   ./blockchain-pump --MODE=empty
   ```

4. Run in full mode and exit when caught up (useful to run as a cron job):
   ```
   ./blockchain-pump --MODE=full --EXIT_APP=true
   ```


Recreation of database
----------------------

To recreate the database in the provided docker compose setup, the volume can be removed with the following command:
```
docker compose down
docker compose volume rm allora-data-pump_postgres_data
# In case of error, check the volume name with: docker volume ls 

```

