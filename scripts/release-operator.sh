#!/usr/bin/env bash
#
# release-operator.sh - Automate the Datadog Operator release process for helm-charts.
#
# This script automates the "Public Helm Charts" section of the Operator Release
# Manager Duties. It handles three phases (typically separate PRs):
#
#   1. crds     - Update datadog-crds chart with new CRDs
#   2. operator - Update datadog-operator chart
#   3. datadog  - Update datadog chart operator dependency
#
# See: https://datadoghq.atlassian.net/wiki/spaces/CONTP/pages/2169208977
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

info()    { echo "[INFO] $*"; }
warn()    { echo "[WARN] $*"; }
success() { echo "[DONE] $*"; }
error()   { echo "[ERROR] $*" >&2; exit 1; }
step()    { echo "  -> $*"; }

# --- Usage ---
usage() {
    cat <<EOF
Usage: $(basename "$0") <phase> --operator-version <version> [options]

Automate the Datadog Operator release process for public helm charts.

Phases:
  crds        Update the datadog-crds chart with new CRDs
  operator    Update the datadog-operator chart
  datadog     Update the datadog chart operator dependencies
Required:
  --operator-version VERSION     Operator version (e.g., 1.26.0-rc.1 or 1.26.0)

Optional:
  --crds-chart-version VERSION        Override CRDs chart version (auto-calculated)
  --operator-chart-version VERSION    Override operator chart version (auto-calculated)
  --datadog-chart-version VERSION     Override datadog chart version (auto-calculated)
  -h, --help                          Show this help

Examples:
  # First release candidate (creates -dev.1 chart versions)
  $(basename "$0") crds     --operator-version 1.26.0-rc.1
  $(basename "$0") operator --operator-version 1.26.0-rc.1
  $(basename "$0") datadog  --operator-version 1.26.0-rc.1

  # Subsequent release candidate (increments dev suffix)
  $(basename "$0") operator --operator-version 1.26.0-rc.2

  # Final release (strips dev suffix)
  $(basename "$0") crds     --operator-version 1.26.0
  $(basename "$0") operator --operator-version 1.26.0
  $(basename "$0") datadog  --operator-version 1.26.0

Version auto-calculation:
  RC release:    current 2.21.0       -> 2.22.0-dev.1 (first RC)
                 current 2.22.0-dev.1 -> 2.22.0-dev.2 (subsequent RC)
  Final release: current 2.22.0-dev.2 -> 2.22.0       (strip suffix)
                 current 2.22.0       -> 2.23.0       (bump minor)
EOF
    exit 0
}

# ============================================================================
# Version helpers
# ============================================================================

is_rc() { [[ "$1" == *"-rc."* ]]; }
is_dev() { [[ "$1" == *"-dev."* ]]; }

# Calculate the next chart version based on current version and release type.
# Args: current_version, release_type (rc|final)
calc_next_version() {
    local current="$1" release_type="$2"

    if [[ "$release_type" == "rc" ]]; then
        if is_dev "$current"; then
            # Increment dev number: 2.22.0-dev.1 -> 2.22.0-dev.2
            local base dev_num
            base="${current%-dev.*}"
            dev_num="${current##*-dev.}"
            echo "${base}-dev.$((dev_num + 1))"
        else
            # Bump minor, add -dev.1: 2.21.0 -> 2.22.0-dev.1
            local major minor
            major="${current%%.*}"
            minor="${current#*.}"
            minor="${minor%%.*}"
            echo "${major}.$((minor + 1)).0-dev.1"
        fi
    else
        # Final release
        if is_dev "$current"; then
            # Strip dev suffix: 2.22.0-dev.2 -> 2.22.0
            echo "${current%-dev.*}"
        else
            # Bump minor: 2.21.0 -> 2.22.0 (only datadog chart, which doesn't recieve any -dev)
            local major minor
            major="${current%%.*}"
            minor="${current#*.}"
            minor="${minor%%.*}"
            echo "${major}.$((minor + 1)).0"
        fi
    fi
}

# Read the 'version:' field from a Chart.yaml (top-level only).
get_chart_version() {
    grep '^version:' "$1" | head -1 | awk '{print $2}'
}

# Read the 'appVersion:' field from a Chart.yaml.
get_app_version() {
    grep '^appVersion:' "$1" | head -1 | awk '{print $2}' | tr -d '"'
}

# ============================================================================
# File editing helpers
# ============================================================================

# Portable in-place sed (macOS and Linux).
sed_i() {
    if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' "$@"
    else
        sed -i "$@"
    fi
}

# Add a changelog entry after the first heading line.
# Args: file, version, entry_line1 [entry_line2 ...]
add_changelog_entry() {
    local file="$1" version="$2"
    shift 2
    local entries=("$@")

    local tmp="${file}.tmp"
    {
        head -1 "$file"
        echo ""
        echo "## ${version}"
        echo ""
        for entry in "${entries[@]}"; do
            echo "$entry"
        done
        tail -n +2 "$file"
    } > "$tmp"
    mv "$tmp" "$file"
}

# Update a dependency version in a YAML file.
# Finds 'name: <dep_name>' and updates the next 'version:' line.
# Args: file, dep_name, new_version
update_dep_version() {
    local file="$1" dep_name="$2" new_version="$3"
    awk -v name="$dep_name" -v ver="$new_version" '
        $0 ~ "name: " name { found=1 }
        found && /version:/ { sub(/version: .*/, "version: " ver); found=0 }
        { print }
    ' "$file" > "${file}.tmp"
    mv "${file}.tmp" "$file"
}

# ============================================================================
# Helm docs
# ============================================================================

run_helm_docs() {
    local helm_docs_bin=""

    if command -v helm-docs &>/dev/null; then
        helm_docs_bin="helm-docs"
    else
        local version="1.14.2" os arch
        os=$(uname)
        arch=$(uname -m)
        info "Downloading helm-docs v${version}..."
        curl --silent --show-error --fail --location \
            --output /tmp/helm-docs.tar.gz \
            "https://github.com/norwoodj/helm-docs/releases/download/v${version}/helm-docs_${version}_${os}_${arch}.tar.gz"
        tar -xf /tmp/helm-docs.tar.gz -C /tmp helm-docs
        helm_docs_bin="/tmp/helm-docs"
    fi

    (cd "$ROOT_DIR" && "$helm_docs_bin")
    success "READMEs updated with helm-docs"
}

# ============================================================================
# RBAC / ClusterRole update
# ============================================================================

# Update clusterrole.yaml from upstream operator role.yaml.
update_clusterrole() {
    local new_version="$1"
    local clusterrole="$ROOT_DIR/charts/datadog-operator/templates/clusterrole.yaml"

    local upstream_url="https://raw.githubusercontent.com/DataDog/datadog-operator/v${new_version}/config/rbac/role.yaml"
    local upstream_file
    upstream_file=$(mktemp)

    if ! curl --silent --show-error --fail --location -o "$upstream_file" "$upstream_url" 2>/dev/null; then
        rm -f "$upstream_file"
        warn "Could not fetch upstream role.yaml for v${new_version}. Skipping clusterrole update."
        warn "Check manually: https://github.com/DataDog/datadog-operator/blob/v${new_version}/config/rbac/role.yaml"
        return 1
    fi

    if python3 "$SCRIPT_DIR/update-clusterrole.py" "$upstream_file" "$clusterrole"; then
        rm -f "$upstream_file"
        success "clusterrole.yaml updated from upstream v${new_version}"
    else
        rm -f "$upstream_file"
        warn "Failed to update clusterrole.yaml automatically"
        return 1
    fi
}

# ============================================================================
# Phase: CRDs
# ============================================================================

phase_crds() {
    info "=== Phase: CRDs (datadog-crds chart) ==="

    local chart_dir="$ROOT_DIR/charts/datadog-crds"
    local chart_yaml="$chart_dir/Chart.yaml"
    local changelog="$chart_dir/CHANGELOG.md"

    # Step 1: Run update-crds.sh
    step "Running update-crds.sh v${OPERATOR_VERSION}..."
    (cd "$ROOT_DIR" && bash "$chart_dir/update-crds.sh" "v${OPERATOR_VERSION}")
    success "CRDs downloaded and formatted"

    # During RC, if update-crds.sh produced no changes, skip the rest —
    # a previous RC already has the CRDs we need.
    if is_rc "$OPERATOR_VERSION" && git diff --quiet -- "$chart_dir"; then
        info "No CRD changes detected for this RC, skipping CRDs chart bump."
        return 0
    fi

    # Step 2: Bump version in Chart.yaml
    step "Updating Chart.yaml version: $(get_chart_version "$chart_yaml") -> ${CRDS_CHART_VERSION}"
    sed_i "s/^version: .*/version: ${CRDS_CHART_VERSION}/" "$chart_yaml"

    # Step 3: Add CHANGELOG entry
    step "Adding CHANGELOG entry for ${CRDS_CHART_VERSION}"
    local msg
    if is_rc "$OPERATOR_VERSION"; then
        msg="* Update CRDs from Datadog Operator v${OPERATOR_VERSION} release candidate tag."
    else
        msg="* Update CRDs from Datadog Operator v${OPERATOR_VERSION}."
    fi
    add_changelog_entry "$changelog" "$CRDS_CHART_VERSION" "$msg"

    # Step 4: Run helm-docs
    step "Running helm-docs..."
    run_helm_docs

    success "=== CRDs phase complete ==="
}

# ============================================================================
# Phase: Operator Chart
# ============================================================================

phase_operator() {
    info "=== Phase: Operator Chart (datadog-operator) ==="

    local chart_dir="$ROOT_DIR/charts/datadog-operator"
    local chart_yaml="$chart_dir/Chart.yaml"
    local changelog="$chart_dir/CHANGELOG.md"
    local values_yaml="$chart_dir/values.yaml"
    local helpers_tpl="$chart_dir/templates/_helpers.tpl"
    local test_file="$ROOT_DIR/test/datadog-operator/operator_deployment_test.go"

    # Get current appVersion for replacement references
    local prev_version
    prev_version=$(get_app_version "$chart_yaml")
    info "Previous operator version: ${prev_version}"

    # Step 1: Update CRDs dependency version
    step "Updating datadog-crds dependency to ${CRDS_CHART_VERSION}"
    update_dep_version "$chart_yaml" "datadog-crds" "$CRDS_CHART_VERSION"

    # Step 2: Helm dependency update
    step "Running helm dependency update..."
    (cd "$chart_dir" && helm dependency update 2>/dev/null)
    success "Chart.lock updated"

    # Step 3: Bump chart version
    step "Updating chart version: $(get_chart_version "$chart_yaml") -> ${OPERATOR_CHART_VERSION}"
    sed_i "s/^version: .*/version: ${OPERATOR_CHART_VERSION}/" "$chart_yaml"

    # Step 4: Update appVersion
    step "Updating appVersion: ${prev_version} -> ${OPERATOR_VERSION}"
    sed_i "s/^appVersion: .*/appVersion: ${OPERATOR_VERSION}/" "$chart_yaml"

    # Step 5: Update _helpers.tpl fallback tag
    step "Updating _helpers.tpl check-image-tag fallback"
    sed_i "s/\"${prev_version}\"/\"${OPERATOR_VERSION}\"/" "$helpers_tpl"

    # Step 6: Update values.yaml image.tag
    step "Updating values.yaml image.tag: ${prev_version} -> ${OPERATOR_VERSION}"
    sed_i "s/^  tag: ${prev_version}$/  tag: ${OPERATOR_VERSION}/" "$values_yaml"

    # Step 7: Add CHANGELOG entry
    step "Adding CHANGELOG entry for ${OPERATOR_CHART_VERSION}"
    add_changelog_entry "$changelog" "$OPERATOR_CHART_VERSION" \
        "* Update Datadog Operator chart for ${OPERATOR_VERSION}."

    # Step 8: Update test assertion
    step "Updating operator_deployment_test.go image assertion"
    sed_i "s|operator:${prev_version}|operator:${OPERATOR_VERSION}|g" "$test_file"

    # Step 9: Run helm-docs
    step "Running helm-docs..."
    run_helm_docs

    # Step 10: Update test baselines
    step "Updating test baselines (make update-test-baselines-operator)..."
    (cd "$ROOT_DIR" && make update-test-baselines-operator)
    success "Test baselines updated"

    # Step 11: Update clusterrole.yaml from upstream RBAC
    step "Updating clusterrole.yaml from upstream v${OPERATOR_VERSION}..."
    update_clusterrole "$OPERATOR_VERSION"

    success "=== Operator chart phase complete ==="
}

# ============================================================================
# Phase: Datadog Chart
# ============================================================================

phase_datadog() {
    info "=== Phase: Datadog Chart (dependency update) ==="

    local chart_dir="$ROOT_DIR/charts/datadog"
    local chart_yaml="$chart_dir/Chart.yaml"
    local changelog="$chart_dir/CHANGELOG.md"
    local values_yaml="$chart_dir/values.yaml"
    local requirements="$chart_dir/requirements.yaml"

    # Get previous operator tag from values.yaml
    local prev_tag
    prev_tag=$(awk '/^operator:/,0' "$values_yaml" | grep '^\s*tag:' | head -1 | awk '{print $2}')
    info "Previous operator image tag in datadog chart: ${prev_tag}"

    # Step 1: Update datadog-operator dependency in requirements.yaml
    step "Updating datadog-operator dependency to ${OPERATOR_CHART_VERSION}"
    update_dep_version "$requirements" "datadog-operator" "$OPERATOR_CHART_VERSION"

    # Step 2: Update datadog-crds dependency in requirements.yaml
    step "Updating datadog-crds dependency to ${CRDS_CHART_VERSION}"
    update_dep_version "$requirements" "datadog-crds" "$CRDS_CHART_VERSION"

    # Step 3: Helm dependency update
    step "Running helm dependency update..."
    (cd "$chart_dir" && helm dependency update 2>/dev/null)
    success "requirements.lock updated"

    # Step 4: Update operator.image.tag in values.yaml
    step "Updating operator.image.tag: ${prev_tag} -> ${OPERATOR_VERSION}"
    awk -v old_tag="$prev_tag" -v new_tag="$OPERATOR_VERSION" '
        /^operator:/ { in_operator=1 }
        in_operator && /^[^ ]/ && !/^operator:/ { in_operator=0 }
        in_operator && /tag:/ && $2 == old_tag { sub(old_tag, new_tag) }
        { print }
    ' "$values_yaml" > "${values_yaml}.tmp"
    mv "${values_yaml}.tmp" "$values_yaml"

    # Step 5: Bump datadog chart version
    step "Updating chart version: $(get_chart_version "$chart_yaml") -> ${DATADOG_CHART_VERSION}"
    sed_i "s/^version: .*/version: ${DATADOG_CHART_VERSION}/" "$chart_yaml"

    # Step 6: Add CHANGELOG entry
    step "Adding CHANGELOG entry for ${DATADOG_CHART_VERSION}"
    add_changelog_entry "$changelog" "$DATADOG_CHART_VERSION" \
        "* Bump Datadog Operator chart dependency to ${OPERATOR_CHART_VERSION}." \
        "* Bump Datadog CRD chart dependency to ${CRDS_CHART_VERSION}." \
        "* Bump Operator image tag to ${OPERATOR_VERSION}."

    # Step 7: Run helm-docs
    step "Running helm-docs..."
    run_helm_docs

    # Step 8: Update test baselines
    step "Updating test baselines (make update-test-baselines-datadog-agent)..."
    (cd "$ROOT_DIR" && make update-test-baselines-datadog-agent)
    success "Test baselines updated"

    success "=== Datadog chart phase complete ==="
}

# ============================================================================
# Prerequisites check
# ============================================================================

check_prereqs() {
    local missing=()

    command -v helm &>/dev/null    || missing+=("helm")
    command -v curl &>/dev/null    || missing+=("curl")
    command -v yq &>/dev/null      || missing+=("yq (required by update-crds.sh)")
    command -v awk &>/dev/null     || missing+=("awk")
    command -v python3 &>/dev/null || missing+=("python3 (required for clusterrole update)")

    if [[ ${#missing[@]} -gt 0 ]]; then
        error "Missing required tools: ${missing[*]}"
    fi

    # Verify we're in the helm-charts repo
    if [[ ! -f "$ROOT_DIR/charts/datadog-operator/Chart.yaml" ]]; then
        error "Must be run from the helm-charts repository (or scripts/ inside it)"
    fi
}

# ============================================================================
# Main
# ============================================================================

PHASE=""
OPERATOR_VERSION=""
CRDS_CHART_VERSION=""
OPERATOR_CHART_VERSION=""
DATADOG_CHART_VERSION=""

[[ $# -eq 0 ]] && usage

PHASE="$1"
shift

while [[ $# -gt 0 ]]; do
    case "$1" in
        --operator-version)       OPERATOR_VERSION="$2"; shift 2 ;;
        --crds-chart-version)     CRDS_CHART_VERSION="$2"; shift 2 ;;
        --operator-chart-version) OPERATOR_CHART_VERSION="$2"; shift 2 ;;
        --datadog-chart-version)  DATADOG_CHART_VERSION="$2"; shift 2 ;;
        -h|--help)                usage ;;
        *)                        error "Unknown option: $1" ;;
    esac
done

[[ -z "$OPERATOR_VERSION" ]] && error "--operator-version is required"

check_prereqs

# Determine release type
RELEASE_TYPE="final"
is_rc "$OPERATOR_VERSION" && RELEASE_TYPE="rc"

echo ""
info "Operator version:  ${OPERATOR_VERSION}"
info "Release type:      ${RELEASE_TYPE}"

# Auto-calculate chart versions.
# Each phase only bumps its own chart version; other versions default to current.
if [[ -z "$CRDS_CHART_VERSION" ]]; then
    current=$(get_chart_version "$ROOT_DIR/charts/datadog-crds/Chart.yaml")
    if [[ "$PHASE" == "crds" ]]; then
        CRDS_CHART_VERSION=$(calc_next_version "$current" "$RELEASE_TYPE")
        info "CRDs chart:        ${current} -> ${CRDS_CHART_VERSION} (auto)"
    else
        CRDS_CHART_VERSION="$current"
        info "CRDs chart:        ${CRDS_CHART_VERSION} (current)"
    fi
else
    info "CRDs chart:        ${CRDS_CHART_VERSION} (manual)"
fi

if [[ -z "$OPERATOR_CHART_VERSION" ]]; then
    current=$(get_chart_version "$ROOT_DIR/charts/datadog-operator/Chart.yaml")
    if [[ "$PHASE" == "operator" ]]; then
        OPERATOR_CHART_VERSION=$(calc_next_version "$current" "$RELEASE_TYPE")
        info "Operator chart:    ${current} -> ${OPERATOR_CHART_VERSION} (auto)"
    else
        OPERATOR_CHART_VERSION="$current"
        info "Operator chart:    ${OPERATOR_CHART_VERSION} (current)"
    fi
else
    info "Operator chart:    ${OPERATOR_CHART_VERSION} (manual)"
fi

if [[ -z "$DATADOG_CHART_VERSION" ]]; then
    current=$(get_chart_version "$ROOT_DIR/charts/datadog/Chart.yaml")
    if [[ "$PHASE" == "datadog" ]]; then
        DATADOG_CHART_VERSION=$(calc_next_version "$current" "$RELEASE_TYPE")
        info "Datadog chart:     ${current} -> ${DATADOG_CHART_VERSION} (auto)"
    else
        DATADOG_CHART_VERSION="$current"
        info "Datadog chart:     ${DATADOG_CHART_VERSION} (current)"
    fi
else
    info "Datadog chart:     ${DATADOG_CHART_VERSION} (manual)"
fi

echo ""

# Execute phase(s)
case "$PHASE" in
    crds)
        phase_crds
        ;;
    operator)
        phase_operator
        ;;
    datadog)
        phase_datadog
        ;;
    -h|--help)
        usage
        ;;
    *)
        error "Unknown phase: $PHASE (use: crds, operator, datadog)"
        ;;
esac

echo ""
success "Release script complete!"
echo ""
info "Next steps:"
info "  1. Review changes: git diff"
info "  2. Create a branch, commit, and open a PR"
