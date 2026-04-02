#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <terraform-plan-json>" >&2
  exit 1
fi

plan_json="$1"

changed_service_ids="$(
  jq -r '
    [
      .resource_changes[]
      | select(.mode == "managed")
      | select(.type | IN("fastly_service_domain", "fastly_service_backend", "fastly_service_vcl"))
      | select(
          (.change.actions | index("create")) or
          (.change.actions | index("update")) or
          (.change.actions | index("delete")) or
          (.change.actions | index("replace"))
        )
      | (.change.after.service_id // .change.before.service_id)
      | select(. != null)
    ]
    | unique[]
  ' "${plan_json}"
)"

if [[ -z "${changed_service_ids}" ]]; then
  echo "No changed Fastly services found in plan; nothing to activate."
  exit 0
fi

service_1_id="$(terraform output -raw service_1_id)"
service_2_id="$(terraform output -raw service_2_id)"

service_1_version="$(terraform console <<< 'var.service_1_version' | tr -d '"' | tail -n1)"
service_2_version="$(terraform console <<< 'var.service_2_version' | tr -d '"' | tail -n1)"

while read -r service_id; do
  [[ -z "${service_id}" ]] && continue

  if [[ "${service_id}" == "${service_1_id}" ]]; then
    echo "Activating changed service: service-1 (${service_id}) version ${service_1_version}"
    ./scripts/activate.sh service-1 "${service_1_version}"
  elif [[ "${service_id}" == "${service_2_id}" ]]; then
    echo "Activating changed service: service-2 (${service_id}) version ${service_2_version}"
    ./scripts/activate.sh service-2 "${service_2_version}"
  else
    echo "Skipping unknown changed service: ${service_id}" >&2
  fi
done <<< "${changed_service_ids}"
