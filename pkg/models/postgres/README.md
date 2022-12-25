# Postgres

## Table of contents

- [Case Study and Justification](#case-study-and-justification)
- [Design Considerations](#design-considerations)
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
* The most taxing queries are likely to be _"retrieve all of my assets from the accounts table"_ and
  _"retrieve all of my transactions from the general ledger"_.
* Set clustering indices _(simple or compound)_ to allow related account and transaction data to be
  co-located on persistent storage. This will result in block reads of data that will retrieve all the
  required data on consecutive pages. This will also, in turn, potentially lead to fewer instances of
  volatile memory page thrashing.
* Configure primary and secondary indices based on queries and table join keys.
* A general ledger table should contain double entries for each transaction: **_source_** and **_destination_**
  accounts for audit and fault tolerance purposes.
* General concurrency considerations whilst holding locks in transactions:
  1. Mutual exclusion will be provided by the concurrency control manager.
  2. No hold-and-wait to be handled through timeouts.
  3. No preemption to be handled through timeouts.
  4. No circular waits to be handled through a consistent ordering of lock acquisitions.

<br>

## Schema Migration and Setup

For security reasons, there are no database schema migration tools provided through the binary. This is to avoid deploying a
payload in a production container that could potentially modify the databases' schema. As an alternative, there are three
SQL files provided that will need to be deployed either manually or through database migration tooling. The files will need
to be deployed in the following order:

1. [user table](users.sql)
2. [accounts table](accounts.sql)
3. [general ledger table](general_ledger_table.sql)
