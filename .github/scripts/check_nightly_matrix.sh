#!/usr/bin/env bash
set -euo pipefail

# Services intentionally excluded from nightly tests (no acceptance tests)
EXCLUDED_SERVICES=(
  scwconfig
)

SERVICES_DIR="internal/services"
NIGHTLY_WORKFLOW=".github/workflows/nightly.yml"

is_excluded() {
  local service="$1"
  for excluded in "${EXCLUDED_SERVICES[@]}"; do
    [[ "$service" == "$excluded" ]] && return 0
  done
  return 1
}

matrix_services=$(grep -E '^\s+- [a-z]' "$NIGHTLY_WORKFLOW" | sed 's/.*- //' | tr -d ' ')

missing=()
for service_path in "$SERVICES_DIR"/*/; do
  service=$(basename "$service_path")
  if is_excluded "$service"; then
    continue
  fi
  if ! echo "$matrix_services" | grep -qx "$service"; then
    missing+=("$service")
  fi
done

if [[ ${#missing[@]} -gt 0 ]]; then
  echo "❌ The following services are missing from the nightly matrix in $NIGHTLY_WORKFLOW:"
  for s in "${missing[@]}"; do
    echo "  - $s"
  done
  echo ""
  echo "Add them to the 'products' matrix or to the EXCLUDED_SERVICES list in this script."
  exit 1
fi

echo "✅ All services are present in the nightly matrix."
