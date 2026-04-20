#!/usr/bin/env python3
"""
Update the datadog-operator clusterrole.yaml from upstream operator role.yaml.

Downloads the upstream config/rbac/role.yaml, filters out resources that are
handled by conditional blocks in the helm chart, adds template directives,
and reconstructs clusterrole.yaml preserving the header and conditional blocks.

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
# When a new operator feature adds RBAC resources that should be opt-in,
# add the resources here AND add a new conditional block to the template.
CONDITIONAL_RESOURCES = {
    # datadogAgentInternal.enabled / datadogAgentProfile.enabled
    "datadogagentinternals",
    "datadogagentinternals/finalizers",
    "datadogagentinternals/status",
    # datadogAgentProfile.enabled
    "datadogagentprofiles",
    "datadogagentprofiles/finalizers",
    "datadogagentprofiles/status",
    # datadogDashboard.enabled
    "datadogdashboards",
    "datadogdashboards/finalizers",
    "datadogdashboards/status",
    # datadogCSIDriver.enabled
    "datadogcsidrivers",
    "datadogcsidrivers/finalizers",
    "datadogcsidrivers/status",
}

# Entire rules to skip (resource lists that are handled by conditional blocks).
# Each entry is (apiGroups, resources) — both as frozensets.
CONDITIONAL_RULES = [
    # clusterRole.allowCreatePodsExec
    (frozenset([""]), frozenset(["pods/exec"])),
    # datadogCSIDriver.enabled — storage.k8s.io csidrivers rules
    (frozenset(["storage.k8s.io"]), frozenset(["csidrivers"])),
]

# Resources to strip from combined rules (when mixed with non-conditional resources).
STRIP_FROM_COMBINED = {"pods/exec"}


HEADER = """\
{{- if .Values.rbac.create -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ include "datadog-operator.fullname" . }}
  labels:
{{ include "datadog-operator.labels" . | indent 4 }}
rules:
"""


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

        # Skip entire rules that match conditional block patterns
        groups_set = frozenset(api_groups)
        res_set = frozenset(resources)
        skip = False
        for cond_groups, cond_res in CONDITIONAL_RULES:
            if groups_set == cond_groups and res_set == cond_res:
                skip = True
                break
        if skip:
            continue

        # Remove 'patch' from nodes verbs (handled by datadogAgentProfile)
        if api_groups == [""] and resources == ["nodes"]:
            r["verbs"] = [v for v in verbs if v != "patch"]

        # Strip conditional resources from combined rules
        if resources:
            new_res = [x for x in resources if x not in CONDITIONAL_RESOURCES and x not in STRIP_FROM_COMBINED]
            if not new_res:
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


def extract_conditional_blocks(clusterrole_file):
    """Extract the conditional blocks from the current clusterrole.yaml."""
    with open(clusterrole_file) as f:
        lines = f.readlines()

    # Find where conditional blocks start: first {{- if .Values. after the rules header
    for i, line in enumerate(lines):
        if "{{- if .Values." in line and i > 7:
            return "".join(lines[i:])
    return ""


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

    conditional_blocks = extract_conditional_blocks(clusterrole_file)

    with open(clusterrole_file, "w") as f:
        f.write(HEADER)
        f.write("\n".join(main_lines))
        f.write("\n")
        f.write(conditional_blocks)


if __name__ == "__main__":
    main()
