package fastly

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
)

func TestResourceFastlyFlattenGooglePubSub(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Pubsub
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Pubsub{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("googlepubsub-endpoint"),
					User:              gofastly.ToPointer("user"),
					SecretKey:         gofastly.ToPointer(privateKey(t)),
					ProjectID:         gofastly.ToPointer("project-id"),
					Topic:             gofastly.ToPointer("topic"),
					ResponseCondition: gofastly.ToPointer("response_condition"),
					Format:            gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
					FormatVersion:     gofastly.ToPointer(2),
					Placement:         gofastly.ToPointer("none"),
					ProcessingRegion:  gofastly.ToPointer("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "googlepubsub-endpoint",
					"user":               "user",
					"secret_key":         privateKey(t),
					"project_id":         "project-id",
					"topic":              "topic",
					"response_condition": "response_condition",
					"format":             `%a %l %u %t %m %U%q %H %>s %b %T`,
					"placement":          "none",
					"format_version":     2,
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenGooglePubSub(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestUserEmailSchemaDefaultFunc(t *testing.T) {
	computeAttributes := ServiceMetadata{ServiceTypeCompute}
	v := NewServiceLoggingGooglePubSub(computeAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_googlepubsub"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Defaults to "" if no environment variable is set
	result, err := loggingResourceSchema["user"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling user DefaultFunc", err)
	}
	if result != "" {
		t.Fatalf("Error matching:\nexpected: \"\"\ngot: %#v", result)
	}

	// Actually set env var and expect it to be used to determine user
	envVarKey := "FASTLY_GOOGLE_PUBSUB_EMAIL"
	mockValue := "example@my-project.iam.gserviceaccount.com"
	originalEnvValue := os.Getenv(envVarKey)
	defer func() {
		err := os.Setenv(envVarKey, originalEnvValue)
		if err != nil {
			t.Fatalf("failed to reset the environment: %s", err)
		}
	}()
	err = os.Setenv(envVarKey, mockValue)
	if err != nil {
		t.Fatalf("failed to mock the environment: %s", err)
	}

	result, err = loggingResourceSchema["user"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling user DefaultFunc", err)
	}
	if result != mockValue {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", mockValue, result)
	}
}

func TestSecretKeySchemaDefaultFunc(t *testing.T) {
	computeAttributes := ServiceMetadata{ServiceTypeCompute}
	v := NewServiceLoggingGooglePubSub(computeAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_googlepubsub"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect secret_key to be sensitive
	if !loggingResourceSchema["secret_key"].Sensitive {
		t.Fatalf("Expected secret_key to be marked as a Sensitive value")
	}

	// Defaults to "" if no environment variable is set
	result, err := loggingResourceSchema["secret_key"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling secret_key DefaultFunc", err)
	}
	if result != "" {
		t.Fatalf("Error matching:\nexpected: \"\"\ngot: %#v", result)
	}

	// Actually set env var and expect it to be used to determine secret_key
	envVarKey := "FASTLY_GOOGLE_PUBSUB_SECRET_KEY"
	mockValue := "-----BEGIN PRIVATE KEY-----\nabc123\n-----END PRIVATE KEY-----\n"
	originalEnvValue := os.Getenv(envVarKey)
	defer func() {
		err := os.Setenv(envVarKey, originalEnvValue)
		if err != nil {
			t.Fatalf("failed to reset the environment: %s", err)
		}
	}()
	err = os.Setenv(envVarKey, mockValue)
	if err != nil {
		t.Fatalf("failed to mock the environment: %s", err)
	}

	result, err = loggingResourceSchema["secret_key"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling secret_key DefaultFunc", err)
	}
	if result != mockValue {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", mockValue, result)
	}
}

func TestAccFastlyServiceVCL_googlepubsublogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Pubsub{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("googlepubsublogger"),
		User:              gofastly.ToPointer("user"),
		SecretKey:         gofastly.ToPointer(privateKey(t)),
		ProjectID:         gofastly.ToPointer("project-id"),
		Topic:             gofastly.ToPointer("topic"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		Format:            gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
		FormatVersion:     gofastly.ToPointer(2),
		Placement:         gofastly.ToPointer("none"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log1AfterUpdate := gofastly.Pubsub{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("googlepubsublogger"),
		User:              gofastly.ToPointer("newuser"),
		SecretKey:         gofastly.ToPointer(privateKey(t)),
		ProjectID:         gofastly.ToPointer("new-project-id"),
		Topic:             gofastly.ToPointer("newtopic"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		Format:            gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
		FormatVersion:     gofastly.ToPointer(2),
		Placement:         gofastly.ToPointer("none"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	log2 := gofastly.Pubsub{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("googlepubsublogger2"),
		User:              gofastly.ToPointer("user2"),
		SecretKey:         gofastly.ToPointer(privateKey(t)),
		ProjectID:         gofastly.ToPointer("project-id"),
		Topic:             gofastly.ToPointer("topicb"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		Format:            gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
		FormatVersion:     gofastly.ToPointer(2),
		Placement:         gofastly.ToPointer("none"),
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
				Config: testAccServiceVCLGooglePubSubConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGooglePubSubAttributes(&service, []*gofastly.Pubsub{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_googlepubsub.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLGooglePubSubConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGooglePubSubAttributes(&service, []*gofastly.Pubsub{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_googlepubsub.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_googlepubsublogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Pubsub{
		ServiceVersion:   gofastly.ToPointer(1),
		Name:             gofastly.ToPointer("googlepubsublogger"),
		User:             gofastly.ToPointer("user"),
		SecretKey:        gofastly.ToPointer(privateKey(t)),
		ProjectID:        gofastly.ToPointer("project-id"),
		Topic:            gofastly.ToPointer("topic"),
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
				Config: testAccServiceVCLGooglePubSubComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLGooglePubSubAttributes(&service, []*gofastly.Pubsub{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_googlepubsub.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLGooglePubSubAttributes(service *gofastly.ServiceDetail, googlepubsub []*gofastly.Pubsub, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		googlepubsubList, err := conn.ListPubsubs(&gofastly.ListPubsubsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Google Cloud Pub/Sub Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(googlepubsubList) != len(googlepubsub) {
			return fmt.Errorf("google Cloud Pub/Sub List count mismatch, expected (%d), got (%d)", len(googlepubsub), len(googlepubsubList))
		}

		log.Printf("[DEBUG] googlepubsubList = %#v\n", googlepubsubList)

		var found int
		for _, s := range googlepubsub {
			for _, sl := range googlepubsubList {
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
						return fmt.Errorf("bad match Google Cloud Pub/Sub logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(googlepubsub) {
			return fmt.Errorf("error matching Google Cloud Pub/Sub Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLGooglePubSubComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-googlepubsub-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	logging_googlepubsub {
		name               = "googlepubsublogger"
		user               = "user"
		secret_key         = file("test_fixtures/fastly_test_privatekey")
		project_id         = "project-id"
	  topic  						 = "topic"
    processing_region = "us"
	}

	package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLGooglePubSubConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-googlepubsub-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }

	logging_googlepubsub {
		name               = "googlepubsublogger"
		user               = "user"
		secret_key         = file("test_fixtures/fastly_test_privatekey")
		project_id         = "project-id"
	  topic  						 = "topic"
		response_condition = "response_condition_test"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version     = 2
		placement          = "none"
    processing_region = "us"
	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLGooglePubSubConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-testing-domain"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }

	logging_googlepubsub {
		name               = "googlepubsublogger"
		user               = "newuser"
		secret_key         = file("test_fixtures/fastly_test_privatekey")
		project_id         = "new-project-id"
	  topic  						 = "newtopic"
		response_condition = "response_condition_test"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version     = 2
		placement          = "none"
	}

	logging_googlepubsub {
		name               = "googlepubsublogger2"
		user               = "user2"
		secret_key         = file("test_fixtures/fastly_test_privatekey")
		project_id         = "project-id"
	  topic  						 = "topicb"
		response_condition = "response_condition_test"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version     = 2
		placement          = "none"
	}

	force_destroy = true
}`, name, domain)
}
