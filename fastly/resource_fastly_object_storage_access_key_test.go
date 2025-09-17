package fastly

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/objectstorage/accesskeys"
)

func TestAccFastlyObjectStorageAccessKey_basic(t *testing.T) {
	resourceName := "fastly_object_storage_access_keys.storage_key1"
	var accessKey accesskeys.AccessKey
	description := fmt.Sprintf("this is a test key %s", acctest.RandString(10))
	permission := "read-write-objects"
	bucket1 := "bucket1"
	bucket2 := "bucket2"

	type Config struct {
		Description string
		Permission  string
		Buckets     []string
	}

	tmplText := `
resource "fastly_object_storage_access_keys" "storage_key1" {
    buckets = ["{{ StringsJoin .Buckets "\", \"" }}"]
    description = "{{ .Description }}"
    permission = "{{ .Permission }}"
}`
	tmpl, err := template.New("test").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(tmplText)
	if err != nil {
		t.Fatal(err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectStorageAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: renderTestConfigTemplate(t, tmpl, Config{
					Description: description,
					Permission:  permission,
					Buckets:     []string{bucket1, bucket2},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectStorageAccessKeyExists(resourceName, &accessKey),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "permission", permission),
					resource.TestCheckResourceAttr(resourceName, "buckets.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "buckets.*", bucket1),
					resource.TestCheckTypeSetElemAttr(resourceName, "buckets.*", bucket2),
					resource.TestCheckResourceAttrSet(resourceName, "secret_key"),
					resource.TestCheckResourceAttrSet(resourceName, "access_key_id"),
				),
			},
		},
	})
}

func testAccCheckObjectStorageAccessKeyExists(n string, accessKey *accesskeys.AccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ObjectStorageAccessKey ID is set")
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		opts := accesskeys.GetInput{
			AccessKeyID: gofastly.ToPointer(rs.Primary.ID),
		}
		ak, err := accesskeys.Get(context.TODO(), conn, &opts)
		if err != nil {
			return err
		}

		*accessKey = *ak

		return nil
	}
}

func testAccCheckObjectStorageAccessKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_object_storage_access_keys" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		akl, err := accesskeys.ListAccessKeys(context.TODO(), conn)
		if err != nil {
			return fmt.Errorf("error getting current accessKeys when deleting Fastly Object Storage AccessKey (%s): %s", rs.Primary.ID, err)
		}

		for _, ak := range akl.Data {
			if ak.AccessKeyID == rs.Primary.ID {
				// accessKey still found
				return fmt.Errorf("tried deleting ObjectStorageAccessKey (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}
