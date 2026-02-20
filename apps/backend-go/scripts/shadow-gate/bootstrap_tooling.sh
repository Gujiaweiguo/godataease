#!/usr/bin/env bash

set -euo pipefail

INSTALL_DIR="./tmp/shadow-gate/tools/bin"
KUBECTL_VERSION="v1.30.3"
HELM_VERSION="v3.15.4"

usage() {
  cat <<'EOF'
Bootstrap kubectl/helm for shadow gate preflight.

USAGE:
  bootstrap_tooling.sh [--install-dir <path>] [--kubectl-version <vX.Y.Z>] [--helm-version <vX.Y.Z>]

OUTPUT:
  Installs binaries into <install-dir> and prints PATH export command.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --install-dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    --kubectl-version)
      KUBECTL_VERSION="$2"
      shift 2
      ;;
    --helm-version)
      HELM_VERSION="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 2
fi

if ! command -v tar >/dev/null 2>&1; then
  echo "tar is required" >&2
  exit 2
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch_raw="$(uname -m)"
case "$arch_raw" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch_raw" >&2
    exit 2
    ;;
esac

mkdir -p "$INSTALL_DIR"
abs_install_dir="$(cd "$INSTALL_DIR" && pwd)"

kubectl_target="$abs_install_dir/kubectl"
helm_target="$abs_install_dir/helm"

if [[ ! -x "$kubectl_target" ]]; then
  kubectl_url="https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/${os}/${arch}/kubectl"
  echo "Downloading kubectl from $kubectl_url"
  curl -fsSL "$kubectl_url" -o "$kubectl_target"
  chmod +x "$kubectl_target"
fi

if [[ ! -x "$helm_target" ]]; then
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT
  helm_archive="${tmpdir}/helm.tar.gz"
  helm_url="https://get.helm.sh/helm-${HELM_VERSION}-${os}-${arch}.tar.gz"
  echo "Downloading helm from $helm_url"
  curl -fsSL "$helm_url" -o "$helm_archive"
  tar -xzf "$helm_archive" -C "$tmpdir"
  cp "${tmpdir}/${os}-${arch}/helm" "$helm_target"
  chmod +x "$helm_target"
fi

echo "Installed tooling directory: $abs_install_dir"
echo "Add to PATH: export PATH=\"$abs_install_dir:\$PATH\""
