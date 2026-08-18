#!/usr/bin/env bash
#
# release.sh — interactive, gated release for huphop.
#
# Flow:
#   1. Safety checks: clean working tree, on `main`, CHANGELOG has an
#      `Unreleased` section, target tag absent, and `nix flake check` passes.
#   2. gum prompt for the bump (major/minor/patch); compute the new version.
#   3. Update VERSION; promote the CHANGELOG `Unreleased` section to the new
#      version + today's date (leaving a fresh empty `Unreleased`).
#   4. Auto-update the flake vendorHash (fake-hash → nix build → parse → write);
#      skip with a warning if nix is absent (no dep change → hash is stable).
#   5. jj-first: commit the bump, set the `main` bookmark, `jj git push`, then
#      create and push the `vX.Y.Z` git tag which triggers the release workflow.
#
# Requires: bash, jj, git. Recommended: gum (prompt), nix (gate + vendorHash).
# Degrades gracefully when gum/nix are missing.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

VERSION_FILE="${REPO_ROOT}/VERSION"
CHANGELOG="${REPO_ROOT}/CHANGELOG.md"
FLAKE="${REPO_ROOT}/flake.nix"
FAKE_HASH="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

# --- output helpers -------------------------------------------------------
if command -v gum >/dev/null 2>&1; then
	HAVE_GUM=1
else
	HAVE_GUM=0
fi

info() { printf '==> %s\n' "$*"; }
warn() { printf 'WARNING: %s\n' "$*" >&2; }
die() {
	printf 'ERROR: %s\n' "$*" >&2
	exit 1
}

# --- prerequisites --------------------------------------------------------
command -v jj >/dev/null 2>&1 || die "jj is required"
command -v git >/dev/null 2>&1 || die "git is required"
if [ "${HAVE_GUM}" -eq 0 ]; then
	die "gum is not installed. Install it (e.g. 'nix profile install nixpkgs#gum' or see https://github.com/charmbracelet/gum) or enter 'nix develop'."
fi

# --- safety checks --------------------------------------------------------
info "Running safety checks"

# Clean working tree (jj: no changes in the working-copy commit vs its parent).
if [ -n "$(jj diff --no-pager 2>/dev/null)" ]; then
	die "working tree is not clean; commit or discard changes first"
fi

# On main: the `main` bookmark must point at the current change or its parent.
CURRENT_BOOKMARKS="$(jj log --no-pager -r '@ | @-' -T 'bookmarks ++ "\n"' 2>/dev/null || true)"
if ! printf '%s\n' "${CURRENT_BOOKMARKS}" | grep -qw "main"; then
	die "not on 'main' (the main bookmark is not at @ or @-)"
fi

# CHANGELOG has an Unreleased section.
[ -f "${CHANGELOG}" ] || die "CHANGELOG.md not found"
grep -qiE '^##[[:space:]]+\[?unreleased\]?' "${CHANGELOG}" ||
	die "CHANGELOG.md has no 'Unreleased' section"

# --- current version + bump prompt ---------------------------------------
[ -f "${VERSION_FILE}" ] || die "VERSION file not found"
CURRENT="$(tr -d '[:space:]' <"${VERSION_FILE}")"
[ -n "${CURRENT}" ] || die "VERSION file is empty"
info "Current version: ${CURRENT}"

IFS='.' read -r MAJOR MINOR PATCH <<<"${CURRENT}"
[ -n "${MAJOR}" ] && [ -n "${MINOR}" ] && [ -n "${PATCH}" ] ||
	die "VERSION '${CURRENT}' is not X.Y.Z"

BUMP="$(gum choose --header "Select version bump (current ${CURRENT}):" patch minor major)"
[ -n "${BUMP}" ] || die "no bump selected"

case "${BUMP}" in
patch) NEW="${MAJOR}.${MINOR}.$((PATCH + 1))" ;;
minor) NEW="${MAJOR}.$((MINOR + 1)).0" ;;
major) NEW="$((MAJOR + 1)).0.0" ;;
*) die "unknown bump '${BUMP}'" ;;
esac
TAG="v${NEW}"
info "New version: ${NEW} (tag ${TAG})"

# Target tag must not already exist.
if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null 2>&1; then
	die "tag ${TAG} already exists"
fi

# nix flake check gate.
if command -v nix >/dev/null 2>&1; then
	info "Running 'nix flake check' (build + tests + coverage gate)"
	nix flake check || die "nix flake check failed; aborting before any changes"
else
	warn "nix not installed; skipping 'nix flake check' gate"
fi

gum confirm "Proceed with releasing ${TAG}?" || die "aborted by user"

# --- 1. update VERSION ----------------------------------------------------
info "Updating VERSION → ${NEW}"
printf '%s\n' "${NEW}" >"${VERSION_FILE}"

# --- 2. promote CHANGELOG Unreleased → version + date --------------------
info "Promoting CHANGELOG Unreleased → ${NEW}"
TODAY="$(date +%Y-%m-%d)"
# Insert a fresh empty Unreleased section above the promoted heading, and
# rename the old Unreleased heading to [NEW] - DATE. Matches "## [Unreleased]"
# or "## Unreleased" (case-insensitive).
awk -v ver="${NEW}" -v today="${TODAY}" '
	!done && $0 ~ /^##[[:space:]]+\[?[Uu]nreleased\]?/ {
		print "## [Unreleased]"
		print ""
		print "## [" ver "] - " today
		done = 1
		next
	}
	{ print }
' "${CHANGELOG}" >"${CHANGELOG}.tmp" && mv "${CHANGELOG}.tmp" "${CHANGELOG}"

# --- 3. auto-update flake vendorHash -------------------------------------
if command -v nix >/dev/null 2>&1; then
	info "Recomputing flake vendorHash"
	CURRENT_HASH="$(grep -oE 'vendorHash = "[^"]*"' "${FLAKE}" | head -n1 | sed -E 's/vendorHash = "(.*)"/\1/')"
	# Set the fake hash, then build to learn the expected hash.
	sed -i -E "s|vendorHash = \"[^\"]*\"|vendorHash = \"${FAKE_HASH}\"|" "${FLAKE}"
	EXPECTED="$(nix build .#default 2>&1 | grep -oE 'got:[[:space:]]+sha256-[A-Za-z0-9+/=]+' | head -n1 | grep -oE 'sha256-[A-Za-z0-9+/=]+' || true)"
	if [ -n "${EXPECTED}" ]; then
		sed -i -E "s|vendorHash = \"[^\"]*\"|vendorHash = \"${EXPECTED}\"|" "${FLAKE}"
		info "vendorHash → ${EXPECTED}"
	else
		# No dep change means the build succeeded with the fake hash removed?
		# Unlikely; restore the previous known-good hash and continue.
		sed -i -E "s|vendorHash = \"[^\"]*\"|vendorHash = \"${CURRENT_HASH}\"|" "${FLAKE}"
		warn "could not parse expected vendorHash; restored previous value ${CURRENT_HASH}"
	fi
else
	warn "nix not installed; skipping vendorHash update (no dependency change → hash is stable)"
fi

# --- 4. jj-first commit + git tag push -----------------------------------
info "Committing the version bump via jj"
jj commit -m "Release ${TAG}"
jj bookmark set main -r @-
info "Pushing main bookmark"
jj git push --bookmark main

info "Creating and pushing tag ${TAG}"
# Resolve the release commit's git SHA (jj's @- / the main bookmark); `git tag`
# does not understand jj's `@-` revision syntax.
REL_COMMIT="$(jj log --no-pager -r @- -T 'commit_id' --no-graph)"
git tag "${TAG}" "${REL_COMMIT}"
git push origin "${TAG}"

info "Done. Tag ${TAG} pushed; GitHub Actions will build and publish the release."
info "Watch: https://github.com/mipmip/huphop/actions"
