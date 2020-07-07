package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type SFTPServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingSFTP(sa ServiceMetadata) ServiceAttributeDefinition {
	return &SFTPServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_sftp",
			serviceMetadata: sa,
		},
	}
}

func (h *SFTPServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the SFTP logging endpoint.",
		},

		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The SFTP address to stream logs to.",
		},

		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The username for the server.",
		},

		"path": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The path to upload log files to. If the path ends in / then it is treated as blockAttributes directory.",
		},

		"ssh_known_hosts": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A list of host keys for all hosts we can connect to over SFTP.",
		},

		// Optional fields
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     22,
			Description: "The port the SFTP service listens on. (Default: 22).",
		},

		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The password for the server. If both password and secret_key are passed, secret_key will be preferred.",
			Sensitive:   true,
		},

		"secret_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The SSH private key for the server. If both password and secret_key are passed, secret_key will be preferred.",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A PGP public key that Fastly will use to encrypt your log files before writing them to disk.",
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently log files are finalized so they can be available for reading (in seconds, default 3600).",
		},

		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "What level of GZIP encoding to have when dumping logs (default 0, no compression).",
		},

		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: "The strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).",
		},

		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted. One of: classic (default), loggly, logplex or blank.",
			ValidateFunc: validateLoggingMessageType(),
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
			Description: "The name of the condition to apply.",
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

func (h *SFTPServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	os, ns := d.GetChange(h.GetKey())

	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)

	removeSFTPLogging := oss.Difference(nss).List()
	addSFTPLogging := nss.Difference(oss).List()

	// DELETE old SFTP logging endpoints.
	for _, oRaw := range removeSFTPLogging {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly SFTP logging endpoint removal opts: %#v", opts)

		if err := deleteSFTP(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated SFTP logging endpoints.
	for _, nRaw := range addSFTPLogging {
		sf := nRaw.(map[string]interface{})

		// @HACK for a TF SDK Issue.
		//
		// This ensures that the required, `name`, field is present.
		//
		// If we have made it this far and `name` is not present, it is most-likely due
		// to a defunct diff as noted here - https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697.
		//
		// This is caused by using a StateFunc in a nested TypeSet. While the StateFunc
		// properly handles setting state with the StateFunc, it returns extra entries
		// during state Gets, specifically `GetChange("logging_sftp")` in this case.
		if v, ok := sf["name"]; !ok || v.(string) == "" {
			continue
		}

		opts := h.buildCreate(sf, serviceID, latestVersion)

		if opts.Password == nil && opts.SecretKey == nil {
			return fmt.Errorf("[ERR] Either password or secret_key must be set")
		}

		log.Printf("[DEBUG] Fastly SFTP logging addition opts: %#v", opts)

		if err := createSFTP(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *SFTPServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh SFTP.
	log.Printf("[DEBUG] Refreshing SFTP logging endpoints for (%s)", d.Id())
	sftpList, err := conn.ListSFTPs(&gofastly.ListSFTPsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up SFTP logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenSFTP(sftpList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting SFTP logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
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

func flattenSFTP(sftpList []*gofastly.SFTP) []map[string]interface{} {
	var ssl []map[string]interface{}
	for _, sl := range sftpList {
		// Convert SFTP logging to a map for saving to state.
		nsl := map[string]interface{}{
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

func (h *SFTPServiceAttributeHandler) buildCreate(sftpMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateSFTPInput {
	df := sftpMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateSFTPInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Address:           gofastly.NullString(df["address"].(string)),
		Name:              gofastly.NullString(df["name"].(string)),
		User:              gofastly.NullString(df["user"].(string)),
		Path:              gofastly.NullString(df["path"].(string)),
		PublicKey:         gofastly.NullString(df["public_key"].(string)),
		SecretKey:         gofastly.NullString(df["secret_key"].(string)),
		SSHKnownHosts:     gofastly.NullString(df["ssh_known_hosts"].(string)),
		Port:              gofastly.Uint(uint(df["port"].(int))),
		Password:          gofastly.NullString(df["password"].(string)),
		GzipLevel:         gofastly.Uint(uint(df["gzip_level"].(int))),
		TimestampFormat:   gofastly.NullString(df["timestamp_format"].(string)),
		MessageType:       gofastly.NullString(df["message_type"].(string)),
		Format:            gofastly.NullString(vla.format),
		FormatVersion:     gofastly.Uint(vla.formatVersion),
		Placement:         gofastly.NullString(vla.placement),
		ResponseCondition: gofastly.NullString(vla.responseCondition),
	}
}

func (h *SFTPServiceAttributeHandler) buildDelete(sftpMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteSFTPInput {
	df := sftpMap.(map[string]interface{})

	return &gofastly.DeleteSFTPInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
