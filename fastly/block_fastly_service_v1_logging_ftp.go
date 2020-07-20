package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type FTPServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingFTP(sa ServiceMetadata) ServiceAttributeDefinition {
	return &FTPServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_ftp",
			serviceMetadata: sa,
		},
	}
}

func (h *FTPServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	of, nf := d.GetChange(h.GetKey())

	if of == nil {
		of = new(schema.Set)
	}
	if nf == nil {
		nf = new(schema.Set)
	}

	ofs := of.(*schema.Set)
	nfs := nf.(*schema.Set)

	removeFTPLogging := ofs.Difference(nfs).List()
	addFTPLogging := nfs.Difference(ofs).List()

	// DELETE old FTP logging endpoints.
	for _, oRaw := range removeFTPLogging {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly FTP logging endpoint removal opts: %#v", opts)

		if err := deleteFTP(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated FTP logging endpoints.
	for _, nRaw := range addFTPLogging {
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
		// during state Gets, specifically `GetChange("logging_ftp")` in this case.
		if v, ok := ef["name"]; !ok || v.(string) == "" {
			continue
		}

		opts := h.buildCreate(ef, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly FTP logging addition opts: %#v", opts)

		if err := createFTP(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *FTPServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh FTP.
	log.Printf("[DEBUG] Refreshing FTP logging endpoints for (%s)", d.Id())
	ftpList, err := conn.ListFTPs(&gofastly.ListFTPsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up FTP logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenFTP(ftpList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting FTP logging endpoints for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *FTPServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the FTP logging endpoint.",
		},

		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The FTP URL to stream logs to.",
		},

		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The username for the server (can be anonymous).",
		},

		"password": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The password for the server (for anonymous use an email address).",
			Sensitive:   true,
		},

		"path": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The path to upload log files to. If the path ends in / then it is treated as blockAttributes directory.",
		},

		// Optional fields
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     21,
			Description: "The port number.",
		},

		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred, in seconds (Default 3600).",
		},

		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The PGP public key that Fastly will use to encrypt your log files before writing them to disk.",
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "Gzip Compression level.",
		},

		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: "specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).",
		},

		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted (default: `classic`)",
			ValidateFunc: validateLoggingMessageType(),
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
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

func flattenFTP(ftpList []*gofastly.FTP) []map[string]interface{} {
	var fsl []map[string]interface{}
	for _, fl := range ftpList {
		// Convert FTP logging to a map for saving to state.
		nfl := map[string]interface{}{
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

func (h *FTPServiceAttributeHandler) buildCreate(ftpMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateFTPInput {
	df := ftpMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateFTPInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              df["name"].(string),
		Address:           df["address"].(string),
		Username:          df["user"].(string),
		Password:          df["password"].(string),
		Path:              df["path"].(string),
		Port:              uint(df["port"].(int)),
		Period:            uint(df["period"].(int)),
		PublicKey:         df["public_key"].(string),
		GzipLevel:         uint8(df["gzip_level"].(int)),
		TimestampFormat:   df["timestamp_format"].(string),
		MessageType:       df["message_type"].(string),
		Format:            vla.format,
		FormatVersion:     vla.formatVersion,
		Placement:         vla.placement,
		ResponseCondition: vla.responseCondition,
	}
}

func (h *FTPServiceAttributeHandler) buildDelete(ftpMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteFTPInput {
	df := ftpMap.(map[string]interface{})

	return &gofastly.DeleteFTPInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
