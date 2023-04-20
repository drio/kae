# Little snitch

## TODO

Create SQL tables:

    - token
    - ping

Create tokens manually via SQL

Add endpoint to:

    - /hb/<token>: update token ping store the value (get post?) and update db

Write background job to every minute check of heart beats

## Structure

- main.go

- server.go

* type server (uses the model to interact with the DB)
  - model
  - logger
  - newServer()
  - methods for the server: addRoutes(),
* interface model
  - type SQLModel
  - NewSQLModel(): creates the tables if necessary and returns the SQLModel
  - All the methods that work against the SQLModel

- db.go
