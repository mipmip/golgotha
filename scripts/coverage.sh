#!/usr/bin/env bash
#
# coverage.sh — run the full test suite with coverage and enforce the gate:
#   * overall coverage >= 70%
#   * each core-logic package >= 80%
#
# It writes the coverage profile to $1 when given, otherwise $PWD/cover.out, and
# runs from the repository root so it works both in a dev shell and inside the
# nix check sandbox (git must be on PATH). Dependency-free: bash + go + coreutils.

set -euo pipefail

# Resolve the repo root as the parent of this script's directory; cd there so
# we never depend on absolute paths or the caller's CWD.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

PROFILE="${1:-${REPO_ROOT}/cover.out}"

# Thresholds.
OVERALL_MIN=70
CORE_MIN=80

# Core-logic packages that must each clear CORE_MIN.
CORE_PKGS=(
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/clonepath"
	"github.com/mipmip/huphop/internal/provider"
	"github.com/mipmip/huphop/internal/cache"
	"github.com/mipmip/huphop/internal/syncer"
)

echo "==> Running tests with coverage (profile: ${PROFILE})"
go test -coverprofile="${PROFILE}" -covermode=atomic ./...

# --- Overall coverage: the "total:" line from go tool cover -func. ---
OVERALL="$(go tool cover -func="${PROFILE}" | LC_ALL=C awk '/^total:/ { gsub(/%/, "", $NF); print $NF }')"

# --- Per-core-package coverage. ---
#
# go tool cover -func emits one line per function prefixed with the file path
# (which starts with the package import path). We average the per-function
# percentages weighted equally is not what we want; instead we recompute the
# statement-weighted percentage per package directly from the profile.
#
# The profile's block lines look like:
#   <import-path>/<file>:<startLine.col>,<endLine.col> <numStmts> <count>
# The header line "mode: atomic" is skipped. For each package we sum numStmts
# and the covered numStmts (count > 0) to get an exact statement-coverage %.
pkg_coverage() {
	# $1 = package import path. LC_ALL=C keeps awk's printf using "." decimals.
	LC_ALL=C awk -v pkg="$1/" '
		NR == 1 { next }                 # skip the "mode:" header
		index($0, pkg) == 1 {
			stmts = $(NF-1)
			count = $NF
			total += stmts
			if (count+0 > 0) covered += stmts
		}
		END {
			if (total == 0) { print "NA"; exit }
			printf "%.1f", (covered / total) * 100
		}
	' "${PROFILE}"
}

FAIL=0

# Note: numeric comparisons use awk (which parses "." decimals regardless of the
# shell locale); printf uses %s so a comma-locale printf never chokes on floats.
printf '\n==> Coverage summary\n'
printf '  %-45s %9s\n' "PACKAGE" "COVERAGE"
printf '  %-45s %8s\n' "OVERALL" "${OVERALL}%"

# Overall gate.
if awk -v v="${OVERALL}" -v m="${OVERALL_MIN}" 'BEGIN { exit !(v+0 < m) }'; then
	echo "  FAIL: overall ${OVERALL}% < ${OVERALL_MIN}%"
	FAIL=1
fi

# Per-core-package gate.
for pkg in "${CORE_PKGS[@]}"; do
	cov="$(pkg_coverage "${pkg}")"
	short="${pkg#github.com/mipmip/huphop/}"
	if [ "${cov}" = "NA" ]; then
		printf '  %-45s %9s\n' "${short}" "NA"
		echo "  FAIL: ${short} produced no coverage data"
		FAIL=1
		continue
	fi
	printf '  %-45s %8s\n' "${short}" "${cov}%"
	if awk -v v="${cov}" -v m="${CORE_MIN}" 'BEGIN { exit !(v+0 < m) }'; then
		echo "  FAIL: ${short} ${cov}% < ${CORE_MIN}%"
		FAIL=1
	fi
done

printf '\n'
if [ "${FAIL}" -ne 0 ]; then
	echo "COVERAGE GATE: FAIL"
	exit 1
fi
echo "COVERAGE GATE: PASS (overall >= ${OVERALL_MIN}%, core packages >= ${CORE_MIN}%)"
