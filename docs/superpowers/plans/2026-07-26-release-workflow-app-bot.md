# Release Workflow GitHub App Bot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `.github/workflows/release.yml`'s `bump-chart-version` job so the chart-bump PR is
opened by a dedicated GitHub App (a distinct actor from `GITHUB_TOKEN`), which makes the PR
actually trigger `ci.yml`, and so the workflow approves + auto-merges that PR itself using
`GITHUB_TOKEN` as a separate actor (avoiding the "can't approve your own PR" rule) — without
weakening branch protection on `main`.

**Architecture:** Two token identities in the same job: an installation token from the
`my-hours-bot` GitHub App (via `actions/create-github-app-token`) does checkout/commit/push/PR-create;
plain `GITHUB_TOKEN` does approve + auto-merge. A new step-output (`steps.bump.outputs.changed`)
replaces the old single-script `exit 0` short-circuit, gating every step from git-identity onward.

**Tech Stack:** GitHub Actions YAML, `actions/create-github-app-token`, GitHub CLI (`gh`).

## Global Constraints

- Pin every action by full 40-char commit SHA with a trailing `# vX.Y.Z` comment — matches every
  existing action reference in this repo's workflows (see `release.yml`, `ci.yml`).
- Do not touch `build-backend` / `build-frontend` — they are unaffected by this change.
- Do not weaken branch protection or bypass required status checks — use `gh pr merge --auto`, not
  an immediate/forced merge (per `CLAUDE.md`: "Don't weaken security steps to make checks pass").
- Secrets used: `secrets.AV_HELPER_BOT_APP_ID`, `secrets.AV_HELPER_BOT_PRIVATE_KEY` (already added
  to the repo by the user). Do not print them or the derived installation token in logs.
- `bump-chart-version` job permissions become `contents: write` + `pull-requests: write` (for
  `GITHUB_TOKEN`'s approve/merge use — see spec "Key constraint: self-approval"). The App's own
  token scope is independent of this block (governed by what was granted on App installation).
- Full design context: `docs/superpowers/specs/2026-07-26-release-workflow-app-bot-design.md`.

---

### Task 1: Rewrite the `bump-chart-version` job in `release.yml`

**Files:**
- Modify: `.github/workflows/release.yml:71-115` (the entire `bump-chart-version` job)

**Interfaces:**
- Consumes: repo secrets `AV_HELPER_BOT_APP_ID`, `AV_HELPER_BOT_PRIVATE_KEY`; the existing
  `actions/checkout` pin already used elsewhere in this file
  (`3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1`).
- Produces: none (this is the last job in the file; no other job depends on its outputs).

There is no unit-testable code here — this is a YAML workflow definition. "Testing" for this task
means: (a) the file is valid YAML, (b) a step-by-step read-through matches the spec's step list,
(c) real behavioral validation happens on the next actual GitHub release (tracked as a manual
follow-up, not automatable here).

- [ ] **Step 1: Read the current job to confirm line range**

Run: `grep -n "bump-chart-version:" -A 60 .github/workflows/release.yml`

Confirm the job starts at `bump-chart-version:` and ends at the last line of the file (the
`gh pr create` block). Note the exact starting line number for the Edit in Step 2.

- [ ] **Step 2: Replace the job body**

Replace the entire `bump-chart-version:` job (from the `bump-chart-version:` line to the end of the
file) with:

```yaml
  bump-chart-version:
    needs: [build-backend, build-frontend]
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - name: Generate app token
        id: app-token
        uses: actions/create-github-app-token@bcd2ba49218906704ab6c1aa796996da409d3eb1 # v3.2.0
        with:
          app-id: ${{ secrets.AV_HELPER_BOT_APP_ID }}
          private-key: ${{ secrets.AV_HELPER_BOT_PRIVATE_KEY }}

      - name: Checkout repository
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          token: ${{ steps.app-token.outputs.token }}

      - name: Compute chart version
        id: version
        run: |
          TAG="${{ github.event.release.tag_name }}"
          echo "version=${TAG#v}" >> "$GITHUB_OUTPUT"

      - name: Bump Chart.yaml version/appVersion
        id: bump
        run: |
          VERSION="${{ steps.version.outputs.version }}"
          FILES="charts/my-hours/Chart.yaml charts/my-hours/charts/backend/Chart.yaml charts/my-hours/charts/frontend/Chart.yaml"
          sed -i "s/^version: .*/version: ${VERSION}/" $FILES
          sed -i "s/^appVersion: .*/appVersion: \"${VERSION}\"/" $FILES

          git add $FILES
          if git diff --cached --quiet; then
            echo "Chart is already at version ${VERSION}, nothing to bump"
            echo "changed=false" >> "$GITHUB_OUTPUT"
          else
            echo "changed=true" >> "$GITHUB_OUTPUT"
          fi

      - name: Configure git identity as the app bot
        if: steps.bump.outputs.changed == 'true'
        env:
          GH_TOKEN: ${{ steps.app-token.outputs.token }}
          APP_SLUG: ${{ steps.app-token.outputs.app-slug }}
        run: |
          BOT_USER_ID=$(gh api "/users/${APP_SLUG}[bot]" --jq .id)
          git config user.name "${APP_SLUG}[bot]"
          git config user.email "${BOT_USER_ID}+${APP_SLUG}[bot]@users.noreply.github.com"

      - name: Commit and push chart bump
        if: steps.bump.outputs.changed == 'true'
        env:
          VERSION: ${{ steps.version.outputs.version }}
        run: |
          BRANCH="chart-bump/${VERSION}"
          git checkout -b "$BRANCH"
          git commit -m "chore: bump chart version to ${VERSION}"
          git push --force origin "$BRANCH"

      - name: Open chart version bump PR
        if: steps.bump.outputs.changed == 'true'
        env:
          GH_TOKEN: ${{ steps.app-token.outputs.token }}
          VERSION: ${{ steps.version.outputs.version }}
          RELEASE_TAG: ${{ github.event.release.tag_name }}
        run: |
          BRANCH="chart-bump/${VERSION}"
          if gh pr view "$BRANCH" >/dev/null 2>&1; then
            echo "PR for $BRANCH already exists, skipping create"
          else
            gh pr create \
              --title "chore: bump chart version to ${VERSION}" \
              --body "Automated version bump following release \`${RELEASE_TAG}\`." \
              --base main \
              --head "$BRANCH"
          fi

      - name: Approve and auto-merge PR
        if: steps.bump.outputs.changed == 'true'
        env:
          GH_TOKEN: ${{ github.token }}
          VERSION: ${{ steps.version.outputs.version }}
        run: |
          BRANCH="chart-bump/${VERSION}"
          gh pr review "$BRANCH" --approve
          gh pr merge "$BRANCH" --auto --squash --delete-branch
```

Note the `git checkout -b "$BRANCH"` moved into the "Commit and push chart bump" step (it must run
after `git add`/`sed` but the branch itself only needs to exist once we know there's something to
commit — functionally equivalent to the original ordering, just split across steps).

- [ ] **Step 3: Validate YAML syntax**

Run:
```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo "valid YAML"
```
Expected: `valid YAML` with no exception.

- [ ] **Step 4: Read back the full file and diff against the spec's step list**

Run: `git diff .github/workflows/release.yml`

Manually confirm, line by line against
`docs/superpowers/specs/2026-07-26-release-workflow-app-bot-design.md`'s "Steps" section:
- `build-backend` / `build-frontend` jobs are byte-for-byte unchanged.
- Job permissions are exactly `contents: write` + `pull-requests: write`.
- Every step after "Bump Chart.yaml version/appVersion" that touches git/gh has
  `if: steps.bump.outputs.changed == 'true'`.
- The App token (`steps.app-token.outputs.token`) is used for checkout, git identity resolution,
  and `gh pr create` — never for the approve/merge step.
- Plain `${{ github.token }}` is used only in the final "Approve and auto-merge PR" step.
- The `actions/create-github-app-token` pin is the full SHA
  `bcd2ba49218906704ab6c1aa796996da409d3eb1` with trailing comment `# v3.2.0`.

Fix any mismatch directly in the file before proceeding.

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "$(cat <<'EOF'
ci: open chart-bump PR via GitHub App, approve+auto-merge via GITHUB_TOKEN

PRs authored with GITHUB_TOKEN don't trigger further workflow runs, so
ci.yml never ran against the chart-bump PR. Opening it as a distinct
GitHub App actor fixes that; GITHUB_TOKEN then approves and enables
auto-merge as a separate actor (GitHub blocks self-approval), so
required status checks and reviews on main stay enforced.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
EOF
)"
```

**Manual follow-up (cannot be automated in this session):** the only real end-to-end validation is
publishing an actual GitHub release and confirming in the Actions tab that (a) the chart-bump PR is
authored by `my-hours-bot[bot]`, (b) `ci.yml` runs against it, (c) it shows an approval from
`github-actions[bot]`, (d) auto-merge completes once checks pass, (e) the branch is deleted after
merge.
