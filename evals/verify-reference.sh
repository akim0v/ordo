#!/usr/bin/env bash
#
# Checks that every eval task is still solvable, by building and testing it
# against its reference solution. No agent is involved.
#
# This is what keeps the suite honest: if a task stops passing here, the fixture
# has rotted against the library, and any agent failure it reports is noise.

set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail=0

for task in "$REPO"/evals/tasks/*/; do
  slug="$(basename "$task")"
  work="$(mktemp -d)"

  cp "$task"/* "$work"/
  cp "$REPO/evals/testdata/reference/$slug/wiring.go" "$work/wiring.go"

  # Written rather than edited in place, so the script does not depend on a
  # particular sed dialect.
  cat > "$work/go.mod" <<GOMOD
module ordoeval/$slug

go 1.27.0

require github.com/akim0v/ordo v0.0.0

replace github.com/akim0v/ordo => $REPO
GOMOD

  if output="$( cd "$work" && go build ./... 2>&1 && go vet ./... 2>&1 && go test ./... 2>&1 )"; then
    printf '  %-28s ok\n' "$slug"
  else
    printf '  %-28s BROKEN\n' "$slug"
    printf '%s\n' "$output" | sed 's/^/      /' | head -10
    fail=1
  fi

  rm -rf "$work"
done

exit "$fail"
