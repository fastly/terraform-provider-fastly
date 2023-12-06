package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
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
					ServiceVersion:   gofastly.ToPointer(1),
					Name:             gofastly.ToPointer("ftp-endpoint"),
					Address:          gofastly.ToPointer("ftp.example.com"),
					Username:         gofastly.ToPointer("username"),
					Password:         gofastly.ToPointer("password"),
					PublicKey:        gofastly.ToPointer(pgpPublicKey(t)),
					Path:             gofastly.ToPointer("/path"),
					Port:             gofastly.ToPointer(21),
					Period:           gofastly.ToPointer(3600),
					GzipLevel:        gofastly.ToPointer(0),
					Format:           gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
					FormatVersion:    gofastly.ToPointer(2),
					TimestampFormat:  gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
					Placement:        gofastly.ToPointer("none"),
					MessageType:      gofastly.ToPointer("classic"),
					CompressionCodec: gofastly.ToPointer("zstd"),
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
					"period":            3600,
					"port":              21,
					"gzip_level":        0,
					"format_version":    2,
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
		out := flattenFTP(c.remote, nil)
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
		Address:           gofastly.ToPointer("ftp.example.com"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("ftp-endpoint"),
		Password:          gofastly.ToPointer("p@ssw0rd"),
		Path:              gofastly.ToPointer("/path"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		Port:              gofastly.ToPointer(27),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Username:          gofastly.ToPointer("user"),
	}

	log1AfterUpdate := gofastly.FTP{
		Address:           gofastly.ToPointer("ftp2.example.com"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(4),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("ftp-endpoint"),
		Password:          gofastly.ToPointer("p@ssw0rd2"),
		Path:              gofastly.ToPointer("/path"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("waf_debug"),
		Port:              gofastly.ToPointer(21),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Username:          gofastly.ToPointer("user"),
	}

	log2 := gofastly.FTP{
		Address:           gofastly.ToPointer("ftp.example.com"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("another-ftp-endpoint"),
		Password:          gofastly.ToPointer("p@ssw0rd"),
		Path:              gofastly.ToPointer("/"),
		Period:            gofastly.ToPointer(360),
		Placement:         gofastly.ToPointer("none"),
		Port:              gofastly.ToPointer(21),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Username:          gofastly.ToPointer("user"),
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
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLFTPAttributes(&service, []*gofastly.FTP{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_ftp.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLFTPConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLFTPAttributes(&service, []*gofastly.FTP{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_ftp.#", "2"),
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
		Address:          gofastly.ToPointer("ftp.example.com"),
		CompressionCodec: gofastly.ToPointer("zstd"),
		GzipLevel:        gofastly.ToPointer(0),
		Name:             gofastly.ToPointer("ftp-endpoint"),
		Password:         gofastly.ToPointer("p@ssw0rd"),
		Path:             gofastly.ToPointer("/path"),
		Period:           gofastly.ToPointer(3600),
		Port:             gofastly.ToPointer(27),
		PublicKey:        gofastly.ToPointer(pgpPublicKey(t)),
		ServiceVersion:   gofastly.ToPointer(1),
		TimestampFormat:  gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Username:         gofastly.ToPointer("user"),
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
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLFTPAttributes(&service, []*gofastly.FTP{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_ftp.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLFTPAttributes(service *gofastly.ServiceDetail, ftps []*gofastly.FTP, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		ftpList, err := conn.ListFTPs(&gofastly.ListFTPsInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up FTP Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(ftpList) != len(ftps) {
			return fmt.Errorf("ftp List count mismatch, expected (%d), got (%d)", len(ftps), len(ftpList))
		}

		log.Printf("[DEBUG] ftpList = %#v\n", ftpList)

		var found int
		for _, e := range ftps {
			for _, el := range ftpList {
				if gofastly.ToValue(e.Name) == gofastly.ToValue(el.Name) {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
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
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

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
    source_code_hash = data.fastly_package_hash.example.hash
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
