package fastly

import (
	"context"
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccFastlyServiceWAFVersion_DeploymentStatus(t *testing.T) {
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
			err := statusCheck.waitForDeployment(context.Background(), wafID, latestVersion)
			hasErrored := err != nil
			if c.ExpectError && !hasErrored {
				t.Fatalf("Error expected to be %v", c.ExpectError)
			}
			if !c.ExpectError && hasErrored {
				t.Fatalf("Error expected to be %v. Error: %v", c.ExpectError, err)
			}
		})
	}
}
