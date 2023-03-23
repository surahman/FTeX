# Postgres

## Table of contents

- [Case Study and Justification](#case-study-and-justification)
- [Design Considerations](#design-considerations)
- [Tablespaces](#tablespaces)
- [Users Table Schema](#users-table-schema)
- [Fiat Accounts Table Schema](#fiat-accounts-table-schema)
- [Fiat General Ledger Table Schema](#fiat-general-ledger-table-schema)
- [Special Purpose Accounts](#special-purpose-accounts)
- [SQL Queries](#sql-queries)
- [Schema Migration and Setup](#schema-migration-and-setup)

<br/>

## Case Study and Justification

A Relational Database is a requirement for this application for the following reasons:
* Anticipated balance of reads and writes.
* Strong ACID semantics requirements.
* Eventual consistency is inadequate.
* Transactions are mandatory to record entries for accounts and in the general ledger.
* Data denormalization is extremely risky and unacceptable.
* Foreign key integrity must be maintained across tables.
* Workloads will frequently require table joins.

The ideal candidate for this application is Cockroach DB. `CRDB` is like ice cream on a scorching day;
it has the sugary sweetness of a SQL database and the soothing coolness of a distributed NoSQL database.

* PostgresSQL dialect.
* Distributed SQL.
* Opensource and well maintained.
* Widely used and respected in industry.
* Provides ACID guarantees.
* Tunable consistency levels.
* Row level TTL support.
* Works with most SQL database drivers that support Postgres.
* Architecture supports masters and hot, as well as warm, replicas for fault tolerance.
* Architecture supports read replicas for horizontal scaling that provides high availability.

For simplicity in setup with CI on GitHub, and the fact that this is project is a toy, I will be developing
on the Postgres database. It would be a quick and easy transition to use this project backed with `CRDB`.

<br/>

## Design Considerations

Before designing the table schemas it is important to understand the fundamental queries that will
be executed. This will allow for the optimization of table columns and indices to improve disk access
and memory consumption on both volatile and persistent memory.

Database schemas should adhere to the Open-Closed principle: open to extension (adding new tables) but
closed to modification.

Table schemas in their most relaxed form should adhere to the Open-Closed principle. Adding new columns
will require allowing `NULL` data types or backfilling data. `NULL` data should be avoided whenever
possible.

* Schema data alignment to reduce memory consumption on persistent and volatile memory.
* The most taxing queries are likely to be _"retrieve all of my assets from the accounts tables"_ and
  _"retrieve all of my transactions from the general ledgers"_.
* Configure indices _(simple or compound)_ for clustering. This will allow related account and transaction data
  to be co-located on persistent storage. as a result, block reads of data that will retrieve all the related
  required data on consecutive blocks. This will also, in turn, potentially lead to fewer instances of
  volatile memory page thrashing. Postgres does not support clustering indices and requires periodic execution
  of the `CLUSTER` command.
* Configure primary and secondary indices based on queries and table join keys.
* A general ledger table should contain double entries for each transaction: **_source_** and **_destination_**
  accounts for audit and fault tolerance purposes.
* General concurrency considerations whilst holding locks in transactions:
  1. Mutual exclusion will be provided by the concurrency control manager.
  2. No hold-and-wait to be handled through timeouts.
  3. No preemption to be handled through timeouts.
  4. No circular waits to be handled through a consistent ordering of lock acquisitions.
* `Tablespace`s should be created for the `production` database's tables to allow for partition resizing
  and performance tuning.

  <br/>

## Tablespaces

Cluster-wide tablespaces will need to be created for each of the tables in the production environment.
These directories will need to be created by the database administrator with the correct privileges
for the Postgres accounts that require access.

| Table Name | Tablespace Name | Location                 |
|------------|-----------------|--------------------------|
| users      | users_data      | `/table_data/ftex_users` |

Due to directory permission issues, the Postgres Docker containers will not utilize tablespaces. These issues can
be mitigated by mounting a volume that can be `chown`ed by the Postgres account. When using a directory on the host,
this can mean configuring permissions to allow any account to `read` and `write` to the directory.

The [tablespaces](schema/tablespaces.sql) can be configured once the data directories have been created and the requisite
permissions have been set.

Liquibase runs all migration change sets within transaction blocks. Tablespace creation cannot be completed
within transaction blocks. The migration scripts will expect the tablespaces to be created beforehand.
The migration scripts can be found [here](schema/migration_tablespace.sql) for tablespaces, and
[here](schema/migration.sql) for without tablespaces.

<br/>

## Users Table Schema

| Name (Struct) | Data Type (Struct) | Column Name | Column Type | Description                                                                                                                                               |
|---------------|--------------------|-------------|-------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| ClientID      | pgtype.UUID        | client_id   | UUID        | The client id ia a unique identifier generated by UUID algorithm. It is the primary key and is automatically generated by the database when not provided. |
| Username      | string             | username    | varchar(32) | The username unique identifier.                                                                                                                           |
| Password      | string             | password    | varchar(32) | User's hashed password.                                                                                                                                   |
| FirstName     | string             | first_name  | varchar(64) | User's first name.                                                                                                                                        |
| LastName      | string             | last_name   | varchar(64) | User's last name.                                                                                                                                         |
| Email         | string             | email       | varchar(64) | Email address.                                                                                                                                            |
| IsDeleted     | bool               | is_deleted  | boolean     | A soft delete indicator that prevents username reassignment.                                                                                              |

The `client_id` has been selected as the primary key. The `client_id` will be the unique identifier that
will attach the user's account to the other tables through a foreign key reference. The login operation will
be required to look up the user credentials to retrieve the `client_id`. A B-Tree index will be automatically constructed on
the `username` due to the unique constraint. This will facilitate username deduplication as well as efficient record lookup.

<br/>

## Fiat Accounts Table Schema

| Name (Struct) | Data Type (Struct) | Column Name | Column Type   | Description                                                           |
|---------------|--------------------|-------------|---------------|-----------------------------------------------------------------------|
| ClientID      | pgtype.UUID        | client_id   | UUID          | Unique identifier for the account holder. References the Users table. |
| Currency      | Currency           | currency    | Currency      | A user defined enum type for the three character currency ISO code.   |
| Balance       | pgtype.Numeric     | balance     | Numeric(18,2) | Current balance of the account correct to two decimal places.         |
| LastTx        | pgtype.Numeric     | last_tx     | Numeric(18,2) | Last transaction amount correct to two decimal places.                |
| LastTxTs      | pgtype.Timestamptz | last_tx_ts  | TIMESTAMPTZ   | Last transactions UTC timestamp.                                      |
| CreatedAt     | pgtype.Timestamptz | created_at  | TIMESTAMPTZ   | UTC timestamp at which the account was created.                       |

A compound primary key has been created on the `ClientID` and `Currency`. Each user may only have one account in each currency.
A B-Tree index has also been created on the `ClientID` to facilitate efficient querying for accounts belonging to a single user.

<br/>

## Fiat General Ledger Table Schema

| Name (Struct) | Data Type (Struct) | Column Name   | Column Type   | Description                                                                                                                                            |
|---------------|--------------------|---------------|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| TxID          | pgtype.UUID        | tx_id         | UUID          | Identifier (primary key) for the transaction. Each key will shared between two entries in the table, once for a deposit and another for a withdrawal.  |
| ClientID      | pgtype.UUID        | client_id     | UUID          | Unique identifier for the account relating to the transaction. References the Accounts table.                                                          |
| Currency      | Currency           | currency      | Currency      | A user defined enum type for the three character currency ISO code.                                                                                    |
| Amount        | pgtype.Numeric     | amount        | Numeric(18,2) | Amount for the transaction correct to two decimal places. A positive value will indicate a deposit whilst a negative value will indicate a withdrawal. |
| TransactedAt  | pgtype.Timestamptz | transacted_at | Numeric(18,2) | Last transactions UTC timestamp.                                                                                                                       |

A compound primary key has been configured on the `tx_id`, `client_id`, and `currency` which will enforce uniqueness. Two
additional indices have been created on the `client_id` and `tx_id` to support efficient record retrieval.

<br/>

## Special Purpose Accounts

| Username     | Purpose                                                                        |
|--------------|--------------------------------------------------------------------------------|
| deposit-fiat | Inbound deposits to the fiat accounts will be associated to this user account. |

Special purpose accounts will be created for the purpose of general ledger entries. These accounts will have random
password generated at creation and will be marked as deleted so disable login capabilities.

<br/>

## SQL Queries
The queries to generate the all the tables can be found in the migration [script](schema/migration.sql).

<br/>

## Schema Migration and Setup

For security reasons, there are no database schema migration tools provided through the binary. This
is to avoid deploying a payload in a production container that could potentially modify the databases'
schema. As an alternative, this project will be making use of [Liquibase](https://docs.liquibase.com/home.html)
for database migrations.

There are two scripts provided:

1. [Migration](schema/migration.sql).
2. [Migration with Tablespace](schema/migration_tablespace.sql).

The Liquibase connection information will need to be configured in the [properties](liquibase.properties) file.

Execute the following commands from the `SQL` directory:

```bash
# Main database setup
liquibase update
```

```bash
# Main database rollback. Specify number of steps.
liquibase rollback-count 5
```


```bash
# Test suite setup
liquibase update --defaultsFile liquibase_testsuite.properties
```

```bash
# Test suite setup
liquibase rollback-count 5 --defaultsFile liquibase_testsuite.properties
```
