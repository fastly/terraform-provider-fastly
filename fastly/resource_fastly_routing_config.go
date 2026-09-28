package fastly

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths/rules"
)

// resourceFastlyRoutingConfig manages a Fastly routing config using automatic
// versioning: every apply reconciles the routing config's draft paths and
// rules against the resource's configuration, then activates the draft. The
// draft/version workflow described by the API is never exposed to Terraform.
//
// Reconciliation and refresh always operate on path/rule IDs this resource
// already knows about (from its own prior Create/Update calls), fetched
// individually via Get, rather than by listing a path's/routing config's
// current children. The list endpoints are only used once, to bootstrap
// state on import.
func resourceFastlyRoutingConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyRoutingConfigCreate,
		ReadContext:   resourceFastlyRoutingConfigRead,
		UpdateContext: resourceFastlyRoutingConfigUpdate,
		DeleteContext: resourceFastlyRoutingConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyRoutingConfigImport,
		},
		CustomizeDiff: resourceFastlyRoutingConfigCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"activated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time the routing config's active version was last activated, in ISO 8601 format.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user-defined name of the routing config. Cannot be changed after creation.",
			},
			"path": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A URL path pattern and the rules used to route matching requests. Fastly manages the draft/active version lifecycle for these automatically; every apply reconciles the desired paths and rules and activates the result.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The URL path pattern (max 2048 characters, starts with `/`).",
						},
						"path_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path identifier.",
						},
						"rule": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Description: "A conditional routing rule for this path. A rule with no `condition` blocks is the default (catch-all) rule for the path.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action_type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The action type (e.g. `service`).",
									},
									"action_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The destination for the action (e.g. a service ID when `action_type` is `service`).",
									},
									"condition": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "A matching criterion evaluated against the incoming request. Omit to make this the default (catch-all) rule for the path.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The header name to match against. Only applicable when `type` is `header`.",
												},
												"operator": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The comparison operator used to evaluate `value` against the request (e.g. `equals`).",
												},
												"type": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The condition category (e.g. `header`).",
												},
												"value": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The value compared against the request using `operator`.",
												},
											},
										},
									},
									"is_default": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether this is the catch-all rule for its path (i.e. it has no conditions).",
									},
									"rule_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The rule identifier.",
									},
								},
							},
						},
					},
				},
			},
			"routing_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The routing config identifier.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current lifecycle state of the routing config (e.g. `draft-only`, `active`, or `active-with-draft`).",
			},
		},
	}
}

// resourceFastlyRoutingConfigCustomizeDiff rejects configurations with
// duplicate `path` blocks, since Create/Update key paths by their `path`
// string and would otherwise silently collide.
func resourceFastlyRoutingConfigCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	desiredPaths := d.Get("path").([]interface{})
	seen := make(map[string]bool, len(desiredPaths))
	for _, raw := range desiredPaths {
		pm := raw.(map[string]interface{})
		pathStr := pm["path"].(string)
		if seen[pathStr] {
			return fmt.Errorf("duplicate path %q: each `path` block must have a unique `path` value", pathStr)
		}
		seen[pathStr] = true
	}
	return nil
}

func resourceFastlyRoutingConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	data, err := routingconfigs.Create(ctx, conn, &routingconfigs.CreateInput{
		Name: new(d.Get("name").(string)),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(data.RoutingConfigID)

	desiredPaths := d.Get("path").([]interface{})
	pathList, err := createRoutingConfigPaths(ctx, conn, data.RoutingConfigID, desiredPaths)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("path", pathList); err != nil {
		return diag.FromErr(err)
	}

	if len(pathList) > 0 {
		if _, err := routingconfigs.Activate(ctx, conn, &routingconfigs.ActivateInput{
			RoutingConfigID: &data.RoutingConfigID,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFastlyRoutingConfigReadTopLevel(ctx, d, meta)
}

func resourceFastlyRoutingConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	id := d.Id()

	oldRaw, newRaw := d.GetChange("path")
	oldPaths := oldRaw.([]interface{})
	desiredPaths := newRaw.([]interface{})

	oldByPath := make(map[string]map[string]interface{}, len(oldPaths))
	for _, raw := range oldPaths {
		pm := raw.(map[string]interface{})
		oldByPath[pm["path"].(string)] = pm
	}

	desiredPathStrings := make(map[string]bool, len(desiredPaths))
	for _, raw := range desiredPaths {
		pm := raw.(map[string]interface{})
		desiredPathStrings[pm["path"].(string)] = true
	}

	changed := false

	for pathStr, pm := range oldByPath {
		if desiredPathStrings[pathStr] {
			continue
		}
		pathID := pm["path_id"].(string)
		if err := paths.Delete(ctx, conn, &paths.DeleteInput{
			RoutingConfigID: &id,
			PathID:          new(pathID),
		}); err != nil && !isNotFoundErr(err) {
			return diag.FromErr(fmt.Errorf("failed to delete routing config path %q: %w", pathStr, err))
		}
		changed = true
	}

	pathList := make([]map[string]any, 0, len(desiredPaths))
	for _, raw := range desiredPaths {
		pm := raw.(map[string]interface{})
		pathStr := pm["path"].(string)
		desiredRules := pm["rule"].([]interface{})

		if old, ok := oldByPath[pathStr]; ok {
			pathID := old["path_id"].(string)
			ruleList, rulesChanged, err := reconcileRoutingConfigRulesTracked(ctx, conn, id, pathID, old["rule"].([]interface{}), desiredRules)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to reconcile rules for routing config path %q: %w", pathStr, err))
			}
			if rulesChanged {
				changed = true
			}
			pathList = append(pathList, map[string]any{
				"path":    pathStr,
				"path_id": pathID,
				"rule":    ruleList,
			})
			continue
		}

		created, err := paths.Create(ctx, conn, &paths.CreateInput{
			RoutingConfigID: &id,
			Path:            new(pathStr),
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to create routing config path %q: %w", pathStr, err))
		}
		ruleList, err := createRoutingConfigRules(ctx, conn, id, created.PathID, desiredRules)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to create rules for routing config path %q: %w", pathStr, err))
		}
		pathList = append(pathList, map[string]any{
			"path":    pathStr,
			"path_id": created.PathID,
			"rule":    ruleList,
		})
		changed = true
	}

	if err := d.Set("path", pathList); err != nil {
		return diag.FromErr(err)
	}

	if changed {
		if _, err := routingconfigs.Activate(ctx, conn, &routingconfigs.ActivateInput{
			RoutingConfigID: &id,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFastlyRoutingConfigReadTopLevel(ctx, d, meta)
}

// resourceFastlyRoutingConfigRead refreshes the routing config, including its
// paths and rules. It never lists: each path/rule already known to state is
// re-fetched individually by ID, and any that now 404 are dropped as removed.
// This intentionally cannot discover a path or rule added out-of-band (i.e.
// not through this resource); doing so would require the list endpoints,
// which are unreliable immediately after activation (see CDTOOL-1745).
func resourceFastlyRoutingConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Routing Config Configuration for (%s)", d.Id())

	if diags := resourceFastlyRoutingConfigReadTopLevel(ctx, d, meta); diags.HasError() || d.Id() == "" {
		return diags
	}

	conn := meta.(*APIClient).conn
	id := d.Id()

	currentPaths := d.Get("path").([]interface{})
	pathList := make([]map[string]any, 0, len(currentPaths))
	for _, raw := range currentPaths {
		pm := raw.(map[string]interface{})
		pathID := pm["path_id"].(string)
		if pathID == "" {
			continue
		}

		got, err := paths.Get(ctx, conn, &paths.GetInput{RoutingConfigID: &id, PathID: &pathID})
		if err != nil {
			if isNotFoundErr(err) {
				continue
			}
			return diag.FromErr(err)
		}

		currentRules := pm["rule"].([]interface{})
		ruleList := make([]map[string]any, 0, len(currentRules))
		for _, rraw := range currentRules {
			rm := rraw.(map[string]interface{})
			ruleID := rm["rule_id"].(string)
			if ruleID == "" {
				continue
			}
			gotRule, err := rules.Get(ctx, conn, &rules.GetInput{RoutingConfigID: &id, PathID: &pathID, RuleID: &ruleID})
			if err != nil {
				if isNotFoundErr(err) {
					continue
				}
				return diag.FromErr(err)
			}
			ruleList = append(ruleList, flattenRoutingConfigRule(*gotRule))
		}

		pathList = append(pathList, map[string]any{
			"path":    got.Path,
			"path_id": got.PathID,
			"rule":    ruleList,
		})
	}

	if err := d.Set("path", pathList); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceFastlyRoutingConfigReadTopLevel refreshes only the routing config's
// own (non-path) attributes.
func resourceFastlyRoutingConfigReadTopLevel(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	id := d.Id()

	data, err := routingconfigs.Get(ctx, conn, &routingconfigs.GetInput{RoutingConfigID: &id})
	if err != nil {
		if isNotFoundErr(err) {
			log.Printf("[WARN] Routing config (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", data.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("routing_config_id", data.RoutingConfigID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", data.State); err != nil {
		return diag.FromErr(err)
	}
	activatedAt := ""
	if data.ActivatedAt != nil {
		activatedAt = data.ActivatedAt.Format(time.RFC3339)
	}
	if err := d.Set("activated_at", activatedAt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyRoutingConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	id := d.Id()

	err := routingconfigs.Delete(ctx, conn, &routingconfigs.DeleteInput{
		RoutingConfigID: &id,
		Force:           new(true),
	})
	if err != nil && !isNotFoundErr(err) {
		return diag.FromErr(err)
	}
	return nil
}

// resourceFastlyRoutingConfigImport bootstraps state for an existing routing
// config by listing its current paths and rules. This is the one place this
// resource relies on the list endpoints, since import has no prior known IDs
// to Get by; it's safe here because an importable routing config necessarily
// already has real content, so the list call won't be the "empty config"
// 404 that gets cached (see CDTOOL-1745).
func resourceFastlyRoutingConfigImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*APIClient).conn
	id := d.Id()

	actualPaths, err := paths.List(ctx, conn, &paths.ListInput{RoutingConfigID: &id})
	if err != nil {
		return nil, fmt.Errorf("failed to list routing config paths: %w", err)
	}

	pathList := make([]map[string]any, 0, len(actualPaths))
	for _, p := range actualPaths {
		actualRules, err := rules.List(ctx, conn, &rules.ListInput{RoutingConfigID: &id, PathID: &p.PathID})
		if err != nil {
			return nil, fmt.Errorf("failed to list rules for routing config path %q: %w", p.Path, err)
		}
		ruleList := make([]map[string]any, 0, len(actualRules))
		for _, r := range actualRules {
			ruleList = append(ruleList, flattenRoutingConfigRule(r))
		}
		pathList = append(pathList, map[string]any{
			"path":    p.Path,
			"path_id": p.PathID,
			"rule":    ruleList,
		})
	}

	if err := d.Set("path", pathList); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// createRoutingConfigPaths creates every desired path (and its rules) for a
// routing config that has none of them yet, returning the resulting `path`
// attribute value built directly from the create responses.
func createRoutingConfigPaths(ctx context.Context, conn *gofastly.Client, routingConfigID string, desiredPaths []interface{}) ([]map[string]any, error) {
	pathList := make([]map[string]any, 0, len(desiredPaths))
	for _, raw := range desiredPaths {
		pm := raw.(map[string]interface{})
		pathStr := pm["path"].(string)

		created, err := paths.Create(ctx, conn, &paths.CreateInput{
			RoutingConfigID: &routingConfigID,
			Path:            new(pathStr),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create routing config path %q: %w", pathStr, err)
		}

		ruleList, err := createRoutingConfigRules(ctx, conn, routingConfigID, created.PathID, pm["rule"].([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to create rules for routing config path %q: %w", pathStr, err)
		}

		pathList = append(pathList, map[string]any{
			"path":    pathStr,
			"path_id": created.PathID,
			"rule":    ruleList,
		})
	}
	return pathList, nil
}

// createRoutingConfigRules creates every desired rule for a path that has
// none yet, returning the resulting `rule` attribute value built directly
// from the create responses.
func createRoutingConfigRules(ctx context.Context, conn *gofastly.Client, routingConfigID, pathID string, desiredRules []interface{}) ([]map[string]any, error) {
	ruleList := make([]map[string]any, 0, len(desiredRules))
	for _, raw := range desiredRules {
		rm := raw.(map[string]interface{})
		action, conditions := expandRoutingConfigRule(rm)

		created, err := rules.Create(ctx, conn, &rules.CreateInput{
			RoutingConfigID: &routingConfigID,
			PathID:          &pathID,
			Action:          &action,
			Conditions:      conditions,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create rule: %w", err)
		}
		ruleList = append(ruleList, flattenRoutingConfigRule(*created))
	}
	return ruleList, nil
}

// reconcileRoutingConfigRulesTracked reconciles a path's rules using only
// rule IDs already known from prior state (never lists). If the desired rule
// set is unchanged (by content, ignoring order), the prior rule maps are
// returned untouched; otherwise every existing rule (by its known ID) is
// deleted and the desired rules are recreated fresh, since rules have no
// user-supplied natural key to reconcile against individually.
// reconcileRoutingConfigRulesTracked returns whether it actually deleted or
// created any rule, so the caller can skip activation when nothing changed.
func reconcileRoutingConfigRulesTracked(ctx context.Context, conn *gofastly.Client, routingConfigID, pathID string, oldRules, desiredRules []interface{}) ([]map[string]any, bool, error) {
	oldIDs := make([]string, len(oldRules))
	oldSigs := make([]string, len(oldRules))
	for i, raw := range oldRules {
		rm := raw.(map[string]interface{})
		action, conditions := expandRoutingConfigRule(rm)
		sig, err := ruleSignature(action, conditions)
		if err != nil {
			return nil, false, err
		}
		oldSigs[i] = sig
		oldIDs[i] = rm["rule_id"].(string)
	}

	desiredActions := make([]rules.Action, len(desiredRules))
	desiredConditions := make([][]rules.Condition, len(desiredRules))
	desiredSigs := make([]string, len(desiredRules))
	for i, raw := range desiredRules {
		rm := raw.(map[string]interface{})
		action, conditions := expandRoutingConfigRule(rm)
		desiredActions[i] = action
		desiredConditions[i] = conditions
		sig, err := ruleSignature(action, conditions)
		if err != nil {
			return nil, false, err
		}
		desiredSigs[i] = sig
	}

	if sameSignatureSet(oldSigs, desiredSigs) {
		result := make([]map[string]any, len(oldRules))
		for i, raw := range oldRules {
			result[i] = raw.(map[string]interface{})
		}
		return result, false, nil
	}

	for _, ruleID := range oldIDs {
		if err := rules.Delete(ctx, conn, &rules.DeleteInput{
			RoutingConfigID: &routingConfigID,
			PathID:          &pathID,
			RuleID:          new(ruleID),
		}); err != nil && !isNotFoundErr(err) {
			return nil, false, fmt.Errorf("failed to delete rule %q: %w", ruleID, err)
		}
	}

	result := make([]map[string]any, 0, len(desiredRules))
	for i := range desiredRules {
		action := desiredActions[i]
		created, err := rules.Create(ctx, conn, &rules.CreateInput{
			RoutingConfigID: &routingConfigID,
			PathID:          &pathID,
			Action:          &action,
			Conditions:      desiredConditions[i],
		})
		if err != nil {
			return nil, false, fmt.Errorf("failed to create rule: %w", err)
		}
		result = append(result, flattenRoutingConfigRule(*created))
	}

	return result, true, nil
}

func expandRoutingConfigRule(rm map[string]interface{}) (rules.Action, []rules.Condition) {
	action := rules.Action{
		Type:  rm["action_type"].(string),
		Value: rm["action_value"].(string),
	}

	rawConditions, _ := rm["condition"].([]interface{})
	conditions := make([]rules.Condition, 0, len(rawConditions))
	for _, raw := range rawConditions {
		cm := raw.(map[string]interface{})
		condition := rules.Condition{
			Operator: cm["operator"].(string),
			Type:     cm["type"].(string),
			Value:    cm["value"].(string),
		}
		if key, ok := cm["key"].(string); ok && key != "" {
			condition.Key = &key
		}
		conditions = append(conditions, condition)
	}

	return action, conditions
}

func flattenRoutingConfigRule(r rules.Data) map[string]any {
	conditions := make([]map[string]any, 0, len(r.Conditions))
	for _, c := range r.Conditions {
		cm := map[string]any{
			"operator": c.Operator,
			"type":     c.Type,
			"value":    c.Value,
		}
		if c.Key != nil {
			cm["key"] = *c.Key
		}
		conditions = append(conditions, cm)
	}

	return map[string]any{
		"action_type":  r.Action.Type,
		"action_value": r.Action.Value,
		"condition":    conditions,
		"is_default":   r.IsDefault,
		"rule_id":      r.RuleID,
	}
}

// ruleSignature produces a normalized, order-independent (across conditions)
// signature for a rule's action and conditions so that actual and desired
// rule sets can be compared for equality regardless of API/HCL ordering.
func ruleSignature(action rules.Action, conditions []rules.Condition) (string, error) {
	sorted := make([]rules.Condition, len(conditions))
	copy(sorted, conditions)
	sort.Slice(sorted, func(i, j int) bool {
		ki, kj := "", ""
		if sorted[i].Key != nil {
			ki = *sorted[i].Key
		}
		if sorted[j].Key != nil {
			kj = *sorted[j].Key
		}
		if sorted[i].Type != sorted[j].Type {
			return sorted[i].Type < sorted[j].Type
		}
		if ki != kj {
			return ki < kj
		}
		if sorted[i].Operator != sorted[j].Operator {
			return sorted[i].Operator < sorted[j].Operator
		}
		return sorted[i].Value < sorted[j].Value
	})

	b, err := json.Marshal(struct {
		Action     rules.Action      `json:"action"`
		Conditions []rules.Condition `json:"conditions"`
	}{Action: action, Conditions: sorted})
	if err != nil {
		return "", fmt.Errorf("failed to compute rule signature: %w", err)
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// sameSignatureSet reports whether two signature slices contain the same
// multiset of values, ignoring order.
func sameSignatureSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sortedA := append([]string(nil), a...)
	sortedB := append([]string(nil), b...)
	sort.Strings(sortedA)
	sort.Strings(sortedB)
	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}
	return true
}

// isNotFoundErr reports whether err is a 404 response from the Fastly API.
func isNotFoundErr(err error) bool {
	httpErr, ok := err.(*gofastly.HTTPError)
	return ok && httpErr.StatusCode == 404
}
