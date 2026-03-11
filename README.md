# go-readability fork patches

This branch contains patches applied on top of [readeck/go-readability](https://codeberg.org/readeck/go-readability) (upstream).

## Patches

| # | Patch | Description |
|---|-------|-------------|
| 1 | `0001-Rename-module-*` | Rename Go module to `github.com/jobindex-open/go-readability` |
| 2 | `0002-Replace-re2go-*` | Replace `internal/re2go` (generated DFA matchers) with stdlib `regexp` |
| 3 | `0003-Dont-score-lists-*` | Don't penalize `<ol>`, `<ul>`, `<dl>`, `<dd>`, `<dt>`, `<li>` in content scoring (job ads use lists) |

## Syncing with upstream

```sh
git checkout patches
./scripts/sync-fork.sh           # rebuilds the 'fork' branch from upstream + patches
./scripts/sync-fork.sh --dry-run # preview only
```

## When upstream has breaking changes

If `sync-fork.sh` fails due to conflicts:

1. Resolve conflicts in the fork branch (`git am --continue`)
2. Regenerate patches: `git format-patch upstream/main -o /tmp/new-patches`
3. Update this branch:
   ```sh
   git checkout patches
   rm patches/*.patch
   cp /tmp/new-patches/*.patch patches/
   git add patches/ && git commit -m "Update patches for latest upstream"
   ```

## Adding a new patch

1. On the `fork` branch, add your commit on top
2. Regenerate: `git format-patch upstream/main -o /tmp/new-patches`
3. Copy to this branch as above
