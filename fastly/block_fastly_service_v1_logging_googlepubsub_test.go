package fastly

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v5/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenGooglePubSub(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Pubsub
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Pubsub{
				{
					ServiceVersion:    1,
					Name:              "googlepubsub-endpoint",
					User:              "user",
					SecretKey:         privateKey(t),
					ProjectID:         "project-id",
					Topic:             "topic",
					ResponseCondition: "response_condition",
					Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
					FormatVersion:     2,
					Placement:         "none",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "googlepubsub-endpoint",
					"user":               "user",
					"secret_key":         privateKey(t),
					"project_id":         "project-id",
					"topic":              "topic",
					"response_condition": "response_condition",
					"format":             `%a %l %u %t %m %U%q %H %>s %b %T`,
					"placement":          "none",
					"format_version":     uint(2),
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
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	v.Register(resource)
	loggingResource := resource.Schema["logging_googlepubsub"]
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
		os.Setenv(envVarKey, originalEnvValue)
	}()
	os.Setenv(envVarKey, mockValue)

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
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	v.Register(resource)
	loggingResource := resource.Schema["logging_googlepubsub"]
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
		os.Setenv(envVarKey, originalEnvValue)
	}()
	os.Setenv(envVarKey, mockValue)

	result, err = loggingResourceSchema["secret_key"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling secret_key DefaultFunc", err)
	}
	if result != mockValue {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", mockValue, result)
	}
}

func TestAccFastlyServiceV1_googlepubsublogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Pubsub{
		ServiceVersion:    1,
		Name:              "googlepubsublogger",
		User:              "user",
		SecretKey:         privateKey(t),
		ProjectID:         "project-id",
		Topic:             "topic",
		ResponseCondition: "response_condition_test",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		FormatVersion:     2,
		Placement:         "none",
	}

	log1_after_update := gofastly.Pubsub{
		ServiceVersion:    1,
		Name:              "googlepubsublogger",
		User:              "newuser",
		SecretKey:         privateKey(t),
		ProjectID:         "new-project-id",
		Topic:             "newtopic",
		ResponseCondition: "response_condition_test",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		FormatVersion:     2,
		Placement:         "waf_debug",
	}

	log2 := gofastly.Pubsub{
		ServiceVersion:    1,
		Name:              "googlepubsublogger2",
		User:              "user2",
		SecretKey:         privateKey(t),
		ProjectID:         "project-id",
		Topic:             "topicb",
		ResponseCondition: "response_condition_test",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		FormatVersion:     2,
		Placement:         "none",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1GooglePubSubConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1GooglePubSubAttributes(&service, []*gofastly.Pubsub{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_googlepubsub.#", "1"),
				),
			},

			{
				Config: testAccServiceV1GooglePubSubConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1GooglePubSubAttributes(&service, []*gofastly.Pubsub{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_googlepubsub.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_googlepubsublogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Pubsub{
		ServiceVersion: 1,
		Name:           "googlepubsublogger",
		User:           "user",
		SecretKey:      privateKey(t),
		ProjectID:      "project-id",
		Topic:          "topic",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1GooglePubSubComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1GooglePubSubAttributes(&service, []*gofastly.Pubsub{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_googlepubsub.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1GooglePubSubAttributes(service *gofastly.ServiceDetail, googlepubsub []*gofastly.Pubsub, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		googlepubsubList, err := conn.ListPubsubs(&gofastly.ListPubsubsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Google Cloud Pub/Sub Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(googlepubsubList) != len(googlepubsub) {
			return fmt.Errorf("Google Cloud Pub/Sub List count mismatch, expected (%d), got (%d)", len(googlepubsub), len(googlepubsubList))
		}

		log.Printf("[DEBUG] googlepubsubList = %#v\n", googlepubsubList)

		var found int
		for _, s := range googlepubsub {
			for _, sl := range googlepubsubList {
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
						return fmt.Errorf("Bad match Google Cloud Pub/Sub logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(googlepubsub) {
			return fmt.Errorf("Error matching Google Cloud Pub/Sub Logging rules")
		}

		return nil
	}
}

func testAccServiceV1GooglePubSubComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
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
	}

	package {
      	filename = "test_fixtures/package/valid.tar.gz"
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
   	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceV1GooglePubSubConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceV1GooglePubSubConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
		placement          = "waf_debug"
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
