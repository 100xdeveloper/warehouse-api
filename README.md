# Warehouse API

A Go-based REST API for managing product inventory.

## Tech Stack
- **Go** (Golang)
- **Chi** (Router)
- **Pgx** (Postgres Driver)
- **Docker** (Database)

## Setup

1. Clone the repo.
2. Create a `.env` file based on `.env.example`.
3. Start the database:
   ```bash
   docker run --name my_db -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres
   ```

1. Run the schema migration:

```Bash
cat schema.sql | docker exec -i my_db psql -U postgres
```

2. Run the app:

```Bash
go run .
```
