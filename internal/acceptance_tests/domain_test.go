package acceptancetests

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceDomain_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("fastly_service_domain.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_domain.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_domain.test", "id"),
					CheckDomainExistsInFastly("fastly_service_cdn.test", domainName, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceDomain_withComment(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainWithComment(serviceName, domainName, "Production domain"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("fastly_service_domain.test", "comment", "Production domain"),
				),
			},
		},
	})
}

func TestAccFastlyServiceDomain_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_domain.test", "name", domainName),
					resource.TestCheckNoResourceAttr("fastly_service_domain.test", "comment"),
				),
			},
			{
				Config: ConfigDomainWithComment(serviceName, domainName, "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("fastly_service_domain.test", "comment", "Updated comment"),
				),
			},
		},
	})
}

func TestAccFastlyServiceDomain_multipleDomains(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain1Name := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	domain2Name := fmt.Sprintf("www.%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainMultiple(serviceName, domain1Name, domain2Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_domain.primary", "name", domain1Name),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "name", domain2Name),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "comment", "Additional domain"),
				),
			},
		},
	})
}

func TestAccFastlyServiceDomain_vclService(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-vcl-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_domain.test", "name", domainName),
					CheckDomainExistsInFastly("fastly_service_cdn.test", domainName, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceDomain_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	additionalDomainName := fmt.Sprintf("www.%s.example.com", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainForImport(serviceName, domainName, additionalDomainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "name", additionalDomainName),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "comment", "Additional domain"),
					// Capture service_id and version for import step
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_domain.additional"]
						if !ok {
							return fmt.Errorf("domain resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				// Test legacy composite ID import: service_id/version/name
				ResourceName: "fastly_service_domain.additional",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, additionalDomainName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyServiceDomain_importWithSubdomain(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	// Domain with multiple levels to ensure proper parsing
	subdomainName := fmt.Sprintf("api.v2.staging.%s.example.com", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainForImport(serviceName, domainName, subdomainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "name", subdomainName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_domain.additional"]
						if !ok {
							return fmt.Errorf("domain resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				// Test legacy composite ID import with complex subdomain
				ResourceName: "fastly_service_domain.additional",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, subdomainName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// CheckDomainExistsInFastly verifies a domain exists in Fastly API
func CheckDomainExistsInFastly(serviceName, domainName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		domain, err := client.GetDomain(context.Background(), &fastly.GetDomainInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           domainName,
		})
		if err != nil {
			return fmt.Errorf("error fetching domain from Fastly: %w", err)
		}

		if domain == nil {
			return fmt.Errorf("domain %s not found in Fastly", domainName)
		}

		return nil
	}
}
