package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/signals"
)

func resourceFastlyNGWAFWorkspaceSignal() *schema.Resource {
	r := resourceFastlyNGWAFSignalBase()

	r.Importer = customNGWAFScopeImporter(scope.ScopeTypeWorkspace, "signal")

	r.Schema["workspace_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the workspace.",
	}

	return r
}

func resourceFastlyNGWAFAccountSignal() *schema.Resource {
	r := resourceFastlyNGWAFSignalBase()

	r.Importer = customNGWAFScopeImporter(scope.ScopeTypeAccount, "signal")

	r.Schema["applies_to"] = &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: "The list of workspace IDs this signal applies to, or the wildcard `*` if it applies to all workspaces.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return r
}

func resourceFastlyNGWAFSignalCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &signals.CreateInput{
		Description: fastly.ToPointer(d.Get("description").(string)),
		Name:        fastly.ToPointer(d.Get("name").(string)),
		Scope:       rsc.scope,
	}

	log.Printf("[DEBUG] CREATE: NGWAF %s signal input: %#v", rsc.scope.Type, i)

	r, err := signals.Create(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.SignalID)

	return resourceFastlyNGWAFSignalRead(ctx, d, meta)
}

func resourceFastlyNGWAFSignalRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &signals.GetInput{
		SignalID: fastly.ToPointer(d.Id()),
		Scope:    rsc.scope,
	}

	log.Printf("[DEBUG] REFRESH: NGWAF %s signal input: %#v", rsc.scope.Type, i)

	r, err := signals.Get(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := flattenNGWAFSignalResponse(d, r); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFSignalUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &signals.UpdateInput{
		Description: fastly.ToPointer(d.Get("description").(string)),
		SignalID:    fastly.ToPointer(d.Id()),
		Scope:       rsc.scope,
	}

	log.Printf("[DEBUG] UPDATE: NGWAF %s signal input: %#v", rsc.scope.Type, i)

	_, err = signals.Update(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFSignalRead(ctx, d, meta)
}

func resourceFastlyNGWAFSignalDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &signals.DeleteInput{
		SignalID: fastly.ToPointer(d.Id()),
		Scope:    rsc.scope,
	}

	log.Printf("[DEBUG] DELETE: NGWAF %s signal scope type: %#v", rsc.scope.Type, i.Scope.Type)
	log.Printf("[DEBUG] DELETE: NGWAF %s signal scope applies to: %#v", rsc.scope.Type, i.Scope.AppliesTo)
	log.Printf("[DEBUG] DELETE: NGWAF %s signal ID: %s", rsc.scope.Type, d.Id())

	err = signals.Delete(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
