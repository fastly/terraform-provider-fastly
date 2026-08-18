package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyServiceCDNAuto_withHeader(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	headerName := fmt.Sprintf("header-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithHeader(serviceName, domainName, headerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.name", headerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.action", "delete"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.type", "cache"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.destination", "http.x-amz-request-id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.priority", "100"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.ignore_if_set", "false"),
					// Adding a header should create and activate version 2
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.#", "0"),
					// Removing the header should create and activate version 3
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withHeaderUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	headerName := fmt.Sprintf("header-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithHeader(serviceName, domainName, headerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.action", "delete"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.type", "cache"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				// Same name, in-place update of action/type/destination/source/priority/
				// ignore_if_set - confirm update, not delete+recreate.
				Config: ConfigCDNAutoWithHeaderUpdated(serviceName, domainName, headerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.name", headerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.action", "set"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.type", "request"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.destination", "http.X-Custom"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.source", "req.http.Host"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.priority", "10"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.ignore_if_set", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withHeaderRequestCondition confirms that a header's
// request_condition can reference a real nested REQUEST-type condition block within the same
// apply, verifying conditions are reconciled before header so the reference resolves within the
// same service version (see servicecdnauto's Create/Update, which reconcile header after ACL and
// after condition for this reason).
func TestAccFastlyServiceCDNAuto_withHeaderRequestCondition(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	headerName := fmt.Sprintf("header-%s", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithHeaderRequestCondition(serviceName, domainName, headerName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.name", headerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.request_condition", conditionName),
				),
			},
			{
				// Removing both the header and the condition it references in the same apply
				// must not fail - see the analogous backendWithRequestCondition test's comment.
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.#", "0"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_headerInvalidAction(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	headerName := fmt.Sprintf("header-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigCDNAutoWithHeaderInvalidAction(serviceName, domainName, headerName),
				ExpectError: regexp.MustCompile(`Attribute header\[0\]\.action value must be one of`),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_importWithHeader(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	headerName := fmt.Sprintf("header-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithHeader(serviceName, domainName, headerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "header.0.name", headerName),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
		},
	})
}
