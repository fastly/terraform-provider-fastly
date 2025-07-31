package fastly

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/common"
)

type resolvedScope struct {
	scope *common.Scope
	ctx   context.Context
}

func resolveScopeAndContext(ctx context.Context, d *schema.ResourceData) (*resolvedScope, error) {
	scope := buildNGWAFScope(d)
	if scope == nil {
		return nil, fmt.Errorf("could not determine rule scope: missing workspace_id or applies_to")
	}

	if scope.Type == common.ScopeTypeWorkspace && len(scope.AppliesTo) > 0 {
		if wsID := scope.AppliesTo[0]; wsID != "" {
			ctx = gofastly.NewContextForResourceID(ctx, wsID)
		}
	}

	return &resolvedScope{
		scope: scope,
		ctx:   ctx,
	}, nil
}

func buildNGWAFScope(d *schema.ResourceData) *common.Scope {
	if v, ok := d.GetOk("workspace_id"); ok {
		wsID := v.(string)
		if wsID != "" {
			return &common.Scope{
				Type:      common.ScopeTypeWorkspace,
				AppliesTo: []string{wsID},
			}
		}
	}

	if v, ok := d.GetOk("applies_to"); ok {
		rawList, ok := v.([]any)
		if !ok || len(rawList) == 0 {
			return nil
		}
		ids := make([]string, len(rawList))
		for i, id := range rawList {
			ids[i] = strings.TrimSpace(id.(string))
		}
		return &common.Scope{
			Type:      common.ScopeTypeAccount,
			AppliesTo: ids,
		}
	}

	return nil
}

func customNGWAFScopeImporter(scope common.ScopeType, resourceType string) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
			switch scope {
			case common.ScopeTypeWorkspace:
				// Expected format: "workspace_id/rule_id"
				parts := strings.SplitN(d.Id(), "/", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid ID format %q. Expected workspace_id/%s_id", d.Id(), resourceType)
				}
				workspaceID := parts[0]
				ruleID := parts[1]

				if err := d.Set("workspace_id", workspaceID); err != nil {
					return nil, fmt.Errorf("failed to set workspace_id: %w", err)
				}
				d.SetId(ruleID)

			case common.ScopeTypeAccount:
				// Only rule ID is needed for account-scoped rules
				if err := d.Set("applies_to", []string{"*"}); err != nil {
					return nil, fmt.Errorf("failed to set applies_to for account %s: %w", resourceType, err)
				}
				d.SetId(d.Id())

			default:
				return nil, fmt.Errorf("unsupported scope type %q", scope)
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}
