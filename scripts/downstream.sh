#!/usr/bin/env bash
# Checks that this GOE working tree doesn't break the apps that depend on it.
#
# Each module is built, vetted and tested twice: once as it is, against the GOE
# version its own go.mod requires, and once as if it had upgraded to this
# working tree: a copy of its go.mod, with a replace pointing GOE here and then
# tidied, is used through -modfile. That resolves dependencies the way the real
# upgrade will (a go.work resolves them differently). Only what fails after the
# upgrade, and not already as it is, is reported. Nothing is written into the
# modules' directories.
#
# usage: scripts/downstream.sh <module dir>[=<test package>,...]...
#
#   Every module is built and vetted whole. Its tests run for ./... or, after
#   "=", only for the listed packages, for example
#   ../app/module=./internal/api/...,./web — use that to leave out tests that
#   reach real services.
#
#   DOWNSTREAM_TEST_FLAGS  extra `go test` flags, for example "-short -race".
set -euo pipefail

if [ $# -eq 0 ]; then
	echo "usage: $0 <module dir>[=<test package>,...]..." >&2
	exit 2
fi

goe_dir=$(cd "$(dirname "$0")/.." && pwd)
goe_module=$(cd "$goe_dir" && go list -m)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# check <out prefix> <test packages...>: builds, vets and tests the module in
# the current directory and writes one line per failure to <out prefix>.fail.
check() {
	local out=$1
	shift
	: >"$out.fail"
	go build ./... >"$out.build" 2>&1 || sed 's/^/build: /' "$out.build" >>"$out.fail"
	go vet ./... >"$out.vet" 2>&1 || grep -E '^[^#].*:[0-9]+:' "$out.vet" | sed 's/^/vet: /' >>"$out.fail" || true
	# shellcheck disable=SC2086 # DOWNSTREAM_TEST_FLAGS is a list of flags
	go test -count=1 -json ${DOWNSTREAM_TEST_FLAGS:-} "$@" >"$out.test" 2>/dev/null || true
	jq -r 'select(.Action == "fail" or .Action == "build-fail")
		| "test: \(.ImportPath // .Package)\(if .Test then " " + .Test else "" end)"' \
		"$out.test" >>"$out.fail" 2>/dev/null || true
	sort -u -o "$out.fail" "$out.fail"
}

status=0
for arg in "$@"; do
	dir=${arg%%=*}
	packages=(./...)
	if [ "$arg" != "$dir" ]; then
		IFS=, read -r -a packages <<<"${arg#*=}"
	fi
	mod_dir=$(cd "$dir" && pwd)
	name=$(basename "$mod_dir")
	echo "== $mod_dir"

	(cd "$mod_dir" && GOWORK=off check "$work/$name.base" "${packages[@]}")

	modfile="$work/$name.mod"
	cp "$mod_dir/go.mod" "$modfile"
	if [ -f "$mod_dir/go.sum" ]; then cp "$mod_dir/go.sum" "$work/$name.sum"; fi
	go mod edit -replace "$goe_module=$goe_dir" "$modfile"
	if ! (cd "$mod_dir" && GOWORK=off go mod tidy -modfile="$modfile" >"$work/$name.tidy" 2>&1); then
		echo "   go mod tidy failed after the upgrade:"
		sed 's/^/     /' "$work/$name.tidy"
		status=1
		continue
	fi
	(cd "$mod_dir" && GOWORK=off GOFLAGS="-modfile=$modfile" check "$work/$name.branch" "${packages[@]}")

	regressions=$(comm -13 "$work/$name.base.fail" "$work/$name.branch.fail")
	fixed=$(comm -23 "$work/$name.base.fail" "$work/$name.branch.fail")
	echo "   failing as it is: $(wc -l <"$work/$name.base.fail" | tr -d ' ')," \
		"against this tree: $(wc -l <"$work/$name.branch.fail" | tr -d ' ')"
	if [ -n "$fixed" ]; then
		echo "   passing only against this tree:"
		printf '     %s\n' "$fixed"
	fi
	if [ -n "$regressions" ]; then
		echo "   FAILING ONLY AGAINST THIS TREE:"
		printf '     %s\n' "$regressions"
		grep -h -A20 -E '^(# |.*:[0-9]+:[0-9]+: )' "$work/$name.branch.build" 2>/dev/null | head -40 | sed 's/^/       /' || true
		status=1
	else
		echo "   no regressions"
	fi
done
exit "$status"
