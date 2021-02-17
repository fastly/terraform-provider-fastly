package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceWAFVersionV1DetermineVersion(t *testing.T) {

	cases := []struct {
		remote  []*gofastly.WAFVersion
		local   int
		Errored bool
	}{
		{
			remote:  []*gofastly.WAFVersion{},
			local:   0,
			Errored: true,
		},
		{
			remote: []*gofastly.WAFVersion{
				{Number: 1},
			},
			local:   1,
			Errored: false,
		},
		{
			remote: []*gofastly.WAFVersion{
				{Number: 1},
				{Number: 2},
			},
			local:   2,
			Errored: false,
		},
		{
			remote: []*gofastly.WAFVersion{
				{Number: 3},
				{Number: 2},
				{Number: 1},
			},
			local:   3,
			Errored: false,
		},
	}

	for _, c := range cases {
		out, err := determineLatestVersion(c.remote)
		if (err == nil) == c.Errored {
			t.Fatalf("Error expected to be %v but wan't", c.Errored)
		}
		if out == nil {
			continue
		}
		if c.local != out.Number {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceWAFVersionV1Add(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	wafVer := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, "", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1AddExistingService(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	wafVer := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, "", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
				),
			},
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1Update(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	wafVerInput1 := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	wafVer1 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput1, "", "")

	wafVerInput2 := testAccFastlyServiceWAFVersionV1BuildConfig(22)
	wafVer2 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput2, "", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput1, 1),
				),
			},
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput2, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1Delete(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	wafVer := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, "", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput, 1),
				),
			},
			{
				Config: testAccFastlyServiceWAFVersionV1(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckEmpty(&service, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1Import(t *testing.T) {

	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	extraHCLMap := map[string]interface{}{
		"http_violation_score_threshold": 10,
	}

	rules := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "log",
			Revision: 1,
		},
		{
			ModSecID: 2037405,
			Status:   "log",
			Revision: 1,
		},
	}

	exclusions := []gofastly.WAFRuleExclusion{
		{
			Name:          gofastly.String("index page"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
			Condition:     gofastly.String("req.url.basename == \"index.html\""),
			Rules: []*gofastly.WAFRule{
				{
					ModSecID: 2029718,
				},
				{
					ModSecID: 2037405,
				},
			},
		},
		{
			Name:          gofastly.String("index php"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
			Condition:     gofastly.String("req.url.basename == \"index.php\""),
			Rules: []*gofastly.WAFRule{
				{
					ModSecID: 2037405,
				},
			},
		},
		{
			Name:          gofastly.String("index asp"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeWAF),
			Condition:     gofastly.String("req.url.basename == \"index.asp\""),
		},
	}

	rulesTF := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules)
	exclusionsTF := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRuleExclusions(exclusions)
	wafVer := testAccFastlyServiceWAFVersionV1ComposeConfiguration(extraHCLMap, rulesTF, exclusionsTF)
	wafSvcCfg := testAccFastlyServiceWAFVersionV1(name, wafVer)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: wafSvcCfg,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
				),
			},
			{
				ResourceName:      "fastly_service_waf_configuration.waf",
				ImportState:       true,
				ImportStateVerify: true,

				// Rule Exclusion should be ignored until it is in GA.
				ImportStateVerifyIgnore: []string{"rule_exclusion"},
			},
			{
				Config:   wafSvcCfg,
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckFastlyServiceWAFVersionV1CheckAttributes(service *gofastly.ServiceDetail, local map[string]interface{}, latestVersion int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(wafResp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
		}

		waf := wafResp.Items[0]
		verResp, err := conn.ListWAFVersions(&gofastly.ListWAFVersionsInput{
			WAFID: waf.ID,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF version records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(verResp.Items) < 1 {
			return fmt.Errorf("[ERR] Expected result size (%d), got (%d)", 1, len(verResp.Items))
		}

		latestVersion, err := testAccFastlyServiceWAFVersionV1GetVersionNumber(verResp.Items, latestVersion)
		if err != nil {
			return err
		}

		remote := testAccFastlyServiceWAFVersionV1ToMap(latestVersion)
		if !reflect.DeepEqual(remote, local) {
			return fmt.Errorf("Error matching:\nexpected: %#v\nand  got: %#v", local, remote)
		}
		return nil
	}
}

func testAccCheckFastlyServiceWAFVersionV1CheckEmpty(service *gofastly.ServiceDetail, latestVersion int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(wafResp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
		}

		waf := wafResp.Items[0]
		verResp, err := conn.ListWAFVersions(&gofastly.ListWAFVersionsInput{
			WAFID: waf.ID,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF version records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(verResp.Items) < 1 {
			return fmt.Errorf("[ERR] Expected result size (%d), got (%d)", 1, len(verResp.Items))
		}

		emptyVersion, err := testAccFastlyServiceWAFVersionV1GetVersionNumber(verResp.Items, latestVersion)
		if err != nil {
			return err
		}

		if !emptyVersion.Locked {
			return fmt.Errorf("[ERR] Expected Locked = (%v),  got (%v)", true, emptyVersion.Locked)
		}
		if emptyVersion.DeployedAt == nil {
			return fmt.Errorf("[ERR] Expected DeployedAt not nil,  got (%v)", emptyVersion.DeployedAt)
		}

		totalRules := emptyVersion.ActiveRulesFastlyBlockCount + emptyVersion.ActiveRulesFastlyLogCount + emptyVersion.ActiveRulesOWASPBlockCount +
			emptyVersion.ActiveRulesOWASPLogCount + emptyVersion.ActiveRulesOWASPScoreCount + emptyVersion.ActiveRulesTrustwaveBlockCount + emptyVersion.ActiveRulesTrustwaveLogCount

		if totalRules != 0 {
			return fmt.Errorf("expected no active rules rules: got %d", totalRules)
		}
		return nil
	}
}

func testAccFastlyServiceWAFVersionV1GetVersionNumber(versions []*gofastly.WAFVersion, number int) (gofastly.WAFVersion, error) {
	for _, v := range versions {
		if v.Number == number {
			return *v, nil
		}
	}
	return gofastly.WAFVersion{}, fmt.Errorf("version number %d not found", number)
}

func testAccFastlyServiceWAFVersionV1ComposeConfiguration(m map[string]interface{}, rules string, exclusions string) string {

	hcl := `
        resource "fastly_service_waf_configuration" "waf" {
          waf_id = fastly_service_v1.foo.waf[0].waf_id
         `
	for k, v := range m {

		switch t := reflect.TypeOf(v).String(); t {
		case "string":
			hcl = hcl + fmt.Sprintf(` %s = "%s"
         `, k, v)
		case "int":
			hcl = hcl + fmt.Sprintf(` %s = %d
         `, k, v)
		case "bool":
			hcl = hcl + fmt.Sprintf(` %s = %v
         `, k, v)
		}
	}
	return hcl + fmt.Sprintf(`
          %s

          %s
        }`, rules, exclusions)
}

func testAccFastlyServiceWAFVersionV1(name, extraHCL string) string {

	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

	# The WAF was updated to insert an ALWAYS_FALSE default condition, which 
	# broke our tests because the terraform state was unaware of the default 
	# condition resource that was being dynamically created by the API. This 
	# meant terraform would flag the difference in state as unexpected, and
	# subsequently produce an error.
	#
	# To resolve this error we define the default condition in our terraform which 
	# prevented the API from creating it, but there was a bug in the API 
	# implementation which meant the name of the condition had to match exactly
	# otherwise it would consider the condition missing.
	#
	# TODO(integralist):
	# Once the bug in the API has been fixed, come back and update the tests so 
	# that we can validate the test terraform code no longer requires the
	# condition name to be ALWAYS_FALSE (e.g. set the name to "foobar").
	#
	# NOTE:
	# If the WAF isn't in place and without that ALWAYS_FALSE condition, the WAF 
	# response object (403) will be served for all inbound traffic. This 
	# condition was added by the WAF team to prevent Fastly from returning a 403 
	# on all of the customer traffic before WAF is provisioned to the service.

	condition {
		name      = "ALWAYS_FALSE"
		priority  = 10
		statement = "!req.url"
		type      = "REQUEST"
	}

  condition {
	name = "prefetch"
	type = "PREFETCH"
	statement = "req.url~+\"index.html\""
  }

  response_object {
	name = "response"
	status = "403"
	response = "Forbidden"
	content = "content"
	request_condition = "ALWAYS_FALSE"
  }

  waf {
	prefetch_condition = "prefetch"
	response_object = "response"
  }

  force_destroy = true
}
  %s
`, name, domainName, backendName, extraHCL)
}

func testAccFastlyServiceWAFVersionV1BuildConfig(threshold int) map[string]interface{} {
	return map[string]interface{}{
		"allowed_http_versions":                "HTTP/1.0 HTTP/1.1",
		"allowed_methods":                      "GET HEAD POST",
		"allowed_request_content_type":         "application/x-www-form-urlencoded|multipart/form-data|text/xml|application/xml",
		"allowed_request_content_type_charset": "utf-8|iso-8859-1",
		"arg_length":                           800,
		"arg_name_length":                      200,
		"combined_file_sizes":                  20000000,
		"critical_anomaly_score":               12,
		"crs_validate_utf8_encoding":           true,
		"error_anomaly_score":                  10,
		"high_risk_country_codes":              "gb",
		"http_violation_score_threshold":       threshold,
		"inbound_anomaly_score_threshold":      threshold,
		"lfi_score_threshold":                  threshold,
		"max_file_size":                        20000000,
		"max_num_args":                         510,
		"notice_anomaly_score":                 8,
		"paranoia_level":                       2,
		"php_injection_score_threshold":        threshold,
		"rce_score_threshold":                  threshold,
		"restricted_extensions":                ".asa/ .asax/ .ascx/ .axd/ .backup/ .bak/ .bat/ .cdx/ .cer/ .cfg/ .cmd/ .com/",
		"restricted_headers":                   "/proxy/ /lock-token/",
		"rfi_score_threshold":                  threshold,
		"session_fixation_score_threshold":     threshold,
		"sql_injection_score_threshold":        threshold,
		"total_arg_length":                     12800,
		"warning_anomaly_score":                20,
		"xss_score_threshold":                  threshold,
	}
}

func testAccFastlyServiceWAFVersionV1ToMap(v gofastly.WAFVersion) map[string]interface{} {
	return map[string]interface{}{
		"allowed_http_versions":                v.AllowedHTTPVersions,
		"allowed_methods":                      v.AllowedMethods,
		"allowed_request_content_type":         v.AllowedRequestContentType,
		"allowed_request_content_type_charset": v.AllowedRequestContentTypeCharset,
		"arg_length":                           v.ArgLength,
		"arg_name_length":                      v.ArgNameLength,
		"combined_file_sizes":                  v.CombinedFileSizes,
		"critical_anomaly_score":               v.CriticalAnomalyScore,
		"crs_validate_utf8_encoding":           v.CRSValidateUTF8Encoding,
		"error_anomaly_score":                  v.ErrorAnomalyScore,
		"high_risk_country_codes":              v.HighRiskCountryCodes,
		"http_violation_score_threshold":       v.HTTPViolationScoreThreshold,
		"inbound_anomaly_score_threshold":      v.InboundAnomalyScoreThreshold,
		"lfi_score_threshold":                  v.LFIScoreThreshold,
		"max_file_size":                        v.MaxFileSize,
		"max_num_args":                         v.MaxNumArgs,
		"notice_anomaly_score":                 v.NoticeAnomalyScore,
		"paranoia_level":                       v.ParanoiaLevel,
		"php_injection_score_threshold":        v.PHPInjectionScoreThreshold,
		"rce_score_threshold":                  v.RCEScoreThreshold,
		"restricted_extensions":                v.RestrictedExtensions,
		"restricted_headers":                   v.RestrictedHeaders,
		"rfi_score_threshold":                  v.RFIScoreThreshold,
		"session_fixation_score_threshold":     v.SessionFixationScoreThreshold,
		"sql_injection_score_threshold":        v.SQLInjectionScoreThreshold,
		"total_arg_length":                     v.TotalArgLength,
		"warning_anomaly_score":                v.WarningAnomalyScore,
		"xss_score_threshold":                  v.XSSScoreThreshold,
	}
}
