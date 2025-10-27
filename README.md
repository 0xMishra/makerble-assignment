# Makerble assignment

Assignment for makerble internship

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

Run api without building the binaries:
```bash
make run/api
```

Connect to the database:
```bash
make db/psql
```

Create a new database migration:
```bash
make db/migrations/new
```

Apply database migrations:
```bash
make db/migrations/up
```

Format all .go files and tidy module dependencies:
```bash
make tidy
```

Run quality control checks:
```bash
make audit
```

Build the application:
```bash
make build/api
```
