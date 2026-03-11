#!/usr/bin/env bash
#
# Sync the fork branch with the latest upstream and re-apply our patches.
#
# This script lives on the 'patches' branch. It:
#   1. Fetches upstream
#   2. Extracts patch files from the current branch (patches/)
#   3. Resets the 'fork' branch to upstream/main
#   4. Applies patches with git am
#
# Usage:
#   ./scripts/sync-fork.sh              # update fork branch
#   ./scripts/sync-fork.sh --dry-run    # preview without changes
#
set -euo pipefail

UPSTREAM_REMOTE="upstream"
UPSTREAM_BRANCH="v2"
FORK_BRANCH="fork"
PATCHES_DIR="patches"

cd "$(git rev-parse --show-toplevel)"

dry_run=false
if [[ "${1:-}" == "--dry-run" ]]; then
    dry_run=true
fi

# Ensure we're on the patches branch
current_branch=$(git branch --show-current)
if [[ "$current_branch" != "patches" ]]; then
    echo "Error: must be run from the 'patches' branch (currently on '$current_branch')."
    exit 1
fi

# Ensure upstream remote exists
if ! git remote get-url "$UPSTREAM_REMOTE" &>/dev/null; then
    echo "Error: remote '$UPSTREAM_REMOTE' not found."
    echo "Add it with: git remote add $UPSTREAM_REMOTE https://codeberg.org/readeck/go-readability"
    exit 1
fi

# Fetch latest upstream
echo "==> Fetching $UPSTREAM_REMOTE..."
git fetch "$UPSTREAM_REMOTE"

upstream_head=$(git rev-parse "$UPSTREAM_REMOTE/$UPSTREAM_BRANCH")
echo "==> Upstream HEAD: $(git log -1 --format='%h %s' "$upstream_head")"

# Collect patches
patches=("$PATCHES_DIR"/*.patch)
if [[ ${#patches[@]} -eq 0 ]]; then
    echo "Error: no patch files found in $PATCHES_DIR/"
    exit 1
fi
echo "==> Found ${#patches[@]} patch(es) to apply:"
for p in "${patches[@]}"; do
    echo "    $(basename "$p")"
done

if $dry_run; then
    echo ""
    echo "Would reset '$FORK_BRANCH' to $UPSTREAM_REMOTE/$UPSTREAM_BRANCH and apply patches."
    echo "(No changes made — dry run)"
    exit 0
fi

# Copy patches to a temp dir (we'll switch branches)
tmp_patches=$(mktemp -d)
cp "$PATCHES_DIR"/*.patch "$tmp_patches/"
trap 'rm -rf "$tmp_patches"' EXIT

# Create or reset the fork branch
if git show-ref --verify --quiet "refs/heads/$FORK_BRANCH"; then
    git checkout "$FORK_BRANCH"
    git reset --hard "$UPSTREAM_REMOTE/$UPSTREAM_BRANCH"
else
    git checkout -b "$FORK_BRANCH" "$UPSTREAM_REMOTE/$UPSTREAM_BRANCH"
fi

echo "==> Applying patches..."
if git am "$tmp_patches"/*.patch; then
    echo ""
    echo "==> Success! Fork branch '$FORK_BRANCH' is up to date."
    echo "    $(git log --oneline "$UPSTREAM_REMOTE/$UPSTREAM_BRANCH..HEAD" | wc -l) commit(s) on top of upstream:"
    git log --oneline "$UPSTREAM_REMOTE/$UPSTREAM_BRANCH..HEAD"
    echo ""
    echo "Switch back to patches branch with: git checkout patches"
else
    echo ""
    echo "==> Patch application failed. Resolve conflicts, then:"
    echo "    git am --continue"
    echo ""
    echo "  After resolving, regenerate patches and update the patches branch:"
    echo "    git format-patch upstream/main -o /tmp/new-patches"
    echo "    git checkout patches"
    echo "    cp /tmp/new-patches/*.patch patches/"
    echo "    git add patches/ && git commit -m 'Update patches for latest upstream'"
    echo ""
    echo "  Or to abort:"
    echo "    git am --abort && git checkout patches"
    exit 1
fi
