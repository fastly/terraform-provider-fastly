package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ElasticSearchServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingElasticSearch(sa ServiceMetadata) ServiceAttributeDefinition {
	return &ElasticSearchServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_elasticsearch",
			serviceMetadata: sa,
		},
	}
}

func (h *ElasticSearchServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	oe, ne := d.GetChange(h.GetKey())

	if oe == nil {
		oe = new(schema.Set)
	}
	if ne == nil {
		ne = new(schema.Set)
	}

	oes := oe.(*schema.Set)
	nes := ne.(*schema.Set)

	setDiff := NewSetDiff(func(logging interface{}) (interface{}, error) {
		// Use the logging endpoint name as the key
		return logging.(map[string]interface{})["name"], nil
	})

	diffResult, err := setDiff.Diff(oes, nes)
	if err != nil {
		return err
	}

	// DELETE old Elasticsearch logging endpoints.
	for _, oRaw := range diffResult.Deleted {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Elasticsearch logging endpoint removal opts: %#v", opts)

		if err := deleteElasticsearch(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Elasticsearch logging endpoints.
	for _, nRaw := range diffResult.Added {
		ef := nRaw.(map[string]interface{})
		opts := h.buildCreate(ef, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Elasticsearch logging addition opts: %#v", opts)

		if err := createElasticsearch(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *ElasticSearchServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh Elasticsearch.
	log.Printf("[DEBUG] Refreshing Elasticsearch logging endpoints for (%s)", d.Id())
	elasticsearchList, err := conn.ListElasticsearch(&gofastly.ListElasticsearchInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Elasticsearch logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenElasticsearch(elasticsearchList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting Elasticsearch logging endpoints for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *ElasticSearchServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Elasticsearch logging endpoint",
		},

		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Elasticsearch URL to stream logs to",
		},

		"index": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Elasticsearch index to send documents (logs) to",
		},

		// Optional fields
		"pipeline": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing",
		},

		"user": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "BasicAuth username for Elasticsearch",
		},

		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "BasicAuth password for Elasticsearch",
			Sensitive:   true,
		},

		"request_max_entries": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "The maximum number of bytes sent in one request. Defaults to `0` for unbounded",
		},

		"request_max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "The maximum number of logs sent in one request. Defaults to `0` for unbounded",
		},

		"tls_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A secure certificate to authenticate the server with. Must be in PEM format",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client certificate used to make authenticated requests. Must be in PEM format",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client private key used to make authenticated requests. Must be in PEM format",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The hostname used to verify the server's certificate. It can either be the Common Name (CN) or a Subject Alternative Name (SAN)",
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
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed.",
			ValidateFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply",
		}
	}

	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
	return nil
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

func flattenElasticsearch(elasticsearchList []*gofastly.Elasticsearch) []map[string]interface{} {
	var esl []map[string]interface{}
	for _, el := range elasticsearchList {
		// Convert Elasticsearch logging to a map for saving to state.
		nel := map[string]interface{}{
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

func (h *ElasticSearchServiceAttributeHandler) buildCreate(elasticsearchMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateElasticsearchInput {
	df := elasticsearchMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateElasticsearchInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		Index:             df["index"].(string),
		URL:               df["url"].(string),
		Pipeline:          df["pipeline"].(string),
		User:              df["user"].(string),
		Password:          df["password"].(string),
		RequestMaxEntries: uint(df["request_max_entries"].(int)),
		RequestMaxBytes:   uint(df["request_max_bytes"].(int)),
		TLSCACert:         df["tls_ca_cert"].(string),
		TLSClientCert:     df["tls_client_cert"].(string),
		TLSClientKey:      df["tls_client_key"].(string),
		TLSHostname:       df["tls_hostname"].(string),
		Format:            vla.format,
		FormatVersion:     uintOrDefault(vla.formatVersion),
		Placement:         vla.placement,
		ResponseCondition: vla.responseCondition,
	}
}

func (h *ElasticSearchServiceAttributeHandler) buildDelete(elasticsearchMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteElasticsearchInput {
	df := elasticsearchMap.(map[string]interface{})

	return &gofastly.DeleteElasticsearchInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
