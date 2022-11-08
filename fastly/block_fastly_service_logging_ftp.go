package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FTPServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type FTPServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingFTP returns a new resource.
func NewServiceLoggingFTP(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&FTPServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_ftp",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *FTPServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *FTPServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The FTP address to stream logs to",
		},
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
		},
		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: GzipLevelDescription,
		},
		"message_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "classic",
			Description:      MessageTypeDescription,
			ValidateDiagFunc: validateLoggingMessageType(),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the FTP logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"password": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The password for the server (for anonymous use an email address)",
			Sensitive:   true,
		},
		"path": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The path to upload log files to. If the path ends in `/` then it is treated as a directory",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred, in seconds (Default `3600`)",
		},
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     21,
			Description: "The port number. Default: `21`",
		},
		"public_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: TimestampFormatDescription,
		},
		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The username for the server (can be `anonymous`)",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
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
			Description: "The name of the condition to apply.",
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
func (h *FTPServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly FTP logging addition opts: %#v", opts)

	return createFTP(conn, opts)
}

// Read refreshes the resource.
func (h *FTPServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing FTP logging endpoints for (%s)", d.Id())
		ftpList, err := conn.ListFTPs(&gofastly.ListFTPsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up FTP logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenFTP(ftpList)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting FTP logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *FTPServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateFTPInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["address"]; ok {
		opts.Address = gofastly.String(v.(string))
	}
	if v, ok := modified["port"]; ok {
		opts.Port = gofastly.Int(v.(int))
	}
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.String(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.Username = gofastly.String(v.(string))
	}
	if v, ok := modified["password"]; ok {
		opts.Password = gofastly.String(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.String(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.Int(v.(int))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.Int(v.(int))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.String(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.String(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update FTP Opts: %#v", opts)
	_, err := conn.UpdateFTP(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *FTPServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly FTP logging endpoint removal opts: %#v", opts)

	return deleteFTP(conn, opts)
}

func createFTP(conn *gofastly.Client, i *gofastly.CreateFTPInput) error {
	_, err := conn.CreateFTP(i)
	return err
}

func deleteFTP(conn *gofastly.Client, i *gofastly.DeleteFTPInput) error {
	err := conn.DeleteFTP(i)
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

func flattenFTP(ftpList []*gofastly.FTP) []map[string]any {
	var fsl []map[string]any
	for _, fl := range ftpList {
		// Convert FTP logging to a map for saving to state.
		nfl := map[string]any{
			"name":               fl.Name,
			"address":            fl.Address,
			"user":               fl.Username,
			"password":           fl.Password,
			"path":               fl.Path,
			"port":               fl.Port,
			"period":             fl.Period,
			"public_key":         fl.PublicKey,
			"gzip_level":         fl.GzipLevel,
			"timestamp_format":   fl.TimestampFormat,
			"format":             fl.Format,
			"format_version":     fl.FormatVersion,
			"message_type":       fl.MessageType,
			"placement":          fl.Placement,
			"response_condition": fl.ResponseCondition,
			"compression_codec":  fl.CompressionCodec,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range nfl {
			if v == "" {
				delete(nfl, k)
			}
		}

		fsl = append(fsl, nfl)
	}

	return fsl
}

func (h *FTPServiceAttributeHandler) buildCreate(ftpMap any, serviceID string, serviceVersion int) *gofastly.CreateFTPInput {
	df := ftpMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(df)
	return &gofastly.CreateFTPInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              gofastly.String(df["name"].(string)),
		Address:           gofastly.String(df["address"].(string)),
		Username:          gofastly.String(df["user"].(string)),
		Password:          gofastly.String(df["password"].(string)),
		Path:              gofastly.String(df["path"].(string)),
		Port:              gofastly.Int(df["port"].(int)),
		Period:            gofastly.Int(df["period"].(int)),
		PublicKey:         gofastly.String(df["public_key"].(string)),
		GzipLevel:         gofastly.Int(df["gzip_level"].(int)),
		TimestampFormat:   gofastly.String(df["timestamp_format"].(string)),
		MessageType:       gofastly.String(df["message_type"].(string)),
		CompressionCodec:  gofastly.String(df["compression_codec"].(string)),
		Format:            gofastly.String(vla.format),
		FormatVersion:     gofastly.Int(intOrDefault(vla.formatVersion)),
		Placement:         gofastly.String(vla.placement),
		ResponseCondition: gofastly.String(vla.responseCondition),
	}
}

func (h *FTPServiceAttributeHandler) buildDelete(ftpMap any, serviceID string, serviceVersion int) *gofastly.DeleteFTPInput {
	df := ftpMap.(map[string]any)

	return &gofastly.DeleteFTPInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
