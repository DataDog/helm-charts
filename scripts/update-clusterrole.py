#!/usr/bin/env python3
"""
Update a datadog-operator ClusterRole template from an upstream operator role.

Downloads the upstream config/rbac/role.yaml, filters out resources that are
handled by conditional blocks in the helm chart, adds template directives,
and replaces the lines between the BEGIN and END AUTO-GENERATED RULES markers
of clusterrole.yaml. Everything outside the markers (the ClusterRole header,
hand-maintained rules and conditional blocks) is preserved verbatim.

Usage:
    python3 update-clusterrole.py <upstream_role_yaml> <clusterrole_yaml>
"""

import copy
import sys

import yaml

# Resources handled by conditional blocks in the helm chart template.
# These are excluded from the main section and added back conditionally
# via {{- if .Values.<feature>.enabled }} blocks.
#
# Each entry is (apiGroup, resource).
#
# When a new operator feature adds RBAC resources that should be opt-in,
# add the (apiGroup, resource) pairs here AND add a new conditional block
# to the template.
#
# If stripping removes all resources from a rule, the entire rule is skipped.
CONDITIONAL_RESOURCES = {
    # clusterRole.allowCreatePodsExec
    ("", "pods/exec"),
    # datadogAgentInternal.enabled / datadogAgentProfile.enabled
    ("datadoghq.com", "datadogagentinternals"),
    ("datadoghq.com", "datadogagentinternals/finalizers"),
    ("datadoghq.com", "datadogagentinternals/status"),
    # datadogAgentProfile.enabled
    ("datadoghq.com", "datadogagentprofiles"),
    ("datadoghq.com", "datadogagentprofiles/finalizers"),
    ("datadoghq.com", "datadogagentprofiles/status"),
    # datadogDashboard.enabled
    ("datadoghq.com", "datadogdashboards"),
    ("datadoghq.com", "datadogdashboards/finalizers"),
    ("datadoghq.com", "datadogdashboards/status"),
    # datadogCSIDriver.enabled
    ("datadoghq.com", "datadogcsidrivers"),
    ("datadoghq.com", "datadogcsidrivers/finalizers"),
    ("datadoghq.com", "datadogcsidrivers/status"),
    ("storage.k8s.io", "csidrivers"),
}


# Only the lines between the lines containing these markers are regenerated.
BEGIN_MARKER = "BEGIN AUTO-GENERATED RULES"
END_MARKER = "END AUTO-GENERATED RULES"


def quote_val(v):
    """Quote special YAML values."""
    if v == "*":
        return "'*'"
    if v.startswith("*") and "/" in v:
        return "'" + v + "'"
    if v == "":
        return '""'
    return v


def format_rule(rule):
    """Format a single RBAC rule as YAML text lines."""
    lines = []
    if "nonResourceURLs" in rule:
        lines.append("- nonResourceURLs:")
        for u in rule["nonResourceURLs"]:
            lines.append("  - " + u)
    if "apiGroups" in rule:
        lines.append("- apiGroups:")
        for g in rule["apiGroups"]:
            lines.append("  - " + quote_val(g))
    if "resourceNames" in rule:
        lines.append("  resourceNames:")
        for n in rule["resourceNames"]:
            lines.append("  - " + n)
    if "resources" in rule:
        lines.append("  resources:")
        for res in rule["resources"]:
            lines.append("  - " + quote_val(res))
    lines.append("  verbs:")
    for v in rule["verbs"]:
        lines.append("  - " + quote_val(v))
    return lines


def filter_rules(rules):
    """Filter out rules/resources handled by conditional blocks."""
    filtered = []
    for rule in rules:
        r = copy.deepcopy(rule)
        api_groups = r.get("apiGroups", [])
        resources = r.get("resources", [])
        verbs = r.get("verbs", [])

        # Remove 'patch' from nodes verbs (handled by datadogAgentProfile)
        if api_groups == [""] and resources == ["nodes"]:
            r["verbs"] = [v for v in verbs if v != "patch"]

        # Strip conditional resources; skip rule entirely if nothing remains
        if resources:
            new_res = [
                res for res in resources
                if not any((grp, res) in CONDITIONAL_RESOURCES for grp in api_groups)
            ]
            # All resources in this rule are conditional; skip the entire rule
            if len(new_res) == 0:
                continue
            if len(new_res) != len(resources):
                r["resources"] = new_res

        filtered.append(r)
    return filtered


def add_template_directives(lines):
    """Insert helm template conditionals into the rules text."""
    result = []
    for line in lines:
        if line == "  - nodes/proxy":
            result.append("  {{- if not .Values.clusterRole.kubeletFineGrainedAuthorization }}")
            result.append(line)
            result.append("  {{- end }}")
        else:
            result.append(line)
    return result


def find_marker(lines, marker, clusterrole_file):
    """Return the index of the only line containing marker."""
    found = [i for i, line in enumerate(lines) if marker in line]
    if len(found) != 1:
        sys.exit(f"error: expected one '{marker}' marker in {clusterrole_file}, found {len(found)}")
    return found[0]


def split_template(clusterrole_file):
    """Return the lines through the BEGIN marker and the lines from the END marker onward."""
    with open(clusterrole_file) as f:
        lines = f.readlines()

    begin = find_marker(lines, BEGIN_MARKER, clusterrole_file)
    end = find_marker(lines, END_MARKER, clusterrole_file)
    if end < begin:
        sys.exit(f"error: '{END_MARKER}' comes before '{BEGIN_MARKER}' in {clusterrole_file}")

    return "".join(lines[:begin + 1]), "".join(lines[end:])


def main():
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <upstream_role_yaml> <clusterrole_yaml>", file=sys.stderr)
        sys.exit(1)

    upstream_file = sys.argv[1]
    clusterrole_file = sys.argv[2]

    with open(upstream_file) as f:
        upstream = yaml.safe_load(f)

    rules = filter_rules(upstream["rules"])

    main_lines = []
    for rule in rules:
        main_lines.extend(format_rule(rule))

    main_lines = add_template_directives(main_lines)

    preserved_head, preserved_tail = split_template(clusterrole_file)

    with open(clusterrole_file, "w") as f:
        f.write(preserved_head)
        f.write("\n".join(main_lines))
        f.write("\n")
        f.write(preserved_tail)


if __name__ == "__main__":
    main()
