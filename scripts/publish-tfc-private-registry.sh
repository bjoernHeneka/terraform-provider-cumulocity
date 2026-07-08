#!/usr/bin/env bash
#
# Publish a provider release to the HCP Terraform (app.terraform.io) private
# registry. The private registry does NOT watch GitHub releases — every
# version must be pushed through the Registry Providers API:
# https://developer.hashicorp.com/terraform/cloud-docs/registry/publish-providers
#
# Prerequisites:
#   - GoReleaser output in ./dist (run the release workflow or
#     `goreleaser release --clean` locally first)
#   - jq and curl installed
#   - The GPG public key registered once via the GPG Keys API (see README)
#
# Required environment variables:
#   TFC_TOKEN    HCP Terraform API token (owners team)
#   TFC_ORG      Organization name
#   GPG_KEY_ID   key-id returned when the GPG key was registered
#   VERSION      Version to publish WITHOUT the leading v, e.g. 0.1.0

set -euo pipefail

: "${TFC_TOKEN:?TFC_TOKEN is required}"
: "${TFC_ORG:?TFC_ORG is required}"
: "${GPG_KEY_ID:?GPG_KEY_ID is required}"
: "${VERSION:?VERSION is required (without leading v)}"

PROVIDER_NAME="cumulocity"
PROJECT="terraform-provider-${PROVIDER_NAME}"
API="https://app.terraform.io/api/v2"
DIST="${DIST_DIR:-dist}"

# Keep the API token out of the process argument list (visible via `ps` or
# /proc/<pid>/cmdline on shared/CI hosts) by passing it through a curl config
# file instead of --header.
auth_cfg="$(mktemp)"
chmod 600 "$auth_cfg"
trap 'rm -f "$auth_cfg"' EXIT
printf 'header = "Authorization: Bearer %s"\n' "$TFC_TOKEN" >"$auth_cfg"

auth=(--config "$auth_cfg")
jsonapi=(--header "Content-Type: application/vnd.api+json")

req() { curl --fail-with-body --silent --show-error "${auth[@]}" "$@"; }

sums_file="${DIST}/${PROJECT}_${VERSION}_SHA256SUMS"
sig_file="${sums_file}.sig"
[ -f "$sums_file" ] || { echo "missing ${sums_file} — run goreleaser first" >&2; exit 1; }
[ -f "$sig_file" ] || { echo "missing ${sig_file} — run goreleaser first" >&2; exit 1; }

echo "==> Ensuring provider record exists (${TFC_ORG}/${PROVIDER_NAME})"
create_provider=$(jq -n \
  --arg name "$PROVIDER_NAME" \
  --arg namespace "$TFC_ORG" \
  '{data:{type:"registry-providers",attributes:{name:$name,namespace:$namespace,"registry-name":"private"}}}')
# 422 "has already been taken" is fine — the record is idempotent for our purposes.
if ! out=$(req "${jsonapi[@]}" --request POST --data "$create_provider" \
  "${API}/organizations/${TFC_ORG}/registry-providers" 2>&1); then
  echo "$out" | grep -q "has already been taken" || { echo "$out" >&2; exit 1; }
  echo "    provider record already exists"
fi

echo "==> Creating version ${VERSION}"
create_version=$(jq -n \
  --arg version "$VERSION" \
  --arg keyid "$GPG_KEY_ID" \
  '{data:{type:"registry-provider-versions",attributes:{version:$version,"key-id":$keyid,protocols:["6.0"]}}}')
version_resp=$(req "${jsonapi[@]}" --request POST --data "$create_version" \
  "${API}/organizations/${TFC_ORG}/registry-providers/private/${TFC_ORG}/${PROVIDER_NAME}/versions")

sums_url=$(echo "$version_resp" | jq -r '.data.links."shasums-upload"')
sig_url=$(echo "$version_resp" | jq -r '.data.links."shasums-sig-upload"')

echo "==> Uploading SHA256SUMS and signature"
curl --fail --silent --show-error --request PUT --upload-file "$sums_file" "$sums_url"
curl --fail --silent --show-error --request PUT --upload-file "$sig_file" "$sig_url"

echo "==> Creating platforms and uploading binaries"
for zip in "${DIST}/${PROJECT}_${VERSION}"_*_*.zip; do
  base=$(basename "$zip" .zip)
  os=$(echo "$base" | awk -F_ '{print $(NF-1)}')
  arch=$(echo "$base" | awk -F_ '{print $NF}')
  shasum=$(grep -F -- "$(basename "$zip")" "$sums_file" | awk '{print $1}')

  create_platform=$(jq -n \
    --arg os "$os" \
    --arg arch "$arch" \
    --arg shasum "$shasum" \
    --arg filename "$(basename "$zip")" \
    '{data:{type:"registry-provider-version-platforms",attributes:{os:$os,arch:$arch,shasum:$shasum,filename:$filename}}}')
  platform_resp=$(req "${jsonapi[@]}" --request POST --data "$create_platform" \
    "${API}/organizations/${TFC_ORG}/registry-providers/private/${TFC_ORG}/${PROVIDER_NAME}/versions/${VERSION}/platforms")
  upload_url=$(echo "$platform_resp" | jq -r '.data.links."provider-binary-upload"')
  curl --fail --silent --show-error --request PUT --upload-file "$zip" "$upload_url"
  echo "    ${os}/${arch} uploaded"
done

echo "==> Done. Source address: app.terraform.io/${TFC_ORG}/${PROVIDER_NAME}"
