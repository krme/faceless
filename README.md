# Codesphere AI hackathon

## Install and run

- Install dependencies with `make install`
- Run postgres database with `make docker-run`
- Run python job server with `make job`
- Run main server with `make`

## Structure

- `/jobs` includes all long running python scripts
- `/api`, `/model`, `/server`  and `/web` contain a Go backend with templ web components and a server for making all API requests for database calls and starting jobs.
