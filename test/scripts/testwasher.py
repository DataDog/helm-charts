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
    flaky_tests = set()       # (package, test) that logged the sentinel
    failed_tests = set()     # (package, test) that received a "fail" action
    tested_packages = set()  # packages that had at least one test-level event
    failed_packages = set()  # packages that had a package-level "fail" (no Test)
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

        # Package-level events (no Test field) are not individual test results.
        # Track package-level fails separately — they are only fatal if the
        # package had no individual test events (indicating an early crash).
        if not test:
            if action == "fail" and package:
                failed_packages.add(package)
            continue

        tested_packages.add(package)

        key = (package, test)

        if action == "output" and FLAKY_SENTINEL in output:
            flaky_tests.add(key)
        elif action == "fail":
            failed_tests.add(key)

    # Build failures are always fatal
    if build_failed:
        print("\nFAIL: build failed.", file=sys.stderr)
        sys.exit(1)

    # A package-level "fail" with no individual test events means the test
    # binary crashed before running tests — treat as fatal.
    crashed_packages = failed_packages - tested_packages
    if crashed_packages:
        print(f"\nFAIL: package(s) failed without running tests:", file=sys.stderr)
        for pkg in sorted(crashed_packages):
            print(f"  {pkg}", file=sys.stderr)
        sys.exit(1)

    # A subtest (e.g. TestFoo/Bar) inherits its parent's flaky marker.
    # Check each failure against both exact match and parent test name.
    def is_flaky(pkg, test_name):
        if (pkg, test_name) in flaky_tests:
            return True
        # Check if a parent test was marked flaky (TestFoo covers TestFoo/Bar)
        if "/" in test_name:
            parent = test_name.split("/")[0]
            if (pkg, parent) in flaky_tests:
                return True
        return False

    # First pass: identify directly flaky failures (marked or inherited from parent)
    non_flaky_failures = {(p, t) for p, t in failed_tests if not is_flaky(p, t)}

    # Second pass: a parent test that only failed because its subtests failed
    # should not count as non-flaky if all its failing subtests are flaky.
    # e.g. TestFoo fails because TestFoo/Bar (flaky) failed — TestFoo is not
    # itself broken.
    parents_to_remove = set()
    for pkg, test_name in list(non_flaky_failures):
        if "/" in test_name:
            continue  # only check top-level tests
        # Find all failing subtests of this parent
        failing_children = {
            (p, t) for p, t in failed_tests
            if p == pkg and t.startswith(test_name + "/")
        }
        if not failing_children:
            continue  # parent failed on its own, not from subtests
        # If all failing children are flaky, the parent failure is just propagation
        if all(is_flaky(p, t) for p, t in failing_children):
            parents_to_remove.add((pkg, test_name))
    non_flaky_failures -= parents_to_remove

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
