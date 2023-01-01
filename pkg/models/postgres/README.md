# Postgres

## Table of contents

- [User Struct](#user-struct)
    - [User](#user)
    - [UserAccount](#useraccount)
    - [UserLoginCredentials](#userlogincredentials)

<br/>

## User Struct

Please see the `Liquibase` migration script for the table [schema](../../../SQL/README.md).

### User

This struct contains the `ClientId` unique identifier and a `IsDeleted` soft delete toggle. The data
in the root of this struct is intended to be internal. Embedded within is the `UserAccount`.

### UserAccount

This struct is created to be exposed for use with the HTTP handlers. This ensures consistency with the
`User` `struct`.

Contents of this struct are `FirstName`, `LastName`, and `Email` of the user. This data is not intended
to be shared with anyone other than the account holder. Embedded within is the `UserLoginCredentials`
struct.

### UserLoginCredentials

This struct contains the `Username` and `Password` for the account. The password is stored in the database
as an encrypted hash (bcrypt).

This struct is also passed by the HTTP handlers to the backend for login requests. In such an event,
the plaintext password is hashed for comparison with the stored password hash as close to the HTTP handler
as possible.
