# Domain Layer

This package contains the core business logic and domain models.

## Structure

- `entities/` - Domain entities (Node, Edge, Graph, etc.)
- `services/` - Domain services (StabilityCalculator, etc.)
- `repositories/` - Repository interfaces for data access

## Principles

- No external dependencies (except standard library)
- Pure business logic
- Immutable entities where possible