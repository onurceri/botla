#!/usr/bin/env bash
set -euo pipefail

THRESHOLD=${THRESHOLD:-90}

go test ./... -coverprofile=coverage.out >/dev/null

FILTERED_FILE="coverage.out"

if [ -f .coverignore ] || [ -n "${COVERAGE_EXCLUDE:-}" ]; then
  FILTERED_FILE="coverage.filtered.out"
  cp coverage.out "$FILTERED_FILE"

  # Build list of patterns to exclude
  EXCLUDE_PATTERNS="${COVERAGE_EXCLUDE:-}"
  if [ -f .coverignore ]; then
    while IFS= read -r line; do
      case "$line" in
        ''|\#*) continue ;;
        *) EXCLUDE_PATTERNS+=" $line" ;;
      esac
    done < .coverignore
  fi

  # Apply exclusions by removing matching lines from the coverprofile
  for pat in $EXCLUDE_PATTERNS; do
    awk -v pat="$pat" 'NR==1 || index($0, pat)==0' "$FILTERED_FILE" > "$FILTERED_FILE.tmp" && mv "$FILTERED_FILE.tmp" "$FILTERED_FILE"
  done
fi

TOTAL_LINE=$(go tool cover -func="$FILTERED_FILE" | tail -n 1)
PCT=$(echo "$TOTAL_LINE" | awk '{print $3}' | tr -d '%')
echo "Total coverage: ${PCT}% (threshold ${THRESHOLD}%)"
PCT_INT=${PCT%.*}
if [ "$PCT_INT" -lt "$THRESHOLD" ]; then
  echo "Coverage below threshold" >&2
  exit 1
fi
exit 0
