#!/bin/sh

set -eu

BINARY_NAME="gx"
PROJECT_NAME="gx"
WINDOWS_EXTENSION=".exe"
WINDOWS_OS="windows"
TEMP_DIR_TEMPLATE="${TMPDIR:-/tmp}/gx-release.XXXXXX"

target_os="${1:?target os is required}"
target_arch="${2:?target arch is required}"
output_dir="${3:?output dir is required}"
version="${VERSION:-dev}"
normalized_version="$(printf '%s' "${version}" | sed 's/^v//')"
archive_base_name="${PROJECT_NAME}_${normalized_version}_${target_os}_${target_arch}"
build_root="$(mktemp -d "${TEMP_DIR_TEMPLATE}")"
package_dir="${build_root}/${archive_base_name}"
go_ldflags="-X github.com/XDwanj/gx/internal/app.Version=${version}"

cleanup() {
	rm -rf "${build_root}"
}

trap cleanup EXIT INT TERM

mkdir -p "${package_dir}" "${output_dir}"
output_dir_abs="$(cd "${output_dir}" && pwd)"

binary_file_name="${BINARY_NAME}"
if [ "${target_os}" = "${WINDOWS_OS}" ]; then
	binary_file_name="${BINARY_NAME}${WINDOWS_EXTENSION}"
fi

GOOS="${target_os}" GOARCH="${target_arch}" CGO_ENABLED=1 go build -trimpath -ldflags "${go_ldflags}" -o "${package_dir}/${binary_file_name}" .
cp README.md LICENSE "${package_dir}/"

if [ "${target_os}" = "${WINDOWS_OS}" ]; then
	archive_path="${output_dir_abs}/${archive_base_name}.zip"
	if ! command -v zip >/dev/null 2>&1; then
		echo "zip is required to package ${target_os}/${target_arch}" >&2
		exit 1
	fi
	(
		cd "${build_root}"
		zip -rq "${archive_path}" "${archive_base_name}"
	)
	printf '%s\n' "${archive_path}"
	exit 0
fi

archive_path="${output_dir_abs}/${archive_base_name}.tar.gz"
tar -C "${build_root}" -czf "${archive_path}" "${archive_base_name}"
printf '%s\n' "${archive_path}"
