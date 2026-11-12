# IssueMatch

IssueMatch recommends open-source issues to developers. It has a Go API on MongoDB and a React UI.

## Requirements

- Docker with Docker Compose, or
- Go 1.26, Node 22 and MongoDB 7.

## Run with Docker

```
make docker-up
make docker-seed
```

The UI is at http://localhost:5173. The API is at http://localhost:8080/api.

`make docker-seed SIZE=medium` or `SIZE=large` loads more data. `make docker-down` stops the containers.

## Run without Docker

Start MongoDB on port 27017. Then start the backend:

```
cd backend
cp .env.example .env
go run ./cmd/seed
go run ./cmd/server
```

And the frontend:

```
cd frontend
npm install
npm run dev
```

Settings are in `backend/.env`. Set `GITHUB_TOKEN` there to import issues from GitHub.

## Real issues only

By default the seed fills the gap between imported issues and the chosen size with generated ones. The `-real` flag turns that off. It needs `GITHUB_TOKEN` and reads up to 500 entries per repository in `FEED_REPOS`.

```
make seed FLAGS=-real
make docker-seed FLAGS=-real
```

With Docker, run `make docker-up` after changing `backend/.env`.

`make test` runs the Go tests.

## Benchmark

The benchmark needs a running API and a seeded database. It sends 200 requests to each scenario, prints a table and appends the rows to `backend/bench/results.csv`.

```
make seed SIZE=small
make bench
make bench FLAGS=-drop-indexes
```

`-drop-indexes` drops all indexes before the run and recreates them after. Repeat with `SIZE=medium` and `SIZE=large`. With Docker, use `make docker-seed SIZE=...` instead of `make seed`.

## Accounts

| Role | Email | Password |
| --- | --- | --- |
| Admin | admin@issuematch.local | admin12345 |
| Demo user | demo@issuematch.local | demo12345 |

The seed also creates synthetic users. They all have the password `password123`.

## API

See [docs/API.md](docs/API.md).
