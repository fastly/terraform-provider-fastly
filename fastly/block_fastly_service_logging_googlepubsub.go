package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GooglePubSubServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type GooglePubSubServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingGooglePubSub returns a new resource.
func NewServiceLoggingGooglePubSub(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&GooglePubSubServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_googlepubsub",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *GooglePubSubServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *GooglePubSubServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Google Cloud Pub/Sub logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"project_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of your Google Cloud Platform project",
		},
		"secret_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Your Google Cloud Platform account secret key. The `private_key` field in your service account authentication JSON. You may optionally provide this secret via an environment variable, `FASTLY_GOOGLE_PUBSUB_SECRET_KEY`.",
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_GOOGLE_PUBSUB_SECRET_KEY", ""),
			Sensitive:   true,
		},
		"topic": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Google Cloud Pub/Sub topic to which logs will be published",
		},
		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Your Google Cloud Platform service account email address. The `client_email` field in your service account authentication JSON. You may optionally provide this via an environment variable, `FASTLY_GOOGLE_PUBSUB_EMAIL`.",
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_GOOGLE_PUBSUB_EMAIL", ""),
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed.",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
}

// Create creates the resource.
func (h *GooglePubSubServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Google Cloud Pub/Sub logging addition opts: %#v", opts)

	return createGooglePubSub(conn, opts)
}

// Read refreshes the resource.
func (h *GooglePubSubServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Google Cloud Pub/Sub logging endpoints for (%s)", d.Id())
		googlepubsubList, err := conn.ListPubsubs(&gofastly.ListPubsubsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Google Cloud Pub/Sub logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		googlepubsubLogList := flattenGooglePubSub(googlepubsubList)

		for _, element := range googlepubsubLogList {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), googlepubsubLogList); err != nil {
			log.Printf("[WARN] Error setting Google Cloud Pub/Sublogging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *GooglePubSubServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdatePubsubInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["topic"]; ok {
		opts.Topic = gofastly.String(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.String(v.(string))
	}
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.String(v.(string))
	}
	if v, ok := modified["project_id"]; ok {
		opts.ProjectID = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Google Cloud Pub/Sub Opts: %#v", opts)
	_, err := conn.UpdatePubsub(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *GooglePubSubServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Google Cloud Pub/Sub logging endpoint removal opts: %#v", opts)

	return deleteGooglePubSub(conn, opts)
}

func createGooglePubSub(conn *gofastly.Client, i *gofastly.CreatePubsubInput) error {
	_, err := conn.CreatePubsub(i)
	return err
}

func deleteGooglePubSub(conn *gofastly.Client, i *gofastly.DeletePubsubInput) error {
	err := conn.DeletePubsub(i)

	errRes, ok := err.(*gofastly.HTTPError)
	if !ok {
		return err
	}

	// 404 response codes don't result in an error propagating because a 404 could
	// indicate that a resource was deleted elsewhere.
	if !errRes.IsNotFound() {
		return err
	}

	return nil
}

func flattenGooglePubSub(googlepubsubList []*gofastly.Pubsub) []map[string]any {
	var flattened []map[string]any
	for _, s := range googlepubsubList {
		// Convert logging to a map for saving to state.
		flatGooglePubSub := map[string]any{
			"name":               s.Name,
			"user":               s.User,
			"secret_key":         s.SecretKey,
			"project_id":         s.ProjectID,
			"topic":              s.Topic,
			"format":             s.Format,
			"format_version":     s.FormatVersion,
			"placement":          s.Placement,
			"response_condition": s.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range flatGooglePubSub {
			if v == "" {
				delete(flatGooglePubSub, k)
			}
		}

		flattened = append(flattened, flatGooglePubSub)
	}

	return flattened
}

func (h *GooglePubSubServiceAttributeHandler) buildCreate(googlepubsubMap any, serviceID string, serviceVersion int) *gofastly.CreatePubsubInput {
	df := googlepubsubMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(df)
	opts := &gofastly.CreatePubsubInput{
		Format:         gofastly.String(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.String(df["name"].(string)),
		Placement:      gofastly.String(vla.placement),
		ProjectID:      gofastly.String(df["project_id"].(string)),
		SecretKey:      gofastly.String(df["secret_key"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Topic:          gofastly.String(df["topic"].(string)),
		User:           gofastly.String(df["user"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	// if vla.placement != "" {
	// 	opts.Placement = gofastly.String(vla.placement)
	// }
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.String(vla.responseCondition)
	}

	return opts
}

func (h *GooglePubSubServiceAttributeHandler) buildDelete(googlepubsubMap any, serviceID string, serviceVersion int) *gofastly.DeletePubsubInput {
	df := googlepubsubMap.(map[string]any)

	return &gofastly.DeletePubsubInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
