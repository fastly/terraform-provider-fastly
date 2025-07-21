package fastly

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

func TestResourceFastlyFlattenSFTP(t *testing.T) {
	cases := []struct {
		remote []*gofastly.SFTP
		local  []map[string]any
	}{
		{
			remote: []*gofastly.SFTP{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("sftp-endpoint"),
					Address:           gofastly.ToPointer("sftp.example.com"),
					User:              gofastly.ToPointer("user"),
					Path:              gofastly.ToPointer("/"),
					PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
					SecretKey:         gofastly.ToPointer(privateKey(t)),
					SSHKnownHosts:     gofastly.ToPointer("sftp.example.com"),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
					Password:          gofastly.ToPointer("password"),
					MessageType:       gofastly.ToPointer("classic"),
					FormatVersion:     gofastly.ToPointer(2),
					Period:            gofastly.ToPointer(3600),
					Port:              gofastly.ToPointer(22),
					TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
					ResponseCondition: gofastly.ToPointer("response_condition"),
					Placement:         gofastly.ToPointer("none"),
					GzipLevel:         gofastly.ToPointer(0),
					CompressionCodec:  gofastly.ToPointer("zstd"),
					ProcessingRegion:  gofastly.ToPointer("eu"),
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
					"gzip_level":         0,
					"format_version":     2,
					"period":             3600,
					"port":               22,
					"response_condition": "response_condition",
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"placement":          "none",
					"compression_codec":  "zstd",
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSFTP(c.remote, nil)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceVCL_logging_sftp_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.SFTP{
		Address:           gofastly.ToPointer("sftp.example.com"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("sftp-endpoint"),
		Password:          gofastly.ToPointer("password"),
		Path:              gofastly.ToPointer("/"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		Port:              gofastly.ToPointer(22),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		SSHKnownHosts:     gofastly.ToPointer("sftp.example.com"),
		SecretKey:         gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:              gofastly.ToPointer("username"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log1AfterUpdate := gofastly.SFTP{
		Address:           gofastly.ToPointer("sftp.example.com"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(3),
		MessageType:       gofastly.ToPointer("blank"),
		Name:              gofastly.ToPointer("sftp-endpoint"),
		Password:          gofastly.ToPointer(""),
		Path:              gofastly.ToPointer("/logs/"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		Port:              gofastly.ToPointer(2600),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		SSHKnownHosts:     gofastly.ToPointer("sftp.example.com"),
		SecretKey:         gofastly.ToPointer(privateKey(t)),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:              gofastly.ToPointer("user"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	log2 := gofastly.SFTP{
		Address:           gofastly.ToPointer("sftp2.example.com"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("loggly"),
		Name:              gofastly.ToPointer("another-sftp-endpoint"),
		Password:          gofastly.ToPointer(""),
		Path:              gofastly.ToPointer("/dir/"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		Port:              gofastly.ToPointer(22),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		SSHKnownHosts:     gofastly.ToPointer("sftp2.example.com"),
		SecretKey:         gofastly.ToPointer(privateKey(t)),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:              gofastly.ToPointer("user"),
		ProcessingRegion:  gofastly.ToPointer("none"),
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
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSFTPAttributes(&service, []*gofastly.SFTP{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_sftp.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLSFTPConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSFTPAttributes(&service, []*gofastly.SFTP{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_sftp.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_sftp_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.SFTP{
		Address:          gofastly.ToPointer("sftp.example.com"),
		CompressionCodec: gofastly.ToPointer("zstd"),
		GzipLevel:        gofastly.ToPointer(0),
		MessageType:      gofastly.ToPointer("classic"),
		Name:             gofastly.ToPointer("sftp-endpoint"),
		Password:         gofastly.ToPointer("password"),
		Path:             gofastly.ToPointer("/"),
		Period:           gofastly.ToPointer(3600),
		Port:             gofastly.ToPointer(22),
		PublicKey:        gofastly.ToPointer(pgpPublicKey(t)),
		SSHKnownHosts:    gofastly.ToPointer("sftp.example.com"),
		SecretKey:        gofastly.ToPointer(""),
		ServiceVersion:   gofastly.ToPointer(1),
		TimestampFormat:  gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:             gofastly.ToPointer("username"),
		ProcessingRegion: gofastly.ToPointer("us"),
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
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLSFTPAttributes(&service, []*gofastly.SFTP{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_sftp.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_sftp_password_secret_key(t *testing.T) {
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
		sftpList, err := conn.ListSFTPs(context.TODO(), &gofastly.ListSFTPsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up SFTP Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(sftpList) != len(sftps) {
			return fmt.Errorf("sftp List count mismatch, expected (%d), got (%d)", len(sftps), len(sftpList))
		}

		log.Printf("[DEBUG] sftpList = %#v\n", sftpList)

		var found int
		for _, s := range sftps {
			for _, sl := range sftpList {
				if gofastly.ToValue(s.Name) == gofastly.ToValue(sl.Name) {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ServiceID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
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
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

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
    processing_region = "us"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
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
    processing_region = "us"
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
    placement = "none"
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
