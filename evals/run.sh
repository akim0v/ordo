#!/usr/bin/env bash
#
# Runs an AI coding agent against every eval task and reports, per task,
# whether it passed on the first attempt and how many repair rounds it needed.
#
#   ./evals/run.sh                 # every task
#   ./evals/run.sh 04 07           # only tasks whose name matches
#
# The agent under test is configurable:
#   ORDO_EVAL_AGENT   command receiving the prompt on stdin, run in the workspace
#   ORDO_EVAL_MODEL   model passed to the default agent      (default: sonnet)
#   ORDO_EVAL_ROUNDS  repair rounds after the first attempt  (default: 3)

set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODEL="${ORDO_EVAL_MODEL:-sonnet}"
ROUNDS="${ORDO_EVAL_ROUNDS:-3}"
RESULTS="$(mktemp -d)/results.tsv"

# The library the tasks compile against is a snapshot with evals/ removed, so an
# agent cannot read the reference solutions through the replace directive.
SNAPSHOT="$(mktemp -d)/ordo"
mkdir -p "$SNAPSHOT"
tar -C "$REPO" --exclude=evals --exclude=.git --exclude=.idea --exclude=.claude -cf - . \
  | tar -C "$SNAPSHOT" -xf -

agent() {
  if [[ -n "${ORDO_EVAL_AGENT:-}" ]]; then
    $ORDO_EVAL_AGENT
  else
    claude -p --permission-mode acceptEdits --model "$MODEL" "$(cat)"
  fi
}

# grade restores the files the agent was told not to touch, then builds, vets
# and tests. Restoring means an agent that edited the domain or the test to make
# things pass does not score for it.
grade() {
  local work="$1" task="$2"
  cp "$task/domain.go" "$work/domain.go" 2>/dev/null
  cp "$task/verify_test.go" "$work/verify_test.go"

  (cd "$work" && go build ./... 2>&1 && go vet ./... 2>&1 && go test ./... 2>&1)
}

printf 'task\tfirst_attempt\titerations\tfinal\tdetail\n' > "$RESULTS"

for task in "$REPO"/evals/tasks/*/; do
  slug="$(basename "$task")"

  if [[ $# -gt 0 ]]; then
    match=0
    for pat in "$@"; do [[ "$slug" == *"$pat"* ]] && match=1; done
    [[ $match -eq 1 ]] || continue
  fi

  echo "=== $slug"
  work="$(mktemp -d)"
  cp "$task"/* "$work"/

  # Written rather than edited in place, so the script does not depend on a
  # particular sed dialect.
  cat > "$work/go.mod" <<GOMOD
module ordoeval/$slug

go 1.27.0

require github.com/akim0v/ordo v0.0.0

replace github.com/akim0v/ordo => $SNAPSHOT
GOMOD

  # The verification test is hidden while the agent works: the prompt states the
  # contract, and the question is whether the docs are enough to meet it.
  rm -f "$work/verify_test.go"

  prompt="$(cat "$task/PROMPT.md")"
  first="no"; iterations=0; final="fail"; detail=""

  for (( round=0; round<=ROUNDS; round++ )); do
    ( cd "$work" && printf '%s\n' "$prompt" | agent ) >/dev/null 2>&1

    if output="$(grade "$work" "$task" 2>&1)"; then
      final="pass"
      [[ $round -eq 0 ]] && first="yes"
      iterations=$round
      break
    fi

    iterations=$round
    detail="$(printf '%s' "$output" | grep -m1 -E 'ordo:|undefined:|cannot|expected|FAIL|\.go:[0-9]+' | cut -c1-160)"
    rm -f "$work/verify_test.go"

    prompt="The previous attempt did not pass. Fix the code in this directory.

go build ./... && go vet ./... && go test ./... reported:

$output"
  done

  printf '%s\t%s\t%s\t%s\t%s\n' "$slug" "$first" "$iterations" "$final" "$detail" >> "$RESULTS"
  printf '    %s after %s iteration(s)\n' "$final" "$iterations"
  rm -rf "$work"
done

echo
echo "| task | first attempt | iterations | final | stall cause |"
echo "| --- | --- | --- | --- | --- |"
tail -n +2 "$RESULTS" | while IFS=$'\t' read -r t f i r d; do
  echo "| \`$t\` | $f | $i | $r | ${d:-—} |"
done

passed=$(tail -n +2 "$RESULTS" | awk -F'\t' '$2=="yes"' | wc -l | tr -d ' ')
total=$(tail -n +2 "$RESULTS" | wc -l | tr -d ' ')
echo
echo "first-attempt pass rate: $passed/$total"
