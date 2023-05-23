# Postgres

## Table of contents

- [Postgres](#postgres)
  - [Table of contents](#table-of-contents)
  - [Case Study and Justification](#case-study-and-justification)
  - [Design Considerations](#design-considerations)
  - [Numeric Rounding](#numeric-rounding)
    - [Rounding Half-Up](#rounding-half-up)
    - [Rounding Half-Even](#rounding-half-even)
  - [Transactions](#transactions)
  - [Tablespaces](#tablespaces)
  - [Users Table Schema](#users-table-schema)
  - [Fiat Accounts Table Schema](#fiat-accounts-table-schema)
  - [Fiat Journal Table Schema](#fiat-journal-table-schema)
  - [Crypto Accounts Table Schema](#crypto-accounts-table-schema)
  - [Crypto Journal Table Schema](#crypto-journal-table-schema)
  - [Special Purpose Accounts](#special-purpose-accounts)
  - [Journal Entries](#journal-entries)
  - [SQL Queries](#sql-queries)
  - [Schema Migration and Setup](#schema-migration-and-setup)

<br/>

## Case Study and Justification

A Relational Database is a requirement for this application for the following reasons:
* Anticipated balance of reads and writes.
* Strong ACID semantics requirements.
* Eventual consistency is inadequate.
* Transactions are mandatory to record entries for accounts and in the journals.
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
  _"retrieve all of my transactions from the journals"_.
* Configure indices _(simple or compound)_ for clustering. This will allow related account and transaction data
  to be co-located on persistent storage. as a result, block reads of data that will retrieve all the related
  required data on consecutive blocks. This will also, in turn, potentially lead to fewer instances of
  volatile memory page thrashing. Postgres does not support clustering indices and requires periodic execution
  of the `CLUSTER` command.
* Configure primary and secondary indices based on queries and table join keys.
* A journal table should contain double entries for each transaction: **_source_** and **_destination_**
  accounts for audit and fault tolerance purposes.
* General concurrency considerations whilst holding locks in transactions:
  1. Mutual exclusion will be provided by the concurrency control manager.
  2. No hold-and-wait to be handled through timeouts.
  3. No preemption to be handled through timeouts.
  4. No circular waits to be handled through a consistent ordering of lock acquisitions.
* `Tablespace`s should be created for the `production` database's tables to allow for partition resizing
  and performance tuning.

<br/>

## Numeric Rounding
The following steps will be taken to minimize the [issues](https://docs.oracle.com/cd/E19957-01/806-3568/ncg_goldberg.html)
associated with IEEE Floating point representation:

* PostgresSQL`Numeric` data type will be used.
* Golang [`decimal.Decimal`](https://pkg.go.dev/github.com/shopspring/decimal) data type will be used.
* [Half-to-Even/Bankers’ Rounding](https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even).
* Fiat assets will be stored with two decimal places.
* Crypto assets will be stored with TBD decimal places.

The PostgresSQL `Money` and `Decimal` types are synonyms for `Numeric`. Numbers associated with assets will be captured as
close to their origin as possible and converted to `decimal.Decimal` with Half-to-Even rounding to the required decimal places.

There is unfortunately no builtin method for Half-to-Even rounding as of PostgresSQL 14. As such, a User Defined Function
(UDF) will be deployed. Arithmetic rounding can lead to errors that can snowball to significant numbers over many calculations.
Please see [this article](https://www.eetimes.com/an-introduction-to-different-rounding-algorithms/) on rounding methods in the EETimes.
This UDF will be used on all functions that store financial values on the database side as a secondary safeguard. This is
necessary and will mean having to incur a calculation overhead.

The UDF does not access or modify data from any tables and thus meets the requirements for the following Postgres function characteristics:

* Immutable
* Strict
* Parallel Safe

The algorithm is as below and the function is located in the Liquibase migrations files:

### Rounding Half-Up
```math
\lceil \frac{\lfloor 2x \rfloor}{2} \rceil
```

### Rounding Half-Even

```text
Inputs:
-------
NUM   <-- num
SCALE <-- scale


Algorithm:
----------
IF number of decimal places of NUM == SCALE THEN
    RETURN NUM

multiplier <-- 10 ^ SCALE

IF (absolute value of difference of NUM Half-Up rounded with SCALE and NUM) * multiplier == 0.5 AND
    (Half-Up rounded value * mulitplier) % 2 != 0 THEN
        rounded = round Half-Up (NUM - (difference of NUM Half-Up rounded with SCALE and NUM)) to SCALE

RETURN rounded
```

<br/>

## Transactions

_**Always design and develop systems with Mechanical Sympathy in mind.**_

It is almost always best to develop transactions that leverage User Defined Procedures
([_UDPs_](https://www.postgresql.org/docs/15/xproc.html)) that return results to the backend upon completion
(success or failure). This helps to minimize the latency introduced by network communication when running staged/phased
transactions on the backend. This will in turn results in lower lock contention, resulting in higher throughput and less
resource pressure on the database instance.

As an alternative, transactions can also be comprised of User Defined Functions
([_UDFs_](https://www.postgresql.org/docs/15/xfunc.html)) that are stored on the backend database. This reduces latency
and network traffic when calling queries within the transaction in the driver code in the backend service.

Transactions for Fiat operations are being developed on the backend service here strictly as a technical presentation.
Transactions involving Crypto operations will be developed, tested, and deployed as transactions within _UDPs_.

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

The query to retrieve all Fiat account balances for a specific client leverages the index on the Fiat Accounts table.
There are a finite number of supported Fiat currencies; the aggregate number of accounts that can be retrieved is
finite. Since the index is a compound index on the `ClientID` and `Currency`, the sort operation in the query should
have a minimal performance impact. The next page cursor is merely the encrypted currency. If a user requests `N`
records, `N + 1` records are retrieved. The `N + 1`th record is used as the cursor to the following page. If only `N`
records are returned by the query it indicates an end to the data set. This iteration method will mean that new accounts
created in the currency range prior to the page cursor will not affect data contained in the pages that follow.

<br/>

## Fiat Journal Table Schema

| Name (Struct) | Data Type (Struct) | Column Name   | Column Type   | Description                                                                                                                                            |
|---------------|--------------------|---------------|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| TxID          | pgtype.UUID        | tx_id         | UUID          | Identifier (primary key) for the transaction. Each key will shared between two entries in the table, once for a deposit and another for a withdrawal.  |
| ClientID      | pgtype.UUID        | client_id     | UUID          | Unique identifier for the account relating to the transaction. References the Accounts table.                                                          |
| Currency      | Currency           | currency      | Currency      | A user defined enum type for the three character currency ISO code.                                                                                    |
| Amount        | pgtype.Numeric     | amount        | Numeric(18,2) | Amount for the transaction correct to two decimal places. A positive value will indicate a deposit whilst a negative value will indicate a withdrawal. |
| TransactedAt  | pgtype.Timestamptz | transacted_at | Numeric(18,2) | Last transactions UTC timestamp.                                                                                                                       |

A compound primary key has been configured on the `tx_id`, `client_id`, and `currency` which will enforce uniqueness. Two
additional indices have been created on the `transacted_at` and `tx_id` to support efficient record retrieval.

The query for Fiat Transaction retrieval will use the `Currency`, `Year`, `Month`, and `Timezone`. The data returned by
the query will be for all transactions for the specified currency during the year and month in the specific timezone.
The returned records will be sorted by transaction timestamps and will leverage the index on the `transacted_at` column.
Data page iteration will adopt the offset-limit method. If a user requests `N` records, `N + 1` records are retrieved.
The `N + 1`th record is used to check if there are more records to be retrieved. If only `N` records are returned by the
query it indicates an end to the data set. The anticipated number of transactions per month is not expected to exceed
1000, and the performance impact should be negligible. The page cursor consists of the record `offset`, `Currency`,
`start`, and `end` dates for the transactions and is encrypted.

<br/>

## Crypto Accounts Table Schema

| Name (Struct) | Data Type (Struct) | Column Name | Column Type   | Description                                                                                                                           |
|---------------|--------------------|-------------|---------------|---------------------------------------------------------------------------------------------------------------------------------------|
| ClientID      | pgtype.UUID        | client_id   | UUID          | Unique identifier for the account holder. References the Users table.                                                                 |
| Ticker        | string             | ticker      | VARCHAR(6)    | The ticker symbol for the cryptocurrency. Each cryptocurrency has a unique ticker symbol.                                             |
| Balance       | pgtype.Numeric     | balance     | Numeric(24,8) | Current balance of the account correct to eight decimal places. This precision is chosen because 1 Satoshi (Sat) is `BTC 0.00000001`. |
| LastTx        | pgtype.Numeric     | last_tx     | Numeric(24,8) | Last transaction amount correct to eight decimal places.                                                                              |
| LastTxTs      | pgtype.Timestamptz | last_tx_ts  | TIMESTAMPTZ   | Last transactions UTC timestamp.                                                                                                      |
| CreatedAt     | pgtype.Timestamptz | created_at  | TIMESTAMPTZ   | UTC timestamp at which the account was created.                                                                                       |

A compound primary key has been created on the `ClientID` and `Ticker`. Each user may only have one account in each cryptocurrency and the ticker is unique for each cryptocurrency.
A B-Tree index has also been created on the `ClientID` to facilitate efficient querying for accounts belonging to a single user.

The query to retrieve all Crypto account balances for a specific client leverages the index on the Crypto Accounts table.
Each is expected to have only a few different Cryptocurrencies. The top 50 Cryptocurrencies by market capitalization can
be viewed [here](https://crypto.com/price), whilst the 12 most popular, as of May 2023, can be viewed
[here](https://finance.yahoo.com/news/12-most-popular-types-cryptocurrency-221243578.html). It is thus assumed that the
aggregate number of accounts that will be retrieved for a single client will be limited to at most 50. Since the index
is a compound index on the `ClientID` and `Ticker`, the sort operation in the query should have a minimal performance
impact. The next page cursor is merely the encrypted ticker. If a user requests `N` records, `N + 1` records are
retrieved. The `N + 1`th record is used as the cursor to the following page. If only `N` records are returned by the
query it indicates an end to the data set. This iteration method will mean that new accounts created in the currency
range prior to the page cursor will not affect data contained in the pages that follow.

<br/>

## Crypto Journal Table Schema

| Name (Struct) | Data Type (Struct) | Column Name   | Column Type   | Description                                                                                                                                              |
|---------------|--------------------|---------------|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| TxID          | pgtype.UUID        | tx_id         | UUID          | Identifier (primary key) for the transaction. Each key will shared between two entries in the table, once for a deposit and another for a withdrawal.    |
| ClientID      | pgtype.UUID        | client_id     | UUID          | Unique identifier for the account relating to the transaction. References the Accounts table.                                                            |
| Ticker        | string             | ticker        | VARCHAR(6)    | The ticker symbol for the cryptocurrency. Each cryptocurrency has a unique ticker symbol.                                                                |
| Amount        | pgtype.Numeric     | amount        | Numeric(24,8) | Amount for the transaction correct to eight decimal places. A positive value will indicate a deposit whilst a negative value will indicate a withdrawal. |
| TransactedAt  | pgtype.Timestamptz | transacted_at | Numeric(24,8) | Last transactions UTC timestamp.                                                                                                                         |

A compound primary key has been configured on the `tx_id`, `client_id`, and `ticker` which will enforce uniqueness. Two
additional indices have been created on the `transacted_at` and `tx_id` to support efficient record retrieval.

The query for Crypto Transaction retrieval will use the `Ticker`, `Year`, `Month`, and `Timezone`. The data returned by
the query will be for all transactions for the specified currency during the year and month in the specific timezone.
The returned records will be sorted by transaction timestamps and will leverage the index on the `transacted_at` column.
Data page iteration will adopt the offset-limit method. If a user requests `N` records, `N + 1` records are retrieved.
The `N + 1`th record is used to check if there are more records to be retrieved. If only `N` records are returned by the
query it indicates an end to the data set. The anticipated number of transactions per month is not expected to exceed
1000, and the performance impact should be negligible. The page cursor consists of the record `offset`, `ticker`,
`start`, and `end` dates for the transactions and is encrypted.

<br/>

## Special Purpose Accounts

| Username          | Purpose                                                                            |
|-------------------|------------------------------------------------------------------------------------|
| fiat-currencies   | Inbound deposits to the fiat accounts will be associated with this user account.   |
| crypto-currencies | Inbound deposits to the crypto accounts will be associated with this user account. |

Special purpose accounts will be created for the purpose of journal entries. These accounts will have random password generated
at creation and will be marked as deleted so disable login capabilities.

<br/>

## Journal Entries

The following Journal entries will be made during deposits, exchanges, purchases, and sale operations on the platform.

* ___Deposit Fiat:___
  * Debit entry for the FTeX Fiat operations account.
  * Credit entry for the client’s destination Fiat currency account.
* ___Exchange/Convert Fiat:___
  * Debit entry for the client’s source Fiat currency account.
  * Credit entry for the client’s destination Fiat currency account.
* ___Purchase Crypto:___
  * Debit entry for the client’s source Fiat currency account.
  * Credit entry for the FTeX Fiat operations account.
  * Debit entry for the FTeX Crypto operations account.
  * Credit entry for the client’s destination cryptocurrency account.
* ___Sell Crypto___
  * Debit entry for the client’s source cryptocurrency account.
  * Credit entry for the FTeX Crypto operations account.
  * Debit entry for the FTeX Fiat operations account.
  * Credit entry for the client’s destination Fiat currency account.

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
liquibase rollback-count 9
```


```bash
# Test suite setup
liquibase update --defaultsFile liquibase_testsuite.properties
```

```bash
# Test suite setup
liquibase rollback-count 9 --defaultsFile liquibase_testsuite.properties
```
