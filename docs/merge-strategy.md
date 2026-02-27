# Merge Strategy

This document describes the branching and merge strategy for the CarVector API repository.

## Overview

We follow **trunk-based development** with a focus on short-lived, iterative changes. This approach minimizes merge conflicts, encourages code review, and enables continuous delivery.

## Branch Structure

| Branch | Purpose | Protection |
|--------|---------|------------|
| `develop` | Main development branch | Require PR, require reviews, require status checks |
| `active-develop` | Active development work (temporary) | None |
| `feature/*` | Feature branches (rare, use active-develop) | None |

### `develop` (Main Development Branch)

- **Source of truth** for all ongoing development
- **Protected** - all changes must go through pull requests
- **Target** for all feature branches

### `active-develop` (Default Working Branch)

Used for daily development work. Developers push here frequently and create PRs to `develop`.

### `feature/*` Branches

Created only when:
- Work cannot be completed in a single day
- Multiple developers need to work on a large feature in parallel
- Experimental work that needs isolation

## Pull Request Workflow

### 1. Start from an Issue

Every PR must be linked to a GitHub issue:

```bash
# Create issue first (e.g., #123)
gh issue create --title "Add user pagination" --body "Implement pagination for users endpoint"
```

### 2. Create a Branch

Branch names should reference the issue:

```bash
# For quick work, use active-develop
git checkout active-develop
git pull

# For larger features, create a feature branch
git checkout -b feature/123-user-pagination develop
```

### 3. Make Changes

Keep changes **small and focused**:

- Single feature or bug fix
- Maximum 1-2 days of work
- If larger, break into multiple PRs

### 4. Create Pull Request

From your branch to `develop`:

```bash
gh pr create --base develop --title "Add user pagination (#123)" --body "Closes #123"
```

**PR title format**: `<brief description> (#<issue-number>)`

**PR body checklist:**

- [ ] Links to related issue(s)
- [ ] Description of changes
- [ ] Testing notes
- [ ] Breaking changes (if any)

### 5. Code Review

- At least **one approval** required
- All **CI checks** must pass (lint, test)
- Address review feedback promptly

### 6. Merge

Use **Squash and Merge**:

```bash
# After approval, PR can be merged
gh pr merge 123 --squash --delete-branch
```

Squash merging keeps `develop` history clean and linear.

## PR Lifecycle

```
Issue #123 created
       |
       v
Branch created (from active-develop or develop)
       |
       v
Commits pushed (short-lived, < 2 days)
       |
       v
PR opened (target: develop)
       |
       v
CI runs (lint, test)
       |
       v
Code review
       |
       v
Approved + checks pass
       |
       v
Squash merge to develop
       |
       v
Issue #123 closed
```

## Best Practices

### Keep PRs Short

- **Ideal**: 100-300 lines of code
- **Maximum**: 500 lines - consider splitting
- **Timeframe**: Complete within 1-2 days

### One Feature per PR

- Don't bundle unrelated changes
- Don't mix refactoring with new features
- Don't include "cleanup" unless related to the issue

### Descriptive Commits

If your PR has multiple commits, write clear commit messages:

```
feat: add user pagination endpoint
fix: handle empty page number
test: add pagination tests
```

### Update `develop` First

Before creating a PR, ensure your branch is up to date:

```bash
git checkout active-develop
git pull origin develop
git checkout -b feature/my-feature
```

### Resolve Conflicts Early

If `develop` has moved ahead:

```bash
git checkout feature/my-feature
git fetch origin
git rebase origin/develop
```

## What Happens on Merge

1. Commits are **squashed** into a single commit on `develop`
2. **CI/CD pipeline** triggers automatically
3. Related **issue is closed** (if PR includes "Closes #N")

## Emergency Fixes

For urgent production issues, you may bypass normal process:

1. Create branch directly from `develop`
2. Make minimal fix
3. PR with "Hotfix" label for fast-track review
4. Merge with approval from any team member

```bash
git checkout -b hotfix/critical-bug develop
# Make fix
git push
gh pr create --base develop --labels hotfix --title "Hotfix: critical bug"
```

## References

- [Trunk-Based Development](https://trunkbaseddevelopment.com/)
- [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow)
