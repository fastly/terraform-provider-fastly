package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenFTP(t *testing.T) {
	cases := []struct {
		remote []*gofastly.FTP
		local  []map[string]any
	}{
		{
			remote: []*gofastly.FTP{
				{
					ServiceVersion:   1,
					Name:             "ftp-endpoint",
					Address:          "ftp.example.com",
					Username:         "username",
					Password:         "password",
					PublicKey:        pgpPublicKey(t),
					Path:             "/path",
					Port:             21,
					Period:           3600,
					GzipLevel:        0,
					Format:           "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:    2,
					TimestampFormat:  "%Y-%m-%dT%H:%M:%S.000",
					Placement:        "none",
					MessageType:      "classic",
					CompressionCodec: "zstd",
				},
			},
			local: []map[string]any{
				{
					"name":              "ftp-endpoint",
					"address":           "ftp.example.com",
					"user":              "username",
					"password":          "password",
					"public_key":        pgpPublicKey(t),
					"path":              "/path",
					"period":            uint(3600),
					"port":              uint(21),
					"gzip_level":        uint8(0),
					"format_version":    uint(2),
					"format":            "%h %l %u %t \"%r\" %>s %b",
					"timestamp_format":  "%Y-%m-%dT%H:%M:%S.000",
					"placement":         "none",
					"message_type":      "classic",
					"compression_codec": "zstd",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenFTP(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceVCL_logging_ftp_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.FTP{
		ServiceVersion:   1,
		Name:             "ftp-endpoint",
		Address:          "ftp.example.com",
		Username:         "user",
		Password:         "p@ssw0rd",
		PublicKey:        pgpPublicKey(t),
		Path:             "/path",
		Port:             27,
		Period:           3600,
		TimestampFormat:  "%Y-%m-%dT%H:%M:%S.000",
		Format:           "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:    2,
		Placement:        "none",
		MessageType:      "classic",
		CompressionCodec: "zstd",
	}

	log1AfterUpdate := gofastly.FTP{
		ServiceVersion:  1,
		Name:            "ftp-endpoint",
		Address:         "ftp2.example.com",
		Username:        "user",
		Password:        "p@ssw0rd2",
		PublicKey:       pgpPublicKey(t),
		Path:            "/path",
		GzipLevel:       4,
		Port:            21,
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		Format:          "%h %l %u %t \"%r\" %>s %b %T",
		FormatVersion:   2,
		Placement:       "waf_debug",
		MessageType:     "classic",
	}

	log2 := gofastly.FTP{
		ServiceVersion:   1,
		Name:             "another-ftp-endpoint",
		Address:          "ftp.example.com",
		Username:         "user",
		Password:         "p@ssw0rd",
		Path:             "/",
		PublicKey:        pgpPublicKey(t),
		Port:             21,
		Period:           360,
		TimestampFormat:  "%Y-%m-%dT%H:%M:%S.000",
		Format:           "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:    2,
		Placement:        "none",
		MessageType:      "classic",
		CompressionCodec: "zstd",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLFTPConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLFTPAttributes(&service, []*gofastly.FTP{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_ftp.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLFTPConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLFTPAttributes(&service, []*gofastly.FTP{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_ftp.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_ftp_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.FTP{
		ServiceVersion:   1,
		Name:             "ftp-endpoint",
		Address:          "ftp.example.com",
		Username:         "user",
		Password:         "p@ssw0rd",
		PublicKey:        pgpPublicKey(t),
		Path:             "/path",
		Port:             27,
		Period:           3600,
		TimestampFormat:  "%Y-%m-%dT%H:%M:%S.000",
		CompressionCodec: "zstd",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLFTPComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLFTPAttributes(&service, []*gofastly.FTP{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_ftp.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLFTPAttributes(service *gofastly.ServiceDetail, ftps []*gofastly.FTP, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		ftpList, err := conn.ListFTPs(&gofastly.ListFTPsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up FTP Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(ftpList) != len(ftps) {
			return fmt.Errorf("ftp List count mismatch, expected (%d), got (%d)", len(ftps), len(ftpList))
		}

		log.Printf("[DEBUG] ftpList = %#v\n", ftpList)

		var found int
		for _, e := range ftps {
			for _, el := range ftpList {
				if e.Name == el.Name {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					el.CreatedAt = nil
					el.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						el.FormatVersion = e.FormatVersion
						el.Format = e.Format
						el.ResponseCondition = e.ResponseCondition
						el.Placement = e.Placement
						el.MessageType = e.MessageType
					}

					if diff := cmp.Diff(e, el); diff != "" {
						return fmt.Errorf("bad match FTP logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(ftps) {
			return fmt.Errorf("error matching FTP Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLFTPComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-ftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_ftp {
    name = "ftp-endpoint"
    address = "ftp.example.com"
    user = "user"
    public_key = file("test_fixtures/fastly_test_publickey")
    password = "p@ssw0rd"
    path = "/path"
    port = 27
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
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

func testAccServiceVCLFTPConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-ftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_ftp {
    name = "ftp-endpoint"
    address = "ftp.example.com"
    user = "user"
    public_key = file("test_fixtures/fastly_test_publickey")
    password = "p@ssw0rd"
    path = "/path"
    port = 27
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    placement = "none"
    compression_codec = "zstd"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLFTPConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-ftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_ftp {
    name = "ftp-endpoint"
    address = "ftp2.example.com"
    user = "user"
    password = "p@ssw0rd2"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/path"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    gzip_level = 4
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    placement = "waf_debug"
  }

  logging_ftp {
    name = "another-ftp-endpoint"
    address = "ftp.example.com"
    user = "user"
    password = "p@ssw0rd"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    period = 360
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    placement = "none"
    compression_codec = "zstd"
  }

  force_destroy = true
}`, name, domain)
}
