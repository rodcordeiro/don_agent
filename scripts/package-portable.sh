#!/usr/bin/env sh
set -eu

VERSION="${VERSION:-dev}"
BUILDER="${BUILDER:-auto}"
OUTPUT_DIR="${OUTPUT_DIR:-dist}"
CLEAN="${CLEAN:-0}"

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

has_command() {
    command -v "$1" >/dev/null 2>&1
}

has_usable_docker() {
    has_command docker && docker info >/dev/null 2>&1
}

select_builder() {
    if [ "$BUILDER" != "auto" ]; then
        printf '%s\n' "$BUILDER"
        return
    fi

    if has_usable_docker; then
        printf '%s\n' "docker"
        return
    fi

    if has_command go; then
        printf '%s\n' "go"
        return
    fi

    printf '%s\n' "Neither docker nor go was found. Install Docker or Go, or set BUILDER explicitly." >&2
    exit 1
}

build_binary() {
    goos="$1"
    goarch="$2"
    output_path="$3"

    if [ "$SELECTED_BUILDER" = "docker" ]; then
        docker compose run --no-deps --rm \
            -e CGO_ENABLED=0 \
            -e GOOS="$goos" \
            -e GOARCH="$goarch" \
            dev go build -trimpath -ldflags="-s -w" -o "/workspace/$output_path" ./cmd/donagent
        return
    fi

    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
        go build -trimpath -ldflags="-s -w" -o "$REPO_ROOT/$output_path" ./cmd/donagent
}

create_archive() {
    package_dir="$1"
    archive_path="$2"
    archive_type="$3"

    rm -f "$archive_path"

    if [ "$archive_type" = "zip" ]; then
        if has_command zip; then
            (cd "$(dirname "$package_dir")" && zip -qr "$archive_path" "$(basename "$package_dir")")
            return
        fi

        if has_command powershell; then
            powershell -NoProfile -ExecutionPolicy Bypass -Command "Compress-Archive -Path '$package_dir/*' -DestinationPath '$archive_path'"
            return
        fi

        printf '%s\n' "zip was not found; cannot create .zip archive." >&2
        exit 1
    fi

    tar -czf "$archive_path" -C "$(dirname "$package_dir")" "$(basename "$package_dir")"
}

cd "$REPO_ROOT"
SELECTED_BUILDER=$(select_builder)

case "$SELECTED_BUILDER" in
    docker|go) ;;
    *) printf '%s\n' "Invalid BUILDER '$SELECTED_BUILDER'. Use auto, docker or go." >&2; exit 1 ;;
esac

case "$OUTPUT_DIR" in
    /*) printf '%s\n' "OUTPUT_DIR must be a relative directory inside the repository." >&2; exit 1 ;;
    *) RESOLVED_OUTPUT_DIR="$REPO_ROOT/$OUTPUT_DIR" ;;
esac

if [ "$CLEAN" = "1" ] && [ -d "$RESOLVED_OUTPUT_DIR" ]; then
    case "$RESOLVED_OUTPUT_DIR" in
        "$REPO_ROOT"/*) rm -rf "$RESOLVED_OUTPUT_DIR" ;;
        *) printf '%s\n' "Refusing to clean output directory outside repository: $RESOLVED_OUTPUT_DIR" >&2; exit 1 ;;
    esac
fi

mkdir -p "$RESOLVED_OUTPUT_DIR"

ARTIFACTS=""

for target in windows/amd64/donagent.exe/zip linux/amd64/donagent/tar.gz; do
    goos=$(printf '%s' "$target" | cut -d / -f 1)
    goarch=$(printf '%s' "$target" | cut -d / -f 2)
    binary=$(printf '%s' "$target" | cut -d / -f 3)
    archive_type=$(printf '%s' "$target" | cut -d / -f 4)
    platform="$goos-$goarch"
    package_name="donagent-$VERSION-$platform"
    package_dir="$RESOLVED_OUTPUT_DIR/portable/$package_name"
    relative_binary="$OUTPUT_DIR/portable/$package_name/$binary"

    rm -rf "$package_dir"
    mkdir -p "$package_dir/assets"

    build_binary "$goos" "$goarch" "$relative_binary"
    cp "$REPO_ROOT/config.example.toml" "$package_dir/config.example.toml"
    cp "$REPO_ROOT/assets/logo.png" "$package_dir/assets/logo.png"

    if [ "$archive_type" = "zip" ]; then
        archive_path="$RESOLVED_OUTPUT_DIR/$package_name.zip"
    else
        archive_path="$RESOLVED_OUTPUT_DIR/$package_name.tar.gz"
    fi

    create_archive "$package_dir" "$archive_path" "$archive_type"
    ARTIFACTS="$ARTIFACTS
 - $archive_path"
done

printf "Portable artifacts created with builder '%s':%s\n" "$SELECTED_BUILDER" "$ARTIFACTS"
