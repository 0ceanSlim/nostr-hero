# Pubkey Quest - Documentation

## API Documentation

Interactive Swagger API docs are available at the live demo:

**[pubkey.quest/swagger](https://pubkey.quest/swagger)**

When running locally, Swagger UI is served at `/swagger/` on your configured port (e.g. `http://localhost:8584/swagger/`).

API annotations live in the Go source code (`cmd/server/api/routes.go`) and docs are generated at build time using [swag](https://github.com/swaggo/swag). See the development guide for regeneration instructions.

## Development

The [development guide](./development/readme.md) covers everything you need to get the game server running locally and start contributing:

- Prerequisites and quick start setup
- Running the server with Air (live-reload)
- Frontend development with Vite
- Database migration
- Regenerating API docs
- Project structure overview
- Deployment and auto-deploy pipeline

## Other Documentation

| Folder | Contents |
|--------|----------|
| [development/](./development/) | Setup, contributing, deployment guides |
| [development/deployment/](./development/deployment/) | Auto-deploy scripts, service templates, webhook setup |
| [development/examples/](./development/examples/) | Example config files for first-time setup |
| [draft/](./draft/) | Planning documents and future feature designs |
| [api/](./api/) | Generated Swagger files (gitignored, built at deploy time) |
