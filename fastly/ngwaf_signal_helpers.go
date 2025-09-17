package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/signals"
)

func resourceFastlyNGWAFSignalBase() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Fastly Next-Gen WAF signal.",
		CreateContext: resourceFastlyNGWAFSignalCreate,
		ReadContext:   resourceFastlyNGWAFSignalRead,
		UpdateContext: resourceFastlyNGWAFSignalUpdate,
		DeleteContext: resourceFastlyNGWAFSignalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The description of the signal.",
				ValidateFunc: validation.StringLenBetween(0, 140),
			},
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				Description:  "The name of the signal. Special characters and periods are not accepted.",
				ValidateFunc: validation.StringLenBetween(3, 25),
			},
			"reference_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The generated reference ID of the signal.",
			},
		},
	}
}

func flattenNGWAFSignalResponse(d *schema.ResourceData, signal *signals.Signal) error {
	if signal == nil {
		return fmt.Errorf("cannot flatten nil signal")
	}
	if signal.Scope.Type == "" || len(signal.Scope.AppliesTo) == 0 {
		return fmt.Errorf("invalid signal scope: type or applies_to is missing")
	}

	s := signal.Scope

	// Convert the string type to scope.Type for comparison with scope constants.
	t := scope.Type(s.Type)

	switch t {
	case scope.ScopeTypeWorkspace:
		if len(s.AppliesTo) == 0 {
			return fmt.Errorf("workspace scope is missing applies_to ID")
		}
		if err := d.Set("workspace_id", s.AppliesTo[0]); err != nil {
			return fmt.Errorf("error setting workspace_id: %w", err)
		}

	case scope.ScopeTypeAccount:
		if err := d.Set("applies_to", s.AppliesTo); err != nil {
			return fmt.Errorf("error setting applies_to: %w", err)
		}

	default:
		return fmt.Errorf("unknown scope type: %q", s.Type)
	}

	if err := d.Set("description", signal.Description); err != nil {
		return fmt.Errorf("error setting description: %w", err)
	}
	if err := d.Set("name", signal.Name); err != nil {
		return fmt.Errorf("error setting name: %w", err)
	}
	if err := d.Set("reference_id", signal.ReferenceID); err != nil {
		return fmt.Errorf("error setting reference_id: %w", err)
	}

	return nil
}
