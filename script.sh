#!/usr/bin/env bash
set -euo pipefail

pkg_name="record"
pkg_version="latest"

repo="${REPO_ROOT:-.}"
reports="${JUNIT_PATHS:-reports/}"

trap "cleanup" EXIT

cleanup() {
	rm -rf "./${pkg_name}" "./${pkg_name}.tar.gz"
}

error() {
	local msg="$1"
	echo "error: $msg" >&2
	exit 1
}

if [[ -z "${TESTLAB_KEY}" ]]; then
	error "environment variable TESTLAB_KEY is empty"
fi

if [[ -z "${TESTLAB_GROUP}" ]]; then
	error "environment variable TESTLAB_GROUP is empty"
fi

os="$(uname -s)"
platform="$(uname -m)"

set -x

if [[ ! (-f "./${pkg_name}") ]]; then
	curl --fail --location --retry 3 "https://github.com/testlabtools/${pkg_name}/releases/${pkg_version}/download/${pkg_name}-${os}-${platform}.tar.gz" -o "./${pkg_name}.tar.gz"
	tar -xvf "${pkg_name}.tar.gz"
fi
chmod +x "${pkg_name}"

./record upload --repo "$repo" --reports "$reports"
