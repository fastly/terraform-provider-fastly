package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceFastlyNGWAFRuleBase() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Fastly Next-Gen WAF rule.",
		CreateContext: resourceFastlyNGWAFRuleCreate,
		ReadContext:   resourceFastlyNGWAFRuleRead,
		UpdateContext: resourceFastlyNGWAFRuleUpdate,
		DeleteContext: resourceFastlyNGWAFRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"action": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of actions to perform when the rule matches.",
				Elem: &schema.Resource{
					Description: "Configuration block for each action.",
					Schema: map[string]*schema.Schema{
						"redirect_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Redirect target (used when `type = redirect`).",
						},
						"response_code": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Response code used with redirect.",
						},
						"signal": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Signal name to exclude (used when `type = exclude_signal`).",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The action type, e.g. `block`, `redirect`, `exclude_signal`.",
						},
					},
				},
			},
			"condition": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Flat list of individual conditions. Each must include `field`, `operator`, and `value`.",
				Elem: &schema.Resource{
					Description: "A single flat condition.",
					Schema: map[string]*schema.Schema{
						"field": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Field to inspect (e.g., `ip`, `path`).",
						},
						"operator": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Operator to apply (e.g., `equals`, `contains`).",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The value to test the field against.",
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A human-readable description of the rule.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the rule is currently enabled.",
			},
			"group_condition": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of grouped conditions with nested logic. Each group must define a `group_operator` and at least one condition.",
				Elem: &schema.Resource{
					Description: "Group of conditions using logical operators.",
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Description: "A list of nested conditions in this group.",
							Elem: &schema.Resource{
								Description: "Nested condition inside a group.",
								Schema: map[string]*schema.Schema{
									"field": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Field to inspect (e.g., `ip`, `path`).",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Operator to apply (e.g., `equals`, `contains`).",
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The value to test the field against.",
									},
								},
							},
						},
						"group_operator": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Logical operator for the group (`any` or `all`).",
						},
					},
				},
			},
			"group_operator": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Logical operator to apply to group conditions (`any` or `all`).",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"any", "all"}, false)),
			},
			"request_logging": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Logging behavior for matching requests (`sampled` or `none`).",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"sampled", "none"}, false)),
			},
			"type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				Description:      "The type of the rule (`request` or `signal`).",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"request", "signal"}, false)),
			},
		},
	}
}
