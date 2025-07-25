package fastly

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v11/fastly"
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/rules"
)

func resourceFastlyNGWAFRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i, err := expandNGWAFRuleCreateInput(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := rules.Create(ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.RuleID)

	return resourceFastlyNGWAFRuleRead(ctx, d, meta)
}

func resourceFastlyNGWAFRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i, err := expandNGWAFRuleReadInput(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := rules.Get(ctx, conn, i)
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

	i, err := expandNGWAFRuleUpdateInput(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = rules.Update(ctx, conn, i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFRuleRead(ctx, d, meta)
}

func resourceFastlyNGWAFRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := rules.Delete(ctx, conn, &rules.DeleteInput{
		RuleID: fastly.ToPointer(d.Id()),
		Scope:  buildNGWAFRuleScope(d),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
