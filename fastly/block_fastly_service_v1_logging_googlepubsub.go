package fastly

import (
	"fmt"
	"log"

	"github.com/fastly/go-fastly/fastly"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var googlepubsubloggingSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique name of the Google Cloud Pub/Sub logging endpoint.",
			},

			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON. ",
			},

			"secret_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON.",
			},

			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of your Google Cloud Platform project.",
			},

			"topic": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Google Cloud Pub/Sub topic to which logs will be published.",
			},

			// Optional
			"format": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Apache style log formatting.",
			},

			"format_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
				ValidateFunc: validateLoggingFormatVersion(),
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
				Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
			},
		},
	},
}

func processGooglePubSub(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	serviceID := d.Id()
	oldLogCfg, newLogCfg := d.GetChange("logging_googlepubsub")

	if oldLogCfg == nil {
		oldLogCfg = new(schema.Set)
	}
	if newLogCfg == nil {
		newLogCfg = new(schema.Set)
	}

	oldLogSet := oldLogCfg.(*schema.Set)
	newLogSet := newLogCfg.(*schema.Set)

	removeGooglePubSubLogging := oldLogSet.Difference(newLogSet).List()
	addGooglePubSubLogging := newLogSet.Difference(oldLogSet).List()

	// DELETE old Google Cloud Pub/Sub logging endpoints.
	for _, oRaw := range removeGooglePubSubLogging {
		of := oRaw.(map[string]interface{})
		opts := buildDeleteGooglePubSub(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Google Cloud Pub/Sub logging endpoint removal opts: %#v", opts)

		if err := deleteGooglePubSub(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Google Cloud Pub/Sub logging endponts.
	for _, nRaw := range addGooglePubSubLogging {
		cfg := nRaw.(map[string]interface{})
		opts := buildCreateGooglePubSub(cfg, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Google Cloud Pub/Sub logging addition opts: %#v", opts)

		if err := createGooglePubSub(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func readGooglePubSub(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	// Refresh Google Cloud Pub/Sub logging endpoints.
	log.Printf("[DEBUG] Refreshing Google Cloud Pub/Sub logging endpoints for (%s)", d.Id())
	googlepubsubList, err := conn.ListPubsubs(&gofastly.ListPubsubsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Google Cloud Pub/Sub logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	googlepubsubLogList := flattenGooglePubSub(googlepubsubList)

	if err := d.Set("logging_googlepubsub", googlepubsubLogList); err != nil {
		log.Printf("[WARN] Error setting Google Cloud Pub/Sublogging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
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

func flattenGooglePubSub(googlepubsubList []*gofastly.Pubsub) []map[string]interface{} {
	var flattened []map[string]interface{}
	for _, s := range googlepubsubList {
		// Convert logging to a map for saving to state.
		flatGooglePubSub := map[string]interface{}{
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

func buildCreateGooglePubSub(googlepubsubMap interface{}, serviceID string, serviceVersion int) *gofastly.CreatePubsubInput {
	df := googlepubsubMap.(map[string]interface{})

	return &gofastly.CreatePubsubInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              fastly.NullString(df["name"].(string)),
		User:              fastly.NullString(df["user"].(string)),
		SecretKey:         fastly.NullString(df["secret_key"].(string)),
		ProjectID:         fastly.NullString(df["project_id"].(string)),
		Topic:             fastly.NullString(df["topic"].(string)),
		Format:            fastly.NullString(df["format"].(string)),
		FormatVersion:     fastly.Uint(uint(df["format_version"].(int))),
		ResponseCondition: fastly.NullString(df["response_condition"].(string)),
		Placement:         fastly.NullString(df["placement"].(string)),
	}
}

func buildDeleteGooglePubSub(googlepubsubMap interface{}, serviceID string, serviceVersion int) *gofastly.DeletePubsubInput {
	df := googlepubsubMap.(map[string]interface{})

	return &gofastly.DeletePubsubInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
