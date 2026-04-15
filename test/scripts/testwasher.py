#!/usr/bin/env python3
"""
Minimal TestWasher for helm-charts E2E tests.

Reads 'go test -json' output from stdin, prints human-readable output,
and exits 0 if all test failures are known-flaky (logged the flake sentinel).
Exits 1 if any non-flaky test failures or build failures are found.

Ported from github.com/DataDog/datadog-agent/tasks/testwasher.py
"""
import json
import sys

FLAKY_SENTINEL = "flakytest: this is a known flaky test"


def main():
    flaky_tests = set()   # (package, test) that logged the sentinel
    failed_tests = set()  # (package, test) that received a "fail" action
    build_failed = False

    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            # Non-JSON output (e.g. build errors before JSON starts) — print as-is
            print(line, flush=True)
            continue

        action = event.get("Action", "")
        package = event.get("Package", "")
        test = event.get("Test", "")
        output = event.get("Output", "")

        # Forward human-readable output to stdout for CI log visibility
        if action == "output":
            sys.stdout.write(output)
            sys.stdout.flush()

        if action == "build-fail":
            build_failed = True

        # Package-level events (no Test field) are not individual test results
        if not test:
            continue

        key = (package, test)

        if action == "output" and FLAKY_SENTINEL in output:
            flaky_tests.add(key)
        elif action == "fail":
            failed_tests.add(key)

    # Build failures are always fatal — no test ran to produce flake markers
    if build_failed:
        print("\nFAIL: build failed.", file=sys.stderr)
        sys.exit(1)

    non_flaky_failures = failed_tests - flaky_tests

    if non_flaky_failures:
        print(f"\nFAIL: {len(non_flaky_failures)} non-flaky test failure(s):", file=sys.stderr)
        for pkg, test in sorted(non_flaky_failures):
            print(f"  {pkg} {test}", file=sys.stderr)
        sys.exit(1)

    if failed_tests:
        print(f"\nPASS (known-flaky failures ignored): {len(failed_tests)} flaky test(s) failed:")
        for pkg, test in sorted(failed_tests):
            print(f"  {pkg} {test}")

    sys.exit(0)


if __name__ == "__main__":
    main()
