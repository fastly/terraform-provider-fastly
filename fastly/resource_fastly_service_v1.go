package fastly

import (
	"errors"
	"fmt"
	"log"
	"time"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var fastlyNoServiceFoundErr = errors.New("No matching Fastly Service found")

func resourceServiceV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceV1Create,
		Read:   resourceServiceV1Read,
		Update: resourceServiceV1Update,
		Delete: resourceServiceV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for this Service",
			},

			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Managed by Terraform",
				Description: "A personal freeform descriptive note",
			},

			"version_comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A personal freeform descriptive note",
			},

			// Active Version represents the currently activated version in Fastly. In
			// Terraform, we abstract this number away from the users and manage
			// creating and activating. It's used internally, but also exported for
			// users to see.
			"active_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			// Cloned Version represents the latest cloned version by the provider. It
			// gets set whenever Terraform detects changes and clones the currently
			// activated version in order to modify it. Active Version and Cloned
			// Version can be different if the Activate field is set to false in order
			// to prevent the service from being activated. It is not used internally,
			// but it is exported for users to see after running `terraform apply`.
			"cloned_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"activate": {
				Type:        schema.TypeBool,
				Description: "Conditionally prevents the Service from being activated",
				Default:     true,
				Optional:    true,
			},

			"default_ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3600,
				Description: "The default Time-to-live (TTL) for the version",
			},

			"default_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The default hostname for the version",
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"domain": 				domainSchema,
			"backend": 				backendSchema,
			"cache_setting": 		cachesettingSchema,
			"condition":			conditionSchema,
			"healthcheck":			healthcheckSchema,
			"director":				directorSchema,
			"gzip":               gzipSchema,
			"header":             headerSchema,
			"s3logging":          s3loggingSchema,
			"papertrail":         papertrailSchema,
			"sumologic":          sumologicSchema,
			"gcslogging":         gcsloggingSchema,
			"bigquerylogging":    bigqueryloggingSchema,
			"syslog":             syslogSchema,
			"logentries":         logentriesSchema,
			"splunk":             splunkSchema,
			"blobstoragelogging": blobstorageloggingSchema,

			"httpslogging":       httpsloggingSchema,
			"logging_elasticsearch": elasticsearchSchema,
			"logging_ftp":           ftpSchema,
			"logging_sftp":          sftpSchema,

			"response_object":    responseobjectSchema,
			"request_setting":    requestsettingSchema,

			"vcl":					vclSchema,

			"snippet": 		  snippetSchema,
			"dynamicsnippet": dynamicsnippetSchema,
			"acl":            aclSchema,
			"dictionary":     dictionarySchema,
		},
	}
}

func resourceServiceV1Create(d *schema.ResourceData, meta interface{}) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn
	service, err := conn.CreateService(&gofastly.CreateServiceInput{
		Name:    d.Get("name").(string),
		Comment: d.Get("comment").(string),
	})

	if err != nil {
		return err
	}

	d.SetId(service.ID)
	return resourceServiceV1Update(d, meta)
}

func resourceServiceV1Update(d *schema.ResourceData, meta interface{}) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn

	// Update Name and/or Comment. No new verions is required for this
	if d.HasChange("name") || d.HasChange("comment") {
		_, err := conn.UpdateService(&gofastly.UpdateServiceInput{
			ID:      d.Id(),
			Name:    d.Get("name").(string),
			Comment: d.Get("comment").(string),
		})
		if err != nil {
			return err
		}
	}

	// Once activated, Versions are locked and become immutable. This is true for
	// versions that are no longer active. For Domains, Backends, DefaultHost and
	// DefaultTTL, a new Version must be created first, and updates posted to that
	// Version. Loop these attributes and determine if we need to create a new version first
	var needsChange bool
	for _, v := range []string{
		"domain",
		"backend",
		"default_host",
		"default_ttl",
		"director",
		"header",
		"gzip",
		"healthcheck",
		"s3logging",
		"papertrail",
		"gcslogging",
		"bigquerylogging",
		"syslog",
		"sumologic",
		"logentries",
		"splunk",
		"blobstoragelogging",
		"httpslogging",
		"logging_elasticsearch",
		"logging_ftp",
		"logging_sftp",
		"response_object",
		"condition",
		"request_setting",
		"cache_setting",
		"snippet",
		"dynamicsnippet",
		"vcl",
		"acl",
		"dictionary",
	} {
		if d.HasChange(v) {
			needsChange = true
		}
	}

	// Update the active version's comment. No new version is required for this
	if d.HasChange("version_comment") && !needsChange {
		latestVersion := d.Get("active_version").(int)
		if latestVersion == 0 {
			// If the service was just created, there is an empty Version 1 available
			// that is unlocked and can be updated
			latestVersion = 1
		}

		opts := gofastly.UpdateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
			Comment: d.Get("version_comment").(string),
		}

		log.Printf("[DEBUG] Update Version opts: %#v", opts)
		_, err := conn.UpdateVersion(&opts)
		if err != nil {
			return err
		}
	}

	initialVersion := false

	if needsChange {
		latestVersion := d.Get("active_version").(int)
		if latestVersion == 0 {
			initialVersion = true
			// If the service was just created, there is an empty Version 1 available
			// that is unlocked and can be updated
			latestVersion = 1
		} else {
			// Clone the latest version, giving us an unlocked version we can modify
			log.Printf("[DEBUG] Creating clone of version (%d) for updates", latestVersion)
			newVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
				Service: d.Id(),
				Version: latestVersion,
			})
			if err != nil {
				return err
			}

			// The new version number is named "Number", but it's actually a string
			latestVersion = newVersion.Number
			d.Set("cloned_version", latestVersion)

			// New versions are not immediately found in the API, or are not
			// immediately mutable, so we need to sleep a few and let Fastly ready
			// itself. Typically, 7 seconds is enough
			log.Print("[DEBUG] Sleeping 7 seconds to allow Fastly Version to be available")
			time.Sleep(7 * time.Second)

			// Update the cloned version's comment
			if d.Get("version_comment").(string) != "" {
				opts := gofastly.UpdateVersionInput{
					Service: d.Id(),
					Version: latestVersion,
					Comment: d.Get("version_comment").(string),
				}

				log.Printf("[DEBUG] Update Version opts: %#v", opts)
				_, err := conn.UpdateVersion(&opts)
				if err != nil {
					return err
				}
			}
		}

		// update general settings

		// If the requested default_ttl is 0, and this is the first
		// version being created, HasChange will return false, but we need
		// to set it anyway, so ensure we update the settings in that
		// case.
		if d.HasChange("default_host") || d.HasChange("default_ttl") || (d.Get("default_ttl") == 0 && initialVersion) {
			opts := gofastly.UpdateSettingsInput{
				Service: d.Id(),
				Version: latestVersion,
				// default_ttl has the same default value of 3600 that is provided by
				// the Fastly API, so it's safe to include here
				DefaultTTL: uint(d.Get("default_ttl").(int)),
			}

			if attr, ok := d.GetOk("default_host"); ok {
				opts.DefaultHost = attr.(string)
			}

			log.Printf("[DEBUG] Update Settings opts: %#v", opts)
			_, err := conn.UpdateSettings(&opts)
			if err != nil {
				return err
			}
		}

		// Conditions need to be updated first, as they can be referenced by other
		// configuration objects (Backends, Request Headers, etc)

		// Find difference in Conditions
		if d.HasChange("condition") {
			if err := processCondition(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in domains
		if d.HasChange("domain") {
			if err := processDomain(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Healthchecks need to be updated BEFORE backends
		if d.HasChange("healthcheck") {
			if err := processHealthCheck(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in backends
		if d.HasChange("backend") {
			if err := processBackend(d, conn, latestVersion); err != nil {
				return err
			}
		}

		if d.HasChange("director") {
			if err := processDirector(d, conn, latestVersion); err != nil {
				return err
			}
		}

		if d.HasChange("header") {
			if err := processHeader(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in Gzips
		if d.HasChange("gzip") {
			if err := processGZIP(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in s3logging
		if d.HasChange("s3logging") {
			if err := processS3Logging(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Papertrail
		if d.HasChange("papertrail") {
			if err := processPapertrail(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Sumologic
		if d.HasChange("sumologic") {
			if err := processSumologic(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in gcslogging
		if d.HasChange("gcslogging") {
			if err := processGCSLogging(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in bigquerylogging
		if d.HasChange("bigquerylogging") {
			if err := processBigQueryLogging(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Syslog
		if d.HasChange("syslog") {
			if err := processSyslog(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Logentries
		if d.HasChange("logentries") {
			if err := processLogEntries(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Splunk logging configurations
		if d.HasChange("splunk") {
			if err := processSplunk(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Blob Storage logging configurations
		if d.HasChange("blobstoragelogging") {
			if err := processBlobStorageLogging(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find differences in HTTPS logging configuration
		if d.HasChange("httpslogging") {
			if err := processHTTPS(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find differences in Elasticsearch logging configuration
		if d.HasChange("logging_elasticsearch") {
			if err := processElasticsearch(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find differences in FTP logging configuration
		if d.HasChange("logging_ftp") {
			if err := processFTP(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find differences in SFTP logging configurations
		if d.HasChange("logging_sftp") {
			if err := processSFTP(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in Response Object
		if d.HasChange("response_object") {
			if err := processResponseObject(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// find difference in request settings
		if d.HasChange("request_setting") {
			if err := processRequestSetting(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in VCLs
		if d.HasChange("vcl") {
			if err := processVCL(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in VCL snippets
		if d.HasChange("snippet") {
			if err := processSnippet(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in VCL dynamic snippets
		if d.HasChange("dynamicsnippet") {
			if err := processDynamicSnippet(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in Cache Settings
		if d.HasChange("cache_setting") {
			if err := processCacheSetting(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in ACLs
		if d.HasChange("acl") {
			if err := processACL(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// Find differences in dictionary
		if d.HasChange("dictionary") {
			if err := processDictionary(d, conn, latestVersion); err != nil {
				return err
			}
		}

		// validate version
		log.Printf("[DEBUG] Validating Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
		valid, msg, err := conn.ValidateVersion(&gofastly.ValidateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error checking validation: %s", err)
		}

		if !valid {
			return fmt.Errorf("[ERR] Invalid configuration for Fastly Service (%s): %s", d.Id(), msg)
		}

		shouldActivate := d.Get("activate").(bool)
		if shouldActivate {
			log.Printf("[DEBUG] Activating Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
			_, err = conn.ActivateVersion(&gofastly.ActivateVersionInput{
				Service: d.Id(),
				Version: latestVersion,
			})
			if err != nil {
				return fmt.Errorf("[ERR] Error activating version (%d): %s", latestVersion, err)
			}

			// Only if the version is valid and activated do we set the active_version.
			// This prevents us from getting stuck in cloning an invalid version
			d.Set("active_version", latestVersion)
		} else {
			log.Printf("[INFO] Skipping activation of Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
			log.Print("[INFO] The Terraform definition is explicitly specified to not activate the changes on Fastly")
			log.Printf("[INFO] Version (%v) has been pushed and validated", latestVersion)
			log.Printf("[INFO] Visit https://manage.fastly.com/configure/services/%s/versions/%v and activate it manually", d.Id(), latestVersion)
		}
	}

	return resourceServiceV1Read(d, meta)
}

func resourceServiceV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Find the Service. Discard the service because we need the ServiceDetails,
	// not just a Service record
	_, err := findService(d.Id(), meta)
	if err != nil {
		switch err {
		case fastlyNoServiceFoundErr:
			log.Printf("[WARN] %s for ID (%s)", err, d.Id())
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	d.Set("name", s.Name)
	d.Set("comment", s.Comment)
	d.Set("version_comment", s.Version.Comment)
	d.Set("active_version", s.ActiveVersion.Number)

	// If CreateService succeeds, but initial updates to the Service fail, we'll
	// have an empty ActiveService version (no version is active, so we can't
	// query for information on it)
	if s.ActiveVersion.Number != 0 {
		settingsOpts := gofastly.GetSettingsInput{
			Service: d.Id(),
			Version: s.ActiveVersion.Number,
		}
		if settings, err := conn.GetSettings(&settingsOpts); err == nil {
			d.Set("default_host", settings.DefaultHost)
			d.Set("default_ttl", settings.DefaultTTL)
		} else {
			return fmt.Errorf("[ERR] Error looking up Version settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
		}

		// Refresh Domains
		if err := readDomain(conn, d, s); err != nil {
			return err
		}

		// Refresh Backends
		if err := readBackend(conn, d, s); err != nil {
			return err
		}

		// refresh directors
		if err := readDirector(conn, d, s); err != nil {
			return err
		}

		// refresh headers
		if err := readHeader(conn, d, s); err != nil {
			return err
		}

		// refresh gzips
		if err := readGZIP(conn, d, s); err != nil {
			return err
		}

		// refresh Healthcheck
		if err := readHealthCheck(conn, d, s); err != nil {
			return err
		}

		// refresh S3 Logging
		if err := readS3Logging(conn, d, s); err != nil {
			return err
		}

		// refresh Papertrail Logging
		if err := readPapertrail(conn, d, s); err != nil {
			return err
		}

		// refresh Sumologic Logging
		if err := readSumologic(conn, d, s); err != nil {
			return err
		}

		// refresh GCS Logging
		if err := readGCSLogging(conn, d, s); err != nil {
			return err
		}

		// refresh BigQuery Logging
		if err := readBigQueryLogging(conn, d, s); err != nil {
			return err
		}

		// refresh Syslog Logging
		if err := readSyslog(conn, d, s); err != nil {
			return err
		}

		// refresh Logentries Logging
		if err := readLogEntries(conn, d, s); err != nil {
			return err
		}

		// refresh Splunk Logging
		if err := readSplunk(conn, d, s); err != nil {
			return err
		}

		// refresh Blob Storage Logging
		if err := readBlobStorageLogging(conn, d, s); err != nil {
			return err
		}

		// Refresh HTTPS
		if err := readHTTPS(conn, d, s); err != nil {
			return err
		}

		// refresh Response Objects
		if err := readResponseObject(conn, d, s); err != nil {
			return err
		}

		// refresh Conditions
		if err := readCondition(conn, d, s); err != nil {
			return err
		}

		// refresh Request Settings
		if err := readRequestSetting(conn, d, s); err != nil {
			return err
		}

		// refresh VCLs
		if err := readVCL(conn, d, s); err != nil {
			return err
		}

		// refresh ACLs
		if err := readACL(conn, d, s); err != nil {
			return err
		}

		// refresh VCL Snippets
		if err := readSnippet(conn, d, s); err != nil {
			return err
		}

		// refresh Dynamic Snippets
		if err := readDynamicSnippet(conn, d, s); err != nil {
			return err
		}

		// refresh Cache Settings
		if err := readCacheSetting(conn, d, s); err != nil {
			return err
		}

		// refresh Dictionaries
		if err := readDictionary(conn, d, s); err != nil {
			return err
		}

	} else {
		log.Printf("[DEBUG] Active Version for Service (%s) is empty, no state to refresh", d.Id())
	}

	return nil
}

func resourceServiceV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Fastly will fail to delete any service with an Active Version.
	// If `force_destroy` is given, we deactivate the active version and then send
	// the DELETE call
	if d.Get("force_destroy").(bool) {
		s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ID: d.Id(),
		})

		if err != nil {
			return err
		}

		if s.ActiveVersion.Number != 0 {
			_, err := conn.DeactivateVersion(&gofastly.DeactivateVersionInput{
				Service: d.Id(),
				Version: s.ActiveVersion.Number,
			})
			if err != nil {
				return err
			}
		}
	}

	err := conn.DeleteService(&gofastly.DeleteServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	_, err = findService(d.Id(), meta)
	if err != nil {
		switch err {
		// we expect no records to be found here
		case fastlyNoServiceFoundErr:
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	// findService above returned something and nil error, but shouldn't have
	return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", d.Id())

}

// findService finds a Fastly Service via the ListServices endpoint, returning
// the Service if found.
//
// Fastly API does not include any "deleted_at" type parameter to indicate
// that a Service has been deleted. GET requests to a deleted Service will
// return 200 OK and have the full output of the Service for an unknown time
// (days, in my testing). In order to determine if a Service is deleted, we
// need to hit /service and loop the returned Services, searching for the one
// in question. This endpoint only returns active or "alive" services. If the
// Service is not included, then it's "gone"
//
// Returns a fastlyNoServiceFoundErr error if the Service is not found in the
// ListServices response.
func findService(id string, meta interface{}) (*gofastly.Service, error) {
	conn := meta.(*FastlyClient).conn

	l, err := conn.ListServices(&gofastly.ListServicesInput{})
	if err != nil {
		return nil, fmt.Errorf("[WARN] Error listing services (%s): %s", id, err)
	}

	for _, s := range l {
		if s.ID == id {
			log.Printf("[DEBUG] Found Service (%s)", id)
			return s, nil
		}
	}

	return nil, fastlyNoServiceFoundErr
}

func flattenDomains(list []*gofastly.Domain) []map[string]interface{} {
	dl := make([]map[string]interface{}, 0, len(list))

	for _, d := range list {
		dl = append(dl, map[string]interface{}{
			"name":    d.Name,
			"comment": d.Comment,
		})
	}

	return dl
}






