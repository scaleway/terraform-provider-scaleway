#!/usr/bin/env bash
set -euo pipefail

# Services intentionally excluded from acceptance tests (no acceptance tests)
EXCLUDED_SERVICES=(
  scwconfig
)

SERVICES_DIR="internal/services"
NIGHTLY_WORKFLOW=".github/workflows/nightly.yml"
ACCEPTANCE_WORKFLOW=".github/workflows/acceptance-tests.yaml"

is_excluded() {
  local service="$1"
  for excluded in "${EXCLUDED_SERVICES[@]}"; do
    [[ "$service" == "$excluded" ]] && return 0
  done
  return 1
}

# Extract all expected services from the services directory
get_all_services() {
  for service_path in "$SERVICES_DIR"/*/; do
    service=$(basename "$service_path")
    if ! is_excluded "$service"; then
      echo "$service"
    fi
  done
}

# Check a matrix for missing services
# Arguments: workflow_file matrix_services (newline-separated) matrix_name
check_matrix() {
  local workflow="$1"
  local matrix_services="$2"
  local matrix_name="$3"

  missing=()
  while IFS= read -r service; do
    if ! echo "$matrix_services" | grep -qx "$service"; then
      missing+=("$service")
    fi
  done < <(get_all_services)

  if [[ ${#missing[@]} -gt 0 ]]; then
    echo "❌ The following services are missing from the '$matrix_name' matrix in $workflow:"
    for s in "${missing[@]}"; do
      echo "  - $s"
    done
    echo ""
    echo "Add them to the 'products' matrix or to the EXCLUDED_SERVICES list in this script."
    return 1
  fi

  echo "✅ All services are present in the '$matrix_name' matrix."
  return 0
}

exit_code=0

# Check nightly workflow (single products matrix)
nightly_services=$(grep -E '^          - [a-z]' "$NIGHTLY_WORKFLOW" | sed 's/.*- //' | tr -d ' ')
check_matrix "$NIGHTLY_WORKFLOW" "$nightly_services" "nightly" || exit_code=1

# Check acceptance tests workflow - terraform job
terraform_services=$(sed -n '/^  terraform:/,/^  [a-z]/p' "$ACCEPTANCE_WORKFLOW" | grep -E '^          - [a-z]' | sed 's/.*- //' | tr -d ' ')
check_matrix "$ACCEPTANCE_WORKFLOW" "$terraform_services" "terraform" || exit_code=1

# Check acceptance tests workflow - opentofu job
opentofu_services=$(sed -n '/^  opentofu:/,/^  [a-z]/p' "$ACCEPTANCE_WORKFLOW" | grep -E '^          - [a-z]' | sed 's/.*- //' | tr -d ' ')
check_matrix "$ACCEPTANCE_WORKFLOW" "$opentofu_services" "opentofu" || exit_code=1

if [[ $exit_code -eq 0 ]]; then
  echo "✅ All services are present in all workflow matrices."
fi

exit $exit_code
