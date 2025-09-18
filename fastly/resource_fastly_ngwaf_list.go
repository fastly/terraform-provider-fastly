package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/lists"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

func resourceFastlyNGWAFWorkspaceList() *schema.Resource {
	r := resourceFastlyNGWAFListBase()

	r.Importer = customNGWAFScopeImporter(scope.ScopeTypeWorkspace, "list")

	r.Schema["workspace_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}

	return r
}

func resourceFastlyNGWAFAccountList() *schema.Resource {
	r := resourceFastlyNGWAFListBase()

	r.Importer = customNGWAFScopeImporter(scope.ScopeTypeAccount, "list")

	// Internal field for provider logic compatibility.
	// Not exposed to users but required by shared scope helpers.
	r.Schema["applies_to"] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Sensitive:   true,
		Description: "INTERNAL: Used to build scope for account-scoped lists. Not user-configurable.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return r
}

func resourceFastlyNGWAFListBase() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Fastly Next-Gen WAF list.",
		CreateContext: resourceFastlyNGWAFListCreate,
		ReadContext:   resourceFastlyNGWAFListRead,
		UpdateContext: resourceFastlyNGWAFListUpdate,
		DeleteContext: resourceFastlyNGWAFListDelete,
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the list.",
			},
			"entries": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The values in the list.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the list.",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The type of list. Accepted values are `string`, `wildcard`, `ip`, `country`, and `signal`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"string", "wildcard", "ip", "country", "signal"}, false)),
			},
		},
	}
}

func resourceFastlyNGWAFListCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Inject applies_to=["*"] if it's account-scoped and not explicitly set
	if _, hasWorkspaceID := d.GetOk("workspace_id"); !hasWorkspaceID {
		if _, hasAppliesTo := d.GetOk("applies_to"); !hasAppliesTo {
			if err := d.Set("applies_to", []string{"*"}); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set applies_to during create: %w", err))
			}
		}
	}

	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := expandNGWAFListCreateInput(d, rsc.scope)

	log.Printf("[DEBUG] CREATE: NGWAF %s list input: %#v", rsc.scope.Type, i)

	list, err := lists.Create(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(list.ListID)

	return resourceFastlyNGWAFListRead(ctx, d, meta)
}

func resourceFastlyNGWAFListRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &lists.GetInput{
		ListID: gofastly.ToPointer(d.Id()),
		Scope:  rsc.scope,
	}

	log.Printf("[DEBUG] READ: NGWAF %s list input: %#v", rsc.scope.Type, i)

	list, err := lists.Get(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := flattenNGWAFListResponse(d, list); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFListUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := expandNGWAFListUpdateInput(d, rsc.scope)
	i.ListID = gofastly.ToPointer(d.Id())

	log.Printf("[DEBUG] UPDATE: NGWAF %s list input: %#v", rsc.scope.Type, i)

	_, err = lists.Update(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFListRead(ctx, d, meta)
}

func resourceFastlyNGWAFListDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &lists.DeleteInput{
		ListID: gofastly.ToPointer(d.Id()),
		Scope:  rsc.scope,
	}

	log.Printf("[DEBUG] DELETE: NGWAF %s list input: %#v", rsc.scope.Type, i)

	if err := lists.Delete(rsc.ctx, conn, i); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

// expandNGWAFListCreateInput builds the input for creating a list.
func expandNGWAFListCreateInput(d *schema.ResourceData, scope *scope.Scope) *lists.CreateInput {
	name := d.Get("name").(string)
	listType := d.Get("type").(string)
	entries := expandStringList(d.Get("entries").([]any))

	input := &lists.CreateInput{
		Name:    &name,
		Type:    &listType,
		Entries: &entries,
		Scope:   scope,
	}

	if desc, ok := d.GetOk("description"); ok {
		description := desc.(string)
		input.Description = &description
	}

	return input
}

// expandNGWAFListUpdateInput builds the input for updating a list.
func expandNGWAFListUpdateInput(d *schema.ResourceData, scope *scope.Scope) *lists.UpdateInput {
	input := &lists.UpdateInput{
		ListID: gofastly.ToPointer(d.Id()),
		Scope:  scope,
	}

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
	}
	if d.HasChange("entries") {
		entries := expandStringList(d.Get("entries").([]any))
		input.Entries = &entries
	}

	return input
}

// expandStringList converts a []any from the schema to []string.
func expandStringList(raw []any) []string {
	result := make([]string, len(raw))
	for i, v := range raw {
		result[i] = v.(string)
	}
	return result
}

func flattenNGWAFListResponse(d *schema.ResourceData, list *lists.List) error {
	if list == nil {
		return fmt.Errorf("cannot flatten nil list")
	}
	if list.Scope.Type == "" {
		return fmt.Errorf("invalid list scope: missing type")
	}

	switch list.Scope.Type {
	case string(scope.ScopeTypeWorkspace):
		// For workspace scope, we rely on the original configuration or importer to set `workspace_id`.
		// No need to set it here, as it is not returned in the list response.
	case string(scope.ScopeTypeAccount):
		// Required for internal provider logic (resolveScopeAndContext), even if API doesn't return applies_to
		if err := d.Set("applies_to", []string{"*"}); err != nil {
			return fmt.Errorf("error setting applies_to for account scope: %w", err)
		}

	default:
		return fmt.Errorf("unknown scope type: %q", list.Scope.Type)
	}

	if err := d.Set("name", list.Name); err != nil {
		return fmt.Errorf("error setting name: %w", err)
	}
	if err := d.Set("description", list.Description); err != nil {
		return fmt.Errorf("error setting description: %w", err)
	}
	if err := d.Set("type", list.Type); err != nil {
		return fmt.Errorf("error setting type: %w", err)
	}
	if err := d.Set("entries", list.Entries); err != nil {
		return fmt.Errorf("error setting entries: %w", err)
	}

	return nil
}
