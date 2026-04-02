#!/bin/sh

set -eu

project_name="agentlib"
repo="agentlibdev/agent-cli"
install_dir="${HOME}/.agentlib/bin"
version="${1:-latest}"

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || return 1
}

download() {
  url=$1
  dest=$2

  if need_cmd curl; then
    curl -fsSL "$url" -o "$dest"
    return
  fi

  if need_cmd wget; then
    wget -qO "$dest" "$url"
    return
  fi

  die "error: need curl or wget"
}

sha256_file() {
  file=$1

  if need_cmd sha256sum; then
    output=$(sha256sum "$file") || die "error: sha256sum failed"
    printf '%s\n' "$output" | awk '{print $1}'
    return
  fi

  if need_cmd shasum; then
    output=$(shasum -a 256 "$file") || die "error: shasum failed"
    printf '%s\n' "$output" | awk '{print $1}'
    return
  fi

  if need_cmd openssl; then
    output=$(openssl dgst -sha256 "$file") || die "error: openssl sha256 failed"
    printf '%s\n' "$output" | awk '{print $2}'
    return
  fi

  die "error: need sha256sum, shasum, or openssl"
}

normalize_os() {
  case "$(uname -s)" in
    Linux)
      printf '%s\n' linux
      ;;
    Darwin)
      printf '%s\n' darwin
      ;;
    *)
      die "error: unsupported operating system: $(uname -s)"
      ;;
  esac
}

normalize_arch() {
  case "$(uname -m)" in
    x86_64|amd64)
      printf '%s\n' amd64
      ;;
    aarch64|arm64)
      printf '%s\n' arm64
      ;;
    *)
      die "error: unsupported architecture: $(uname -m)"
      ;;
  esac
}

release_base_url() {
  if [ "$version" = "latest" ]; then
    printf 'https://github.com/%s/releases/latest/download\n' "$repo"
    return
  fi

  printf 'https://github.com/%s/releases/download/%s\n' "$repo" "$version"
}

latest_release_tag() {
  info=$1
  tag=$(awk -F'"' '/"tag_name"[[:space:]]*:/ { print $4; exit }' "$info")
  [ -n "$tag" ] || die "error: could not determine latest release tag"
  printf '%s\n' "$tag"
}

tmp_parent="${TMPDIR:-/tmp}"
umask 077
tmpdir=""
trap 'rm -rf "${tmpdir:-}"' EXIT INT HUP TERM
tmpdir=$(mktemp -d "${tmp_parent%/}/agentlib-install.XXXXXX") || die "error: could not create temporary directory"

os=$(normalize_os)
arch=$(normalize_arch)
release_tag=$version
if [ "$version" = "latest" ]; then
  latest_info="${tmpdir}/latest-release.json"
  download "https://api.github.com/repos/${repo}/releases/latest" "$latest_info"
  release_tag=$(latest_release_tag "$latest_info")
fi

base_url=$(release_base_url)
archive="${project_name}_${release_tag}_${os}_${arch}.tar.gz"
checksum_file="${project_name}_checksums.txt"
archive_path="${tmpdir}/${archive}"
checksum_path="${tmpdir}/${checksum_file}"
binary_path="${tmpdir}/${project_name}"

download "${base_url}/${checksum_file}" "$checksum_path"
download "${base_url}/${archive}" "$archive_path"

expected_sha=$(awk -v file="$archive" '$2 == file { print $1 }' "$checksum_path")
[ -n "$expected_sha" ] || die "error: could not find checksum for $archive"

actual_sha=$(sha256_file "$archive_path")
[ "$actual_sha" = "$expected_sha" ] || die "error: checksum mismatch for $archive"

tar -xzf "$archive_path" -C "$tmpdir"
[ -f "$binary_path" ] || die "error: extracted binary not found at $binary_path"

mkdir -p "$install_dir"
cp "$binary_path" "$install_dir/$project_name"
chmod 755 "$install_dir/$project_name"

case ":${PATH:-}:" in
  *":$install_dir:"*)
    printf 'agentlib installed to %s/%s\n' "$install_dir" "$project_name"
    printf 'PATH already includes %s\n' "$install_dir"
    ;;
  *)
    printf 'agentlib installed to %s/%s\n' "$install_dir" "$project_name"
    printf 'Add this to your shell profile:\n'
    printf '  export PATH="%s:$PATH"\n' "$install_dir"
    ;;
esac
