package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccFastlyServiceWAFVersionDeploymentStatus(t *testing.T) {

	wafID := "waf-id"
	latestVersion := &gofastly.WAFVersion{}
	d := &schema.ResourceData{}

	cases := []struct {
		status      string
		ExpectError bool
	}{
		{
			status:      "",
			ExpectError: true,
		},
		{
			status:      gofastly.WAFVersionDeploymentStatusFailed,
			ExpectError: true,
		},
		{
			status:      gofastly.WAFVersionDeploymentStatusCompleted,
			ExpectError: false,
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Status %v", c.status), func(t *testing.T) {
			statusCheck := &WAFDeploymentChecker{
				Timeout:    d.Timeout(schema.TimeoutCreate),
				MinTimeout: 0,
				Delay:      0,
				Check: func(wafID string, version int) (*gofastly.WAFVersion, error) {
					return &gofastly.WAFVersion{
						LastDeploymentStatus: c.status,
					}, nil
				},
			}
			err := statusCheck.waitForDeployment(wafID, latestVersion)
			hasErrored := err != nil
			if c.ExpectError && !hasErrored {
				t.Fatalf("Error expected to be %v but wan't", c.ExpectError)
			}
			if !c.ExpectError && hasErrored {
				t.Fatalf("Error expected to be %v but wan't. Error: %v", c.ExpectError, err)
			}
		})
	}
}
