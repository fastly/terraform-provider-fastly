package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyServiceCDNAuto_withDictionary(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithDictionary(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "dictionary.0.dictionary_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.write_only", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withMultipleDictionaries(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	dictionaryName1 := fmt.Sprintf("dict_1_%s", acctest.RandString(10))
	dictionaryName2 := fmt.Sprintf("dict_2_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleDictionaries(serviceName, domainName, dictionaryName1, dictionaryName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName1),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "dictionary.0.dictionary_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.write_only", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.1.name", dictionaryName2),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "dictionary.1.dictionary_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.1.write_only", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.1.force_destroy", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withDictionaryUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithDictionary(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.write_only", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				// Toggling write_only is implemented as a delete-then-create (see
				// dictionary.ops.Update): the Fastly API rejects an in-place update of
				// write_only, so this dictionary is empty here specifically to stay under
				// the force_destroy guard - dictionary_id is NOT preserved across this step.
				Config: ConfigCDNAutoWithDictionaryWriteOnly(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.write_only", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withWriteOnlyDictionaryForceDestroy verifies that a non-empty
// write_only dictionary cannot have write_only toggled off (implemented as delete-then-create,
// see dictionary.ops.Update) without force_destroy, even though its items can't be inspected
// via the API to run the usual emptiness check.
func TestAccFastlyServiceCDNAuto_withWriteOnlyDictionaryForceDestroy(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithDictionaryWriteOnly(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.write_only", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "false"),
				),
			},
			{
				Config: ConfigCDNAutoWithDictionaryWriteOnly(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					AddDictionaryItem("fastly_service_cdn_auto.test", "dictionary.0"),
				),
			},
			{
				Config:      ConfigCDNAutoWithDictionary(serviceName, domainName, dictionaryName),
				ExpectError: regexp.MustCompile("cannot delete or change write_only"),
			},
			{
				// force_destroy is persisted here without touching write_only, so this step
				// doesn't trigger the guard - it just sets prev.ForceDestroy=true in state
				// so the following step's write_only toggle is permitted.
				Config: ConfigCDNAutoWithDictionaryWriteOnlyForceDestroy(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.write_only", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "true"),
				),
			},
			{
				Config: ConfigCDNAutoWithDictionaryForceDestroy(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.write_only", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "true"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withDictionaryForceDestroy(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))
	dictionaryNameUpdated := fmt.Sprintf("dict_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithDictionary(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "false"),
				),
			},
			{
				Config: ConfigCDNAutoWithDictionary(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					AddDictionaryItem("fastly_service_cdn_auto.test", "dictionary.0"),
				),
			},
			{
				Config:      ConfigCDNAutoWithDictionary(serviceName, domainName, dictionaryNameUpdated),
				ExpectError: regexp.MustCompile("cannot delete dictionary"),
			},
			{
				Config: ConfigCDNAutoWithDictionaryForceDestroy(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "true"),
				),
			},
			{
				Config: ConfigCDNAutoWithDictionaryForceDestroy(serviceName, domainName, dictionaryNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.force_destroy", "true"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withDictionary(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigComputeAutoWithDictionary(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "dictionary.0.dictionary_id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withDictionaryForceDestroy(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))
	dictionaryNameUpdated := fmt.Sprintf("dict_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithDictionary(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.0.force_destroy", "false"),
				),
			},
			{
				Config: ConfigComputeAutoWithDictionary(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					AddDictionaryItem("fastly_service_compute_auto.test", "dictionary.0"),
				),
			},
			{
				Config:      ConfigComputeAutoWithDictionary(serviceName, domainName, dictionaryNameUpdated),
				ExpectError: regexp.MustCompile("cannot delete dictionary"),
			},
			{
				Config: ConfigComputeAutoWithDictionaryForceDestroy(serviceName, domainName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "dictionary.0.force_destroy", "true"),
				),
			},
		},
	})
}
