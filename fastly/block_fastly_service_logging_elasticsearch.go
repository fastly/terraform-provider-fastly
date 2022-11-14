package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			Sensitive:   true,
		},
		"pipeline": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing",
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
			Sensitive:        true,
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
			Default:     "%h %l %u %t \"%r\" %>s %b",
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
func (h *ElasticSearchServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Elasticsearch logging addition opts: %#v", opts)

	return createElasticsearch(conn, opts)
}

// Read refreshes the resource.
func (h *ElasticSearchServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Elasticsearch logging endpoints for (%s)", d.Id())
		elasticsearchList, err := conn.ListElasticsearch(&gofastly.ListElasticsearchInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Elasticsearch logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenElasticsearch(elasticsearchList)

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
func (h *ElasticSearchServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateElasticsearchInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["index"]; ok {
		opts.Index = gofastly.String(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.String(v.(string))
	}
	if v, ok := modified["pipeline"]; ok {
		opts.Pipeline = gofastly.String(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.String(v.(string))
	}
	if v, ok := modified["password"]; ok {
		opts.Password = gofastly.String(v.(string))
	}
	if v, ok := modified["request_max_entries"]; ok {
		opts.RequestMaxEntries = gofastly.Int(v.(int))
	}
	if v, ok := modified["request_max_bytes"]; ok {
		opts.RequestMaxBytes = gofastly.Int(v.(int))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_ca_cert"]; ok {
		opts.TLSCACert = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_client_cert"]; ok {
		opts.TLSClientCert = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_client_key"]; ok {
		opts.TLSClientKey = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_hostname"]; ok {
		opts.TLSHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}

	log.Printf("[DEBUG] Update Elasticsearch Opts: %#v", opts)
	_, err := conn.UpdateElasticsearch(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *ElasticSearchServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Elasticsearch logging endpoint removal opts: %#v", opts)

	return deleteElasticsearch(conn, opts)
}

func createElasticsearch(conn *gofastly.Client, i *gofastly.CreateElasticsearchInput) error {
	_, err := conn.CreateElasticsearch(i)
	return err
}

func deleteElasticsearch(conn *gofastly.Client, i *gofastly.DeleteElasticsearchInput) error {
	err := conn.DeleteElasticsearch(i)

	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func flattenElasticsearch(elasticsearchList []*gofastly.Elasticsearch) []map[string]any {
	var esl []map[string]any
	for _, el := range elasticsearchList {
		// Convert Elasticsearch logging to a map for saving to state.
		nel := map[string]any{
			"name":                el.Name,
			"response_condition":  el.ResponseCondition,
			"format":              el.Format,
			"index":               el.Index,
			"url":                 el.URL,
			"pipeline":            el.Pipeline,
			"user":                el.User,
			"password":            el.Password,
			"request_max_entries": el.RequestMaxEntries,
			"request_max_bytes":   el.RequestMaxBytes,
			"placement":           el.Placement,
			"tls_ca_cert":         el.TLSCACert,
			"tls_client_cert":     el.TLSClientCert,
			"tls_client_key":      el.TLSClientKey,
			"tls_hostname":        el.TLSHostname,
			"format_version":      el.FormatVersion,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range nel {
			if v == "" {
				delete(nel, k)
			}
		}

		esl = append(esl, nel)
	}

	return esl
}

func (h *ElasticSearchServiceAttributeHandler) buildCreate(elasticsearchMap any, serviceID string, serviceVersion int) *gofastly.CreateElasticsearchInput {
	resource := elasticsearchMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateElasticsearchInput{
		Format:            gofastly.String(vla.format),
		FormatVersion:     vla.formatVersion,
		Index:             gofastly.String(resource["index"].(string)),
		Name:              gofastly.String(resource["name"].(string)),
		Password:          gofastly.String(resource["password"].(string)),
		Pipeline:          gofastly.String(resource["pipeline"].(string)),
		Placement:         gofastly.String(vla.placement),
		RequestMaxBytes:   gofastly.Int(resource["request_max_bytes"].(int)),
		RequestMaxEntries: gofastly.Int(resource["request_max_entries"].(int)),
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		TLSCACert:         gofastly.String(resource["tls_ca_cert"].(string)),
		TLSClientCert:     gofastly.String(resource["tls_client_cert"].(string)),
		TLSClientKey:      gofastly.String(resource["tls_client_key"].(string)),
		TLSHostname:       gofastly.String(resource["tls_hostname"].(string)),
		URL:               gofastly.String(resource["url"].(string)),
		User:              gofastly.String(resource["user"].(string)),
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

func (h *ElasticSearchServiceAttributeHandler) buildDelete(elasticsearchMap any, serviceID string, serviceVersion int) *gofastly.DeleteElasticsearchInput {
	resource := elasticsearchMap.(map[string]any)

	return &gofastly.DeleteElasticsearchInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
