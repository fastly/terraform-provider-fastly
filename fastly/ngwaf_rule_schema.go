package fastly

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
		CustomizeDiff: customdiff.All(
			validateGroupConditionNotEmpty,
		),
		Schema: map[string]*schema.Schema{
			"action": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of actions to perform when the rule matches.",
				Elem: &schema.Resource{
					Description: "Configuration block for each action.",
					Schema: map[string]*schema.Schema{
						"allow_interactive": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Specifies if interaction is allowed (used when `type = browser_challenge`).",
						},
						"deception_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "specifies the type of deception (used when `type = deception`).",
						},
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
							Description: "The action type. One of: `add_signal`, `allow`, `block`, `browser_challenge`, `dynamic_challenge`, `exclude_signal`, `verify_token` or for rate limit rule valid values: `log_request`, `block_signal`, `browser_challenge`, `verify_token`",
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
				Description: "List of grouped conditions with nested logic. Each group must define a `group_operator` and at least one condition or multival_condition.",
				Elem: &schema.Resource{
					Description: "Group of conditions using logical operators.",
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:        schema.TypeList,
							Optional:    true,
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
						"multival_condition": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of nested multival conditions in this group. Each multival list must define a `field, operator, group_operator` and at least one condition.",
							Elem: &schema.Resource{
								Description: "Nested multival condition inside a group.",
								Schema: map[string]*schema.Schema{
									"condition": {
										Type:        schema.TypeList,
										Required:    true,
										MinItems:    1,
										Description: "A list of nested conditions in this multival.",
										Elem: &schema.Resource{
											Description: "Nested condition inside a multival.",
											Schema: map[string]*schema.Schema{
												"field": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Field to inspect (e.g., `name`, `value`, `signal_id`).",
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
									"field": {
										Type:             schema.TypeString,
										Required:         true,
										Description:      "Enums for multival condition field. Accepted values are `post_parameter`, `query_parameter`, `request_cookie`, `request_header`, `response_header`, and `signal`.",
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"post_parameter", "query_parameter", "request_cookie", "request_header", "response_header", "signal"}, false)),
									},
									"group_operator": {
										Type:             schema.TypeString,
										Required:         true,
										Description:      "Logical operator for the multival condition. Accepted values are `any` and `all`.",
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"any", "all"}, false)),
									},
									"operator": {
										Type:             schema.TypeString,
										Required:         true,
										Description:      "Indicates whether the supplied conditions will check for existence or non-existence of matching field values. Accepted values are `exists` and `does_not_exist`.",
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"exists", "does_not_exist"}, false)),
									},
								},
							},
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
			"multival_condition": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of multival conditions with nested logic. Each multival list must define a `field, operator, group_operator` and at least one condition.",
				Elem: &schema.Resource{
					Description: "Group of conditions using logical operators.",
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Description: "A list of nested conditions in this list.",
							Elem: &schema.Resource{
								Description: "Nested condition inside a group.",
								Schema: map[string]*schema.Schema{
									"field": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Field to inspect (e.g., `name`, `value`, `signal_id`).",
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
						"field": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Enums for multival condition field.. Accepted values are `post_parameter`, `query_parameter`, `request_cookie`, `request_header`, `response_header`, and `signal`.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"post_parameter", "query_parameter", "request_cookie", "request_header", "response_header", "signal"}, false)),
						},
						"group_operator": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Logical operator for the group. Accepted values are `any` and `all`.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"any", "all"}, false)),
						},
						"operator": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Indicates whether the supplied conditions will check for existence or non-existence of matching field values. Accepted values are `exists` and `does_not_exist`.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"exists", "does_not_exist"}, false)),
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

// validateGroupConditionNotEmpty ensures that each group_condition block
// contains at least one of: condition or multival_condition.
func validateGroupConditionNotEmpty(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	groupConditions, ok := diff.GetOk("group_condition")
	if !ok {
		return nil
	}

	groupConditionList := groupConditions.([]interface{})

	for i, gc := range groupConditionList {
		groupConditionMap := gc.(map[string]interface{})

		// Check for both condition and multival_condition
		conditions, condExists := groupConditionMap["condition"]
		multivalConditions, multivalExists := groupConditionMap["multival_condition"]

		// Count non-nil, non-empty lists
		condLen := 0
		if condExists && conditions != nil {
			if condList, ok := conditions.([]interface{}); ok {
				condLen = len(condList)
			}
		}

		multivalLen := 0
		if multivalExists && multivalConditions != nil {
			if multivalList, ok := multivalConditions.([]interface{}); ok {
				multivalLen = len(multivalList)
			}
		}

		// Both must be zero (empty or missing) to be invalid
		if condLen == 0 && multivalLen == 0 {
			return fmt.Errorf("group_condition[%d]: must define at least one 'condition' or 'multival_condition' block", i)
		}
	}

	return nil
}
