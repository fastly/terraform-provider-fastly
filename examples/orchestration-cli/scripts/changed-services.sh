#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <terraform-plan-json>" >&2
  exit 1
fi

plan_json="$1"

jq '
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
    | {
        address,
        type,
        actions: .change.actions,
        service_id: (.change.after.service_id // .change.before.service_id)
      }
    | select(.service_id != null)
  ]
  | sort_by(.service_id)
  | group_by(.service_id)
  | map({
      service_id: .[0].service_id,
      changes: map({
        address,
        type,
        actions
      })
    })
' "${plan_json}"
