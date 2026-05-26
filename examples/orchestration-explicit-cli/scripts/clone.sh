#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <service-1|service-2> <active|latest|version-number>" >&2
  exit 1
fi

service_name="$1"
version="$2"

case "$service_name" in
  service-1|service_1) output_name="service_1_id" ;;
  service-2|service_2) output_name="service_2_id" ;;
  *)
    echo "unknown service name: $service_name" >&2
    exit 1
    ;;
esac

service_id="$(terraform output -raw "${output_name}")"

fastly service version clone \
  --service-id="${service_id}" \
  --version="${version}" \
  --json \
  --non-interactive
