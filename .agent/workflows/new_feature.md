---
description: Workflow for starting a new feature branch and working with database models
---

# New Feature Development Workflow

This workflow guides you through creating a new feature branch, specifically for tasks involving database changes that require `make models`.

## 1. Start from a Clean State

Ensure you are on the latest `main` branch to avoid conflicts later.

```bash
git checkout main
git pull upstream main
```

## 2. Create a New Feature Branch

Use a descriptive name for your branch.

```bash
git checkout -b feat/your-feature-name
```

## 3. Initial State Verification

Before making any changes, ensure the repository is in a good state. This prevents "fixing" issues that aren't yours.

```bash
// turbo
make models
make test
```

## 4. Development Loop

### Modifying Database Queries
If your feature involves database changes:

1.  Edit `gtfsdb/query.sql` (or `gtfsdb/schema.sql` for schema changes).
2.  **IMMEDIATELY** run `make models` to check your SQL syntax and regenerate Go code.
    ```bash
    // turbo
    make models
    ```
3.  If `make models` fails, fix the SQL syntax in `query.sql` and retry. Do not write Go code until `make models` passes.

### Implementing Go Logic
1.  Use the generated structs in `gtfsdb/models.go` and methods in `gtfsdb/query.sql.go`.
2.  Implement your feature in `internal/`.
3.  Periodically run lint and tests:
    ```bash
    make lint
    make test
    ```

## 5. Pre-Commit Quality Gate

Before committing, ensure everything is clean.

```bash
// turbo
make models
go fmt ./...
make lint
make test
```

## 6. Commit and Push

```bash
git add .
git commit -m "feat: description of your feature"
git push origin feat/your-feature-name
```
