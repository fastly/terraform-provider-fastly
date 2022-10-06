package fastly

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenSFTP(t *testing.T) {
	cases := []struct {
		remote []*gofastly.SFTP
		local  []map[string]any
	}{
		{
			remote: []*gofastly.SFTP{
				{
					ServiceVersion:    1,
					Name:              "sftp-endpoint",
					Address:           "sftp.example.com",
					User:              "user",
					Path:              "/",
					PublicKey:         pgpPublicKey(t),
					SecretKey:         privateKey(t),
					SSHKnownHosts:     "sftp.example.com",
					Format:            "%h %l %u %t \"%r\" %>s %b",
					Password:          "password",
					MessageType:       "classic",
					FormatVersion:     2,
					Period:            3600,
					Port:              22,
					TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
					ResponseCondition: "response_condition",
					Placement:         "none",
					GzipLevel:         0,
					CompressionCodec:  "zstd",
				},
			},
			local: []map[string]any{
				{
					"name":               "sftp-endpoint",
					"address":            "sftp.example.com",
					"user":               "user",
					"path":               "/",
					"ssh_known_hosts":    "sftp.example.com",
					"public_key":         pgpPublicKey(t),
					"secret_key":         privateKey(t),
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"password":           "password",
					"message_type":       "classic",
					"gzip_level":         uint8(0),
					"format_version":     uint(2),
					"period":             uint(3600),
					"port":               uint(22),
					"response_condition": "response_condition",
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"placement":          "none",
					"compression_codec":  "zstd",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSFTP(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceVCLLoggingSftpBasic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.SFTP{
		ServiceVersion: 1,

		// Configured
		Name:              "sftp-endpoint",
		Address:           "sftp.example.com",
		User:              "username",
		Password:          "password",
		PublicKey:         pgpPublicKey(t),
		Path:              "/",
		SSHKnownHosts:     "sftp.example.com",
		Placement:         "none",
		ResponseCondition: "response_condition_test",
		CompressionCodec:  "zstd",

		// Defaults
		Port:            22,
		Period:          3600,
		MessageType:     "classic",
		GzipLevel:       0,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		FormatVersion:   2,
		Format:          "%h %l %u %t \"%r\" %>s %b",
	}

	log1AfterUpdate := gofastly.SFTP{
		ServiceVersion:    1,
		Name:              "sftp-endpoint",
		Address:           "sftp.example.com",
		User:              "user",
		PublicKey:         pgpPublicKey(t),
		SecretKey:         privateKey(t),
		Path:              "/logs/",
		SSHKnownHosts:     "sftp.example.com",
		MessageType:       "blank",
		Port:              2600,
		Format:            "%h %l %u %t \"%r\" %>s %b %T",
		Placement:         "waf_debug",
		ResponseCondition: "response_condition_test",
		GzipLevel:         3,

		// Defaults
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		FormatVersion:   2,
	}

	log2 := gofastly.SFTP{
		ServiceVersion:    1,
		Name:              "another-sftp-endpoint",
		Address:           "sftp2.example.com",
		User:              "user",
		Path:              "/dir/",
		PublicKey:         pgpPublicKey(t),
		SecretKey:         privateKey(t),
		SSHKnownHosts:     "sftp2.example.com",
		ResponseCondition: "response_condition_test",
		MessageType:       "loggly",
		Placement:         "none",
		CompressionCodec:  "zstd",

		// Defaults
		Port:            22,
		Period:          3600,
		GzipLevel:       0,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		FormatVersion:   2,
		Format:          "%h %l %u %t \"%r\" %>s %b",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSFTPConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSFTPAttributes(&service, []*gofastly.SFTP{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_sftp.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLSFTPConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSFTPAttributes(&service, []*gofastly.SFTP{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_sftp.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCLLoggingSftpBasicCompute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.SFTP{
		ServiceVersion: 1,

		// Configured
		Name:             "sftp-endpoint",
		Address:          "sftp.example.com",
		User:             "username",
		Password:         "password",
		PublicKey:        pgpPublicKey(t),
		Path:             "/",
		SSHKnownHosts:    "sftp.example.com",
		CompressionCodec: "zstd",

		// Defaults
		Port:            22,
		Period:          3600,
		MessageType:     "classic",
		GzipLevel:       0,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSFTPComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLSFTPAttributes(&service, []*gofastly.SFTP{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_sftp.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCLLoggingSftpPasswordSecretKey(t *testing.T) {
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceVCLSFTPConfigNoPasswordSecretKey(name, domain),
				ExpectError: regexp.MustCompile("either password or secret_key must be set"),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLSFTPAttributes(service *gofastly.ServiceDetail, sftps []*gofastly.SFTP, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		sftpList, err := conn.ListSFTPs(&gofastly.ListSFTPsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up SFTP Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(sftpList) != len(sftps) {
			return fmt.Errorf("sftp List count mismatch, expected (%d), got (%d)", len(sftps), len(sftpList))
		}

		log.Printf("[DEBUG] sftpList = %#v\n", sftpList)

		var found int
		for _, s := range sftps {
			for _, sl := range sftpList {
				if s.Name == sl.Name {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					sl.CreatedAt = nil
					sl.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						sl.FormatVersion = s.FormatVersion
						sl.Format = s.Format
						sl.ResponseCondition = s.ResponseCondition
						sl.Placement = s.Placement
					}

					if diff := cmp.Diff(s, sl); diff != "" {
						return fmt.Errorf("bad match SFTP logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(sftps) {
			return fmt.Errorf("error matching SFTP Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLSFTPComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-sftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_sftp {
    name = "sftp-endpoint"
    address = "sftp.example.com"
    user = "username"
    password  = "password"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    ssh_known_hosts = "sftp.example.com"
    message_type = "classic"
    compression_codec = "zstd"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSFTPConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-sftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  condition {
    name = "response_condition_test"
    type = "RESPONSE"
    priority = 8
    statement = "resp.status == 418"
  }

  logging_sftp {
    name = "sftp-endpoint"
    address = "sftp.example.com"
    user = "username"
    password = "password"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    ssh_known_hosts = "sftp.example.com"
    message_type = "classic"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    placement = "none"
    response_condition = "response_condition_test"
    compression_codec = "zstd"
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSFTPConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-sftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  condition {
    name = "response_condition_test"
    type = "RESPONSE"
    priority = 8
    statement = "resp.status == 418"
  }

  logging_sftp {
    name = "sftp-endpoint"
    address = "sftp.example.com"
    port = 2600
    user = "user"
    public_key = file("test_fixtures/fastly_test_publickey")
    secret_key = file("test_fixtures/fastly_test_privatekey")
    path = "/logs/"
    ssh_known_hosts = "sftp.example.com"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    message_type = "blank"
    response_condition = "response_condition_test"
    placement = "waf_debug"
    gzip_level = 3
  }

  logging_sftp {
    name = "another-sftp-endpoint"
    address = "sftp2.example.com"
    user = "user"
    public_key = file("test_fixtures/fastly_test_publickey")
    secret_key = file("test_fixtures/fastly_test_privatekey")
    path = "/dir/"
    ssh_known_hosts = "sftp2.example.com"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    message_type = "loggly"
    response_condition = "response_condition_test"
    placement = "none"
    compression_codec = "zstd"
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSFTPConfigNoPasswordSecretKey(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-sftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_sftp {
    name = "sftp-endpoint"
    address = "sftp.example.com"
    user = "username"
    path = "/"
    ssh_known_hosts = "sftp.example.com"
  }
  force_destroy = true

}`, name, domain)
}
