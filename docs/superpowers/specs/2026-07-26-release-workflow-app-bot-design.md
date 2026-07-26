# Release workflow: GitHub App bot for the chart-bump PR — design

Date: 2026-07-26

## Problem

`release.yml`'s `bump-chart-version` job opens a PR (`chart-bump/<version>`) using `GITHUB_TOKEN`
(actor `github-actions[bot]`). Two problems:

1. `gh pr create` fails outright: `GitHub Actions is not permitted to create or approve pull
   requests`, unless the repo-wide toggle is enabled — and even then, GitHub actively deprioritizes
   PR-workflow trust for PRs authored by `GITHUB_TOKEN`.
2. More fundamentally: pushes/PRs made with `GITHUB_TOKEN` do not trigger further workflow runs
   (GitHub's anti-recursion rule). So even once PR creation succeeds, the repo's `pull_request`-
   triggered CI (`ci.yml`) never runs against the chart-bump PR — defeating the purpose of using a
   PR (getting it checked before merge into a branch-protected `main`).

## Goal

Redo the PR-opening step to use a dedicated GitHub App ("my-hours-bot", already created; installed
on this repo; permissions Contents RW / Pull requests RW / Metadata R). Its installation token is a
distinct actor from `GITHUB_TOKEN`, so PRs it opens trigger `ci.yml` normally. Then have the
workflow approve and auto-merge that PR so the release process stays hands-off, without weakening
branch protection on `main` (required status checks + required reviews stay enforced for everyone).

## Non-goals

- No change to `build-backend` / `build-frontend` (GHCR push via `GITHUB_TOKEN` is unaffected).
- No change to branch protection rules themselves — this only changes who acts and how.
- No second bot account/App. One App (`my-hours-bot`) handles checkout/commit/push/PR-create;
  plain `GITHUB_TOKEN` (as `github-actions[bot]`) handles approve/merge.

## Key constraint: self-approval

GitHub rejects a PR review approval when the approver is the same actor as the PR author,
regardless of token type. Since the App now authors the PR, the App's own token cannot approve it.
`GITHUB_TOKEN` *can* approve it, because `github-actions[bot]` is a different actor than
`my-hours-bot[bot]` — this only requires the repo setting **"Allow GitHub Actions to create and
approve pull requests"** (already enabled) plus `pull-requests: write` permission on the job.

## Design

### `bump-chart-version` job — permissions

```yaml
permissions:
  contents: write       # GITHUB_TOKEN needs this to perform the squash-merge itself
  pull-requests: write  # GITHUB_TOKEN needs this to approve + enable auto-merge
```

The App's own installation token is not governed by this block — its scope comes from what was
granted to the App on installation (Contents RW / Pull requests RW / Metadata R), independent of
the workflow's `GITHUB_TOKEN` permissions.

### Steps

Note: the original script did "bump files → exit 0 if nothing changed → else commit/push/PR" as
one shell step, so `exit 0` skipped the rest for free. Splitting create-vs-approve across two
token identities means it's no longer one step — so the "nothing to bump" case is instead carried
as an explicit step output (`changed`) and every step from git-identity onward is gated on it with
`if:`, rather than relying on an early `exit`.

1. **Generate app token** — `actions/create-github-app-token@bcd2ba49218906704ab6c1aa796996da409d3eb1`
   (pinned SHA, `# v3.2.0`) with `app-id: ${{ secrets.AV_HELPER_BOT_APP_ID }}`,
   `private-key: ${{ secrets.AV_HELPER_BOT_PRIVATE_KEY }}`. Outputs `token`, `app-slug`.
2. **Checkout** — existing `actions/checkout` pin, `with: token: ${{ steps.app-token.outputs.token
   }}` (default `persist-credentials: true`, so the later `git push` authenticates via the App
   automatically — no manual remote URL rewriting needed).
3. **Compute chart version** (`id: version`) — unchanged logic, outputs `version`.
4. **Bump Chart.yaml files** (`id: bump`) — same `sed` edits as today, then `git add` the three
   files and record whether anything actually changed:
   ```
   if git diff --cached --quiet; then
     echo "Chart is already at version ${VERSION}, nothing to bump"
     echo "changed=false" >> "$GITHUB_OUTPUT"
   else
     echo "changed=true" >> "$GITHUB_OUTPUT"
   fi
   ```
5. **Resolve bot identity + configure git** (`if: steps.bump.outputs.changed == 'true'`) —
   `gh api /users/${{ steps.app-token.outputs.app-slug }}[bot] --jq .id` (using the app token as
   `GH_TOKEN`) to get the bot's numeric user id, then:
   ```
   git config user.name "${{ steps.app-token.outputs.app-slug }}[bot]"
   git config user.email "<user-id>+${{ steps.app-token.outputs.app-slug }}[bot]@users.noreply.github.com"
   ```
   This is GitHub's documented format for attributing commits to an App correctly (name/avatar in
   the UI), not just a cosmetic guess.
6. **Commit + push** (`if: steps.bump.outputs.changed == 'true'`) — same
   `git commit` + `git push --force origin "$BRANCH"` behavior as today (force-push covers reruns
   against an existing bump branch for the same tag).
7. **Open PR** (`if: steps.bump.outputs.changed == 'true'`) — same `gh pr view` / `gh pr create`
   reuse-or-create logic as today, `GH_TOKEN: ${{ steps.app-token.outputs.token }}` (the App).
8. **Approve + auto-merge (new)** (`if: steps.bump.outputs.changed == 'true'`) — `GH_TOKEN: ${{
   github.token }}` (plain `GITHUB_TOKEN`):
   ```
   gh pr review "$BRANCH" --approve
   gh pr merge "$BRANCH" --auto --squash --delete-branch
   ```
   `--auto` enables GitHub's native auto-merge: the merge only happens once required status checks
   report success, respecting existing branch protection rather than bypassing it.

### Action pin

`actions/create-github-app-token` has no existing pin in this repo. Use the latest release,
`v3.2.0`, pinned by full commit SHA per repo convention: `bcd2ba49218906704ab6c1aa796996da409d3eb1`.

## Error handling

- If the App token generation fails (bad app id / key / not installed on repo), the job fails fast
  at step 1 — no partial state.
- The existing "nothing to bump" (`git diff --cached --quiet`) check is preserved, now as the
  `steps.bump.outputs.changed` gate on every later step, so reruns against an already-bumped chart
  don't fail or open an empty PR.
- `gh pr review --approve` on a PR that's already approved is idempotent (no-op, not an error) —
  relevant if this job reruns against an existing bump branch/PR.

## Testing / validation

No local emulation of GitHub App auth is practical. Validation is: publish a real GitHub release
against this repo, confirm in the Actions tab that (a) the chart-bump PR is authored by
`my-hours-bot[bot]`, (b) `ci.yml` runs against that PR, (c) the PR shows an approval from
`github-actions[bot]`, (d) auto-merge engages and completes once checks pass, (e) the branch is
deleted after merge.
