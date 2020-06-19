package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var elasticsearchSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique name of the Elasticsearch logging endpoint.",
			},

			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Elasticsearch URL to stream logs to.",
			},

			"index": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Elasticsearch index to send documents (logs) to.",
			},

			// Optional fields
			"pipeline": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing.",
			},

			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "BasicAuth user.",
			},

			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "BasicAuth password.",
				Sensitive:   true,
			},

			"request_max_entries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The maximum number of logs sent in one request.",
			},

			"request_max_bytes": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The maximum number of bytes sent in one request.",
			},

			"format": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Apache-style string or VCL variables to use for log formatting.",
			},

			"format_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
				ValidateFunc: validateLoggingFormatVersion(),
			},

			"tls_ca_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A secure certificate to authenticate the server with. Must be in PEM format.",
				Sensitive:   true,
				// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
				StateFunc: trimSpaceStateFunc,
			},

			"tls_client_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The client certificate used to make authenticated requests. Must be in PEM format.",
				Sensitive:   true,
				// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
				StateFunc: trimSpaceStateFunc,
			},

			"tls_client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The client private key used to make authenticated requests. Must be in PEM format.",
				Sensitive:   true,
				// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
				StateFunc: trimSpaceStateFunc,
			},

			"tls_hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The hostname used to verify the server's certificate. It can either be the Common Name (CN) or a Subject Alternative Name (SAN).",
			},

			"placement": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Where in the generated VCL the logging call should be placed.",
				ValidateFunc: validateLoggingPlacement(),
			},

			"response_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the condition to apply",
			},
		},
	},
}

func processElasticsearch(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	serviceID := d.Id()
	oe, ne := d.GetChange("logging_elasticsearch")

	if oe == nil {
		oe = new(schema.Set)
	}
	if ne == nil {
		ne = new(schema.Set)
	}

	oes := oe.(*schema.Set)
	nes := ne.(*schema.Set)

	removeElasticsearchLogging := oes.Difference(nes).List()
	addElasticsearchLogging := nes.Difference(oes).List()

	// DELETE old Elasticsearch logging endpoints.
	for _, oRaw := range removeElasticsearchLogging {
		of := oRaw.(map[string]interface{})
		opts := buildDeleteElasticsearch(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Elasticsearch logging endpoint removal opts: %#v", opts)

		if err := deleteElasticsearch(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Elasticsearch logging endpoints.
	for _, nRaw := range addElasticsearchLogging {
		ef := nRaw.(map[string]interface{})

		// @HACK for a TF SDK Issue.
		//
		// This ensures that the required, `name`, field is present.
		//
		// If we have made it this far and `name` is not present, it is most-likely due
		// to a defunct diff as noted here - https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697.
		//
		// This is caused by using a StateFunc in a nested TypeSet. While the StateFunc
		// properly handles setting state with the StateFunc, it returns extra entries
		// during state Gets, specifically `GetChange("logging_elasticsearch")` in this case.
		if v, ok := ef["name"]; !ok || v.(string) == "" {
			continue
		}

		opts := buildCreateElasticsearch(ef, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Elasticsearch logging addition opts: %#v", opts)

		if err := createElasticsearch(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func readElasticsearch(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	// Refresh Elasticsearch.
	log.Printf("[DEBUG] Refreshing Elasticsearch logging endpoints for (%s)", d.Id())
	elasticsearchList, err := conn.ListElasticsearch(&gofastly.ListElasticsearchInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Elasticsearch logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenElasticsearch(elasticsearchList)

	if err := d.Set("logging_elasticsearch", ell); err != nil {
		log.Printf("[WARN] Error setting Elasticsearch logging endpoints for (%s): %s", d.Id(), err)
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

func buildCreateElasticsearch(elasticsearchMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateElasticsearchInput {
	df := elasticsearchMap.(map[string]interface{})

	return &gofastly.CreateElasticsearchInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              gofastly.NullString(df["name"].(string)),
		Index:             gofastly.NullString(df["index"].(string)),
		URL:               gofastly.NullString(df["url"].(string)),
		Pipeline:          gofastly.NullString(df["pipeline"].(string)),
		User:              gofastly.NullString(df["user"].(string)),
		Password:          gofastly.NullString(df["password"].(string)),
		RequestMaxEntries: gofastly.Uint(uint(df["request_max_entries"].(int))),
		RequestMaxBytes:   gofastly.Uint(uint(df["request_max_bytes"].(int))),
		TLSCACert:         gofastly.NullString(df["tls_ca_cert"].(string)),
		TLSClientCert:     gofastly.NullString(df["tls_client_cert"].(string)),
		TLSClientKey:      gofastly.NullString(df["tls_client_key"].(string)),
		TLSHostname:       gofastly.NullString(df["tls_hostname"].(string)),
		Format:            gofastly.NullString(df["format"].(string)),
		FormatVersion:     gofastly.Uint(uint(df["format_version"].(int))),
		Placement:         gofastly.NullString(df["placement"].(string)),
		ResponseCondition: gofastly.NullString(df["response_condition"].(string)),
	}
}

func buildDeleteElasticsearch(elasticsearchMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteElasticsearchInput {
	df := elasticsearchMap.(map[string]interface{})

	return &gofastly.DeleteElasticsearchInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
