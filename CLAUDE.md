# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

```bash
# Setup and dependencies
make dependencies          # Download and tidy Go modules
make tools                # Install development tools

# Development
make build                # Build binary to ./dist/gomodoro
make gen                  # Generate code (GraphQL schema/clients)
make lint                 # Run golangci-lint
make lint-fix             # Run golangci-lint with auto-fix
make test                 # Run all tests

# Running the application
go run main.go init       # Initialize config file
go run main.go start      # Start TUI application
go run main.go serve      # Start GraphQL API server
go run main.go remain     # Show remaining time if running
```

## Architecture Overview

Gomodoro is a Pomodoro timer with clean architecture implementing domain-driven design:

### Core Structure
- `/cmd/` - CLI commands using Cobra framework
- `/internal/core/` - Business logic (entities: Pomodoro, Task; services: PomodoroService, TaskService)
- `/internal/storage/` - Data persistence interfaces and file-based implementations
- `/internal/api/` - GraphQL API server
- `/internal/tui/` - Terminal UI using tcell
- `/internal/graph/` - GraphQL schema, resolvers, and generated code
- `/internal/config/` - YAML configuration management with validation

### Key Patterns
- **Event-Driven Architecture**: Event bus for loose coupling between components
- **Repository Pattern**: Storage interfaces with concrete implementations
- **Service Layer**: Business logic encapsulated in service classes
- **Clean Architecture**: Clear separation between domain, application, and infrastructure layers

## External Integrations

- **Toggl**: Automatic time tracking during work sessions
- **Pixela**: Habit tracking visualization (GitHub-like activity graphs)
- Both integrations are optional and configured via `~/.gomodoro/config.yaml`

## Configuration

- Config file: `~/.gomodoro/config.yaml`
- Environment variable prefix: `GOMODORO_`
- Validation using go-playground/validator
- Home directory expansion supported

## GraphQL Development

- Schema files: `internal/graph/schema/*.graphqls`
- Generated code: `internal/graph/generated/`
- Client generation: Uses Khan/genqlient
- Run `make gen` after schema changes

## Key Technologies

- **CLI**: Cobra + Viper
- **TUI**: gdamore/tcell
- **GraphQL**: 99designs/gqlgen (server), Khan/genqlient (client)
- **HTTP**: go-chi/chi
- **Logging**: uber.org/zap
- **Linting**: golangci-lint with 60+ enabled rules

## Notes

- No test files currently exist - testing infrastructure is set up but tests need to be written
- Code generation is required after GraphQL schema changes
- Binary includes git hash for version tracking