package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

// ElasticSearchServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ElasticSearchServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingElasticSearch returns a new resource.
func NewServiceLoggingElasticSearch(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ElasticSearchServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_elasticsearch",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *ElasticSearchServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ElasticSearchServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"index": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Elasticsearch index to send documents (logs) to",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Elasticsearch logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "BasicAuth password for Elasticsearch",
			Sensitive:   !DisplaySensitiveFields,
		},
		"pipeline": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing",
		},
		"processing_region": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "none",
			Description:  "Region where logs will be processed before streaming to BigQuery. Valid values are 'none', 'us' and 'eu'.",
			ValidateFunc: validation.StringInSlice([]string{"none", "us", "eu"}, false),
		},
		"request_max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "The maximum number of logs sent in one request. Defaults to `0` for unbounded",
		},
		"request_max_entries": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "The maximum number of bytes sent in one request. Defaults to `0` for unbounded",
		},
		"tls_ca_cert": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A secure certificate to authenticate the server with. Must be in PEM format",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"tls_client_cert": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The client certificate used to make authenticated requests. Must be in PEM format",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"tls_client_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The client private key used to make authenticated requests. Must be in PEM format",
			Sensitive:        !DisplaySensitiveFields,
			ValidateDiagFunc: validateStringTrimmed,
		},
		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The hostname used to verify the server's certificate. It can either be the Common Name (CN) or a Subject Alternative Name (SAN)",
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Elasticsearch URL to stream logs to",
		},
		"user": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "BasicAuth username for Elasticsearch",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     LoggingElasticsearchDefaultFormat,
			Description: "Apache-style string or VCL variables to use for log formatting.",
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
			Description: "The name of the condition to apply",
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
func (h *ElasticSearchServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Elasticsearch logging addition opts: %#v", opts)

	_, err := conn.CreateElasticsearch(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	return err
}

// Read refreshes the resource.
func (h *ElasticSearchServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Elasticsearch logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListElasticsearch(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListElasticsearchInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Elasticsearch logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenElasticsearch(remoteState)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting Elasticsearch logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *ElasticSearchServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateElasticsearchInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["index"]; ok {
		opts.Index = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["pipeline"]; ok {
		opts.Pipeline = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["password"]; ok {
		opts.Password = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["request_max_entries"]; ok {
		opts.RequestMaxEntries = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["request_max_bytes"]; ok {
		opts.RequestMaxBytes = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_ca_cert"]; ok {
		opts.TLSCACert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_client_cert"]; ok {
		opts.TLSClientCert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_client_key"]; ok {
		opts.TLSClientKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_hostname"]; ok {
		opts.TLSHostname = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Elasticsearch Opts: %#v", opts)
	_, err := conn.UpdateElasticsearch(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *ElasticSearchServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Elasticsearch logging endpoint removal opts: %#v", opts)

	err := conn.DeleteElasticsearch(gofastly.NewContextForResourceID(ctx, d.Id()), opts)

	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenElasticsearch models data into format suitable for saving to Terraform state.
func flattenElasticsearch(remoteState []*gofastly.Elasticsearch) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.Index != nil {
			data["index"] = *resource.Index
		}
		if resource.URL != nil {
			data["url"] = *resource.URL
		}
		if resource.Pipeline != nil {
			data["pipeline"] = *resource.Pipeline
		}
		if resource.User != nil {
			data["user"] = *resource.User
		}
		if resource.Password != nil {
			data["password"] = *resource.Password
		}
		if resource.RequestMaxEntries != nil {
			data["request_max_entries"] = *resource.RequestMaxEntries
		}
		if resource.RequestMaxBytes != nil {
			data["request_max_bytes"] = *resource.RequestMaxBytes
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.TLSCACert != nil {
			data["tls_ca_cert"] = *resource.TLSCACert
		}
		if resource.TLSClientCert != nil {
			data["tls_client_cert"] = *resource.TLSClientCert
		}
		if resource.TLSClientKey != nil {
			data["tls_client_key"] = *resource.TLSClientKey
		}
		if resource.TLSHostname != nil {
			data["tls_hostname"] = *resource.TLSHostname
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
		}
		if resource.ProcessingRegion != nil {
			data["processing_region"] = *resource.ProcessingRegion
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func (h *ElasticSearchServiceAttributeHandler) buildCreate(elasticsearchMap any, serviceID string, serviceVersion int) *gofastly.CreateElasticsearchInput {
	resource := elasticsearchMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateElasticsearchInput{
		Format:            gofastly.ToPointer(vla.format),
		FormatVersion:     vla.formatVersion,
		Index:             gofastly.ToPointer(resource["index"].(string)),
		Name:              gofastly.ToPointer(resource["name"].(string)),
		Password:          gofastly.ToPointer(resource["password"].(string)),
		Pipeline:          gofastly.ToPointer(resource["pipeline"].(string)),
		RequestMaxBytes:   gofastly.ToPointer(resource["request_max_bytes"].(int)),
		RequestMaxEntries: gofastly.ToPointer(resource["request_max_entries"].(int)),
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		TLSCACert:         gofastly.ToPointer(resource["tls_ca_cert"].(string)),
		TLSClientCert:     gofastly.ToPointer(resource["tls_client_cert"].(string)),
		TLSClientKey:      gofastly.ToPointer(resource["tls_client_key"].(string)),
		TLSHostname:       gofastly.ToPointer(resource["tls_hostname"].(string)),
		URL:               gofastly.ToPointer(resource["url"].(string)),
		User:              gofastly.ToPointer(resource["user"].(string)),
		ProcessingRegion:  gofastly.ToPointer(resource["processing_region"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.ToPointer(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.ToPointer(vla.responseCondition)
	}

	return opts
}

func (h *ElasticSearchServiceAttributeHandler) buildDelete(elasticsearchMap any, serviceID string, serviceVersion int) *gofastly.DeleteElasticsearchInput {
	resource := elasticsearchMap.(map[string]any)

	return &gofastly.DeleteElasticsearchInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
