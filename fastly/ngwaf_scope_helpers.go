package fastly

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
)

type resolvedScope struct {
	scope *scope.Scope
	ctx   context.Context
}

func resolveScopeAndContext(ctx context.Context, d *schema.ResourceData) (*resolvedScope, error) {
	s := buildNGWAFScope(d) // renamed to s
	if s == nil {
		return nil, fmt.Errorf("could not determine rule scope: missing workspace_id or applies_to")
	}

	if s.Type == scope.ScopeTypeWorkspace && len(s.AppliesTo) > 0 {
		if wsID := s.AppliesTo[0]; wsID != "" {
			ctx = gofastly.NewContextForResourceID(ctx, wsID)
		}
	}

	return &resolvedScope{
		scope: s,
		ctx:   ctx,
	}, nil
}

func buildNGWAFScope(d *schema.ResourceData) *scope.Scope {
	if v, ok := d.GetOk("workspace_id"); ok {
		wsID := v.(string)
		if wsID != "" {
			return &scope.Scope{
				Type:      scope.ScopeTypeWorkspace,
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
		return &scope.Scope{
			Type:      scope.ScopeTypeAccount,
			AppliesTo: ids,
		}
	}

	return nil
}

func customNGWAFScopeImporter(scopeType scope.Type, resourceType string) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
			switch scopeType {
			case scope.ScopeTypeWorkspace:
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

			case scope.ScopeTypeAccount:
				// Only rule ID is needed for account-scoped rules
				if err := d.Set("applies_to", []string{"*"}); err != nil {
					return nil, fmt.Errorf("failed to set applies_to for account %s: %w", resourceType, err)
				}
				d.SetId(d.Id())

			default:
				return nil, fmt.Errorf("unsupported scope type %q", scopeType)
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}
