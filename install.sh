#!/usr/bin/env sh

set -eu

repository="missedone/jira-cli"
install_dir="${JIRA_CLI_INSTALL_DIR:-/usr/local/bin}"
requested_version="${1:-${JIRA_CLI_VERSION:-latest}}"

log() {
	printf '%s\n' "$*"
}

fail() {
	printf 'jira-cli installer: %s\n' "$*" >&2
	exit 1
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

if [ "$#" -gt 1 ]; then
	fail "usage: install.sh [VERSION]"
fi

require_command curl
require_command tar
require_command uname
require_command mktemp
require_command install

case "$(uname -s)" in
	Darwin)
		asset_os="macOS"
		;;
	Linux)
		asset_os="linux"
		;;
	*)
		fail "unsupported operating system: $(uname -s)"
		;;
esac

case "$(uname -m)" in
	x86_64 | amd64)
		asset_arch="x86_64"
		;;
	arm64 | aarch64)
		asset_arch="arm64"
		;;
	i386 | i486 | i586 | i686)
		[ "$asset_os" = "linux" ] || fail "unsupported architecture on macOS: $(uname -m)"
		asset_arch="i386"
		;;
	armv6l | armv7l)
		[ "$asset_os" = "linux" ] || fail "unsupported architecture on macOS: $(uname -m)"
		asset_arch="armv6"
		;;
	*)
		fail "unsupported architecture: $(uname -m)"
		;;
esac

if [ "$requested_version" = "latest" ]; then
	if ! latest_url=$(curl --fail --silent --show-error --location --output /dev/null --write-out '%{url_effective}' \
		"https://github.com/${repository}/releases/latest"); then
		fail "unable to determine the latest release"
	fi
	case "$latest_url" in
		"https://github.com/${repository}/releases/tag/"*) tag="${latest_url##*/}" ;;
		*) fail "no published stable release found; pass an explicit published tag" ;;
	esac
else
	case "$requested_version" in
		v*) tag="$requested_version" ;;
		*) tag="v${requested_version}" ;;
	esac
fi

case "$tag" in
	v[0-9]*) ;;
	*) fail "invalid release version: ${tag}" ;;
esac

version="${tag#v}"
asset="jira_${version}_${asset_os}_${asset_arch}.tar.gz"
download_url="https://github.com/${repository}/releases/download/${tag}/${asset}"
checksum_url="https://github.com/${repository}/releases/download/${tag}/checksums.txt"

temp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t jira-cli)
trap 'rm -rf "$temp_dir"' EXIT HUP INT TERM

archive_path="${temp_dir}/${asset}"
checksum_path="${temp_dir}/checksums.txt"

log "Downloading jira-cli ${tag} for ${asset_os}/${asset_arch}..."
if ! curl --fail --silent --show-error --location --retry 3 --output "$archive_path" "$download_url"; then
	fail "unable to download ${download_url}"
fi
if ! curl --fail --silent --show-error --location --retry 3 --output "$checksum_path" "$checksum_url"; then
	fail "unable to download release checksums"
fi

expected_checksum=$(awk -v asset="$asset" '$2 == asset || $2 == "*" asset { print $1; exit }' "$checksum_path")
[ -n "$expected_checksum" ] || fail "checksum not found for ${asset}"

if command -v sha256sum >/dev/null 2>&1; then
	actual_checksum=$(sha256sum "$archive_path" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
	actual_checksum=$(shasum -a 256 "$archive_path" | awk '{print $1}')
else
	fail "sha256sum or shasum is required to verify the download"
fi

[ "$actual_checksum" = "$expected_checksum" ] || fail "checksum verification failed for ${asset}"

tar -xzf "$archive_path" -C "$temp_dir"
binary_path="${temp_dir}/${asset%.tar.gz}/bin/jira"
[ -f "$binary_path" ] || fail "jira binary not found in ${asset}"

if [ ! -d "$install_dir" ]; then
	if ! mkdir -p "$install_dir" 2>/dev/null; then
		command -v sudo >/dev/null 2>&1 || fail "cannot create ${install_dir}; set JIRA_CLI_INSTALL_DIR to a writable directory"
		sudo mkdir -p "$install_dir"
	fi
fi

if [ -w "$install_dir" ]; then
	install -m 0755 "$binary_path" "${install_dir}/jira"
else
	command -v sudo >/dev/null 2>&1 || fail "cannot write to ${install_dir}; set JIRA_CLI_INSTALL_DIR to a writable directory"
	sudo install -m 0755 "$binary_path" "${install_dir}/jira"
fi

log "Installed jira-cli ${tag} to ${install_dir}/jira"
case ":${PATH:-}:" in
	*":${install_dir}:"*) ;;
	*) log "Add ${install_dir} to PATH to run jira from your shell." ;;
esac
