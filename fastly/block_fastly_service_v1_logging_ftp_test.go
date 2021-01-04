package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenFTP(t *testing.T) {
	cases := []struct {
		remote []*gofastly.FTP
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.FTP{
				{
					ServiceVersion:  1,
					Name:            "ftp-endpoint",
					Address:         "ftp.example.com",
					Username:        "username",
					Password:        "password",
					PublicKey:       pgpPublicKey(t),
					Path:            "/path",
					Port:            21,
					Period:          3600,
					GzipLevel:       3,
					Format:          "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:   2,
					TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
					Placement:       "none",
					MessageType:     "classic",
				},
			},
			local: []map[string]interface{}{
				{
					"name":             "ftp-endpoint",
					"address":          "ftp.example.com",
					"user":             "username",
					"password":         "password",
					"public_key":       pgpPublicKey(t),
					"path":             "/path",
					"period":           uint(3600),
					"port":             uint(21),
					"gzip_level":       uint8(3),
					"format_version":   uint(2),
					"format":           "%h %l %u %t \"%r\" %>s %b",
					"timestamp_format": "%Y-%m-%dT%H:%M:%S.000",
					"placement":        "none",
					"message_type":     "classic",
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

func TestAccFastlyServiceV1_logging_ftp_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.FTP{
		ServiceVersion:  1,
		Name:            "ftp-endpoint",
		Address:         "ftp.example.com",
		Username:        "user",
		Password:        "p@ssw0rd",
		PublicKey:       pgpPublicKey(t),
		Path:            "/path",
		Port:            27,
		GzipLevel:       3,
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		Format:          "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:   2,
		Placement:       "none",
		MessageType:     "classic",
	}

	log1_after_update := gofastly.FTP{
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
		ServiceVersion:  1,
		Name:            "another-ftp-endpoint",
		Address:         "ftp.example.com",
		Username:        "user",
		Password:        "p@ssw0rd",
		Path:            "/",
		PublicKey:       pgpPublicKey(t),
		Port:            21,
		GzipLevel:       3,
		Period:          360,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		Format:          "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:   2,
		Placement:       "none",
		MessageType:     "classic",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1FTPConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1FTPAttributes(&service, []*gofastly.FTP{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_ftp.#", "1"),
				),
			},

			{
				Config: testAccServiceV1FTPConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1FTPAttributes(&service, []*gofastly.FTP{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_ftp.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_ftp_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.FTP{
		ServiceVersion:  1,
		Name:            "ftp-endpoint",
		Address:         "ftp.example.com",
		Username:        "user",
		Password:        "p@ssw0rd",
		PublicKey:       pgpPublicKey(t),
		Path:            "/path",
		Port:            27,
		GzipLevel:       3,
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1FTPComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1FTPAttributes(&service, []*gofastly.FTP{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_ftp.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1FTPAttributes(service *gofastly.ServiceDetail, ftps []*gofastly.FTP, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		ftpList, err := conn.ListFTPs(&gofastly.ListFTPsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up FTP Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(ftpList) != len(ftps) {
			return fmt.Errorf("FTP List count mismatch, expected (%d), got (%d)", len(ftps), len(ftpList))
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
						return fmt.Errorf("Bad match FTP logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(ftps) {
			return fmt.Errorf("Error matching FTP Logging rules")
		}

		return nil
	}
}

func testAccServiceV1FTPComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-ftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_ftp {
    name       = "ftp-endpoint"
    address    = "ftp.example.com"
    user       = "user"
    public_key = file("test_fixtures/fastly_test_publickey")
    password         = "p@ssw0rd"
    path             = "/path"
    port             = 27
    gzip_level       = 3
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    message_type     = "classic"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1FTPConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-ftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_ftp {
    name       = "ftp-endpoint"
    address    = "ftp.example.com"
    user       = "user"
    public_key = file("test_fixtures/fastly_test_publickey")
    password         = "p@ssw0rd"
    path             = "/path"
    port             = 27
    format           = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    gzip_level       = 3
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    placement        = "none"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1FTPConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-ftp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_ftp {
    name       = "ftp-endpoint"
    address    = "ftp2.example.com"
    user       = "user"
    password   = "p@ssw0rd2"
    public_key = file("test_fixtures/fastly_test_publickey")
    path             = "/path"
    format           = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    gzip_level       = 4
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    placement        = "waf_debug"
  }

  logging_ftp {
    name       = "another-ftp-endpoint"
    address    = "ftp.example.com"
    user       = "user"
    password   = "p@ssw0rd"
    public_key = file("test_fixtures/fastly_test_publickey")
    path             = "/"
    period           = 360
    format           = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    gzip_level       = 3
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    placement        = "none"
  }

  force_destroy = true
}
`, name, domain)
}
