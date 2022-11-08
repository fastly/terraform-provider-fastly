package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SFTPServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type SFTPServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingSFTP returns a new resource.
func NewServiceLoggingSFTP(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&SFTPServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_sftp",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *SFTPServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *SFTPServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The SFTP address to stream logs to",
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
			Description: "The unique name of the SFTP logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The password for the server. If both `password` and `secret_key` are passed, `secret_key` will be preferred",
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
			Description: "How frequently log files are finalized so they can be available for reading (in seconds, default `3600`)",
		},
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     22,
			Description: "The port the SFTP service listens on. (Default: `22`)",
		},
		"public_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"secret_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The SSH private key for the server. If both `password` and `secret_key` are passed, `secret_key` will be preferred",
			Sensitive:        true,
			ValidateDiagFunc: validateStringTrimmed,
		},
		"ssh_known_hosts": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A list of host keys for all hosts we can connect to over SFTP",
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
			Description: "The username for the server",
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
func (h *SFTPServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	if *opts.Password == "" && *opts.SecretKey == "" {
		return fmt.Errorf("either password or secret_key must be set")
	}

	log.Printf("[DEBUG] Fastly SFTP logging addition opts: %#v", opts)

	return createSFTP(conn, opts)
}

// Read refreshes the resource.
func (h *SFTPServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing SFTP logging endpoints for (%s)", d.Id())
		sftpList, err := conn.ListSFTPs(&gofastly.ListSFTPsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up SFTP logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenSFTP(sftpList)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting SFTP logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *SFTPServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateSFTPInput{
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
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.String(v.(string))
	}
	if v, ok := modified["ssh_known_hosts"]; ok {
		opts.SSHKnownHosts = gofastly.String(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.String(v.(string))
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
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.String(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.Int(v.(int))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.String(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update SFTP Opts: %#v", opts)
	_, err := conn.UpdateSFTP(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *SFTPServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly SFTP logging endpoint removal opts: %#v", opts)

	return deleteSFTP(conn, opts)
}

func createSFTP(conn *gofastly.Client, i *gofastly.CreateSFTPInput) error {
	_, err := conn.CreateSFTP(i)
	return err
}

func deleteSFTP(conn *gofastly.Client, i *gofastly.DeleteSFTPInput) error {
	err := conn.DeleteSFTP(i)

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

func flattenSFTP(sftpList []*gofastly.SFTP) []map[string]any {
	var ssl []map[string]any
	for _, sl := range sftpList {
		// Convert SFTP logging to a map for saving to state.
		nsl := map[string]any{
			"name":               sl.Name,
			"address":            sl.Address,
			"user":               sl.User,
			"path":               sl.Path,
			"ssh_known_hosts":    sl.SSHKnownHosts,
			"port":               sl.Port,
			"password":           sl.Password,
			"secret_key":         sl.SecretKey,
			"public_key":         sl.PublicKey,
			"period":             sl.Period,
			"gzip_level":         sl.GzipLevel,
			"timestamp_format":   sl.TimestampFormat,
			"message_type":       sl.MessageType,
			"format":             sl.Format,
			"format_version":     sl.FormatVersion,
			"response_condition": sl.ResponseCondition,
			"placement":          sl.Placement,
			"compression_codec":  sl.CompressionCodec,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nsl {
			if v == "" {
				delete(nsl, k)
			}
		}

		ssl = append(ssl, nsl)
	}

	return ssl
}

func (h *SFTPServiceAttributeHandler) buildCreate(sftpMap any, serviceID string, serviceVersion int) *gofastly.CreateSFTPInput {
	df := sftpMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(df)
	return &gofastly.CreateSFTPInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Address:           gofastly.String(df["address"].(string)),
		Name:              gofastly.String(df["name"].(string)),
		User:              gofastly.String(df["user"].(string)),
		Path:              gofastly.String(df["path"].(string)),
		PublicKey:         gofastly.String(df["public_key"].(string)),
		SecretKey:         gofastly.String(df["secret_key"].(string)),
		SSHKnownHosts:     gofastly.String(df["ssh_known_hosts"].(string)),
		Port:              gofastly.Int(df["port"].(int)),
		Password:          gofastly.String(df["password"].(string)),
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

func (h *SFTPServiceAttributeHandler) buildDelete(sftpMap any, serviceID string, serviceVersion int) *gofastly.DeleteSFTPInput {
	df := sftpMap.(map[string]any)

	return &gofastly.DeleteSFTPInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
