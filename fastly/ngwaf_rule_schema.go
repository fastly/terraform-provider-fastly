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
				Description: "The description of the rule.",
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
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Logical operator for the group. Accepted values are `any` and `all`.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"any", "all"}, false)),
						},
					},
				},
			},
			"group_operator": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Logical operator to apply to group conditions. Accepted values are `any` and `all`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"any", "all"}, false)),
			},
			"rate_limit": {
				Description: "Block specifically for rate_limit rules.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_identifiers": {
							Description: "List of client identifiers used for rate limiting. Can only be length 1 or 2.",
							Type:        schema.TypeSet,
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Key for the Client Identifier.",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name for the Client Identifier.",
									},
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of the Client Identifier.",
									},
								},
							},
						},
						"duration": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Duration in seconds for the rate limit.",
						},
						"interval": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Time interval for the rate limit in seconds. Accepted values are 60, 600, and 3600 (seconds).",
							ValidateFunc: validation.IntInSlice([]int{60, 600, 3600}),
						},
						"signal": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Reference ID of the custom signal this rule uses to count requests.",
						},
						"threshold": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Rate limit threshold. Minimum 1 and maximum 10,000.",
							ValidateFunc: validation.IntBetween(1, 10000),
						},
					},
				},
			},
			"request_logging": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Logging behavior for matching requests. Accepted values are `sampled` and `none`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"sampled", "none"}, false)),
			},
			"type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				Description:      "The type of the rule. Accepted values are `request` and `signal`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"request", "signal"}, false)),
			},
		},
	}
}
