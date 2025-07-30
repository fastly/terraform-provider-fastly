package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v11/fastly"
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/rules"
)

func resourceFastlyNGWAFRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i, err := expandNGWAFRuleCreateInput(d, rsc.scope)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] CREATE: NGWAF %s rule input: %#v", rsc.scope.Type, i)

	r, err := rules.Create(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.RuleID)

	return resourceFastlyNGWAFRuleRead(ctx, d, meta)
}

func resourceFastlyNGWAFRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &rules.GetInput{
		RuleID: fastly.ToPointer(d.Id()),
		Scope:  rsc.scope,
	}

	log.Printf("[DEBUG] REFRESH: NGWAF %s rule input: %#v", rsc.scope.Type, i)

	r, err := rules.Get(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := flattenNGWAFRuleResponse(d, r); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i, err := expandNGWAFRuleUpdateInput(d, rsc.scope)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] UPDATE: NGWAF %s rule input: %#v", rsc.scope.Type, i)

	_, err = rules.Update(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFRuleRead(ctx, d, meta)
}

func resourceFastlyNGWAFRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	rsc, err := resolveScopeAndContext(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	i := &rules.DeleteInput{
		RuleID: fastly.ToPointer(d.Id()),
		Scope:  rsc.scope,
	}

	log.Printf("[DEBUG] DELETE: NGWAF %s rule input: %#v", rsc.scope.Type, i)

	err = rules.Delete(rsc.ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
