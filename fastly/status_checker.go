package fastly

import (
	"context"
	"fmt"
	"time"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	WAFStatusCheckDelay      = 5 * time.Second
	WAFStatusCheckMinTimeout = 5 * time.Second
)

type WAFDeploymentStatusCheck func(wafID string, version int) (*gofastly.WAFVersion, error)

type WAFDeploymentChecker struct {
	Timeout    time.Duration
	Delay      time.Duration
	MinTimeout time.Duration
	Check      WAFDeploymentStatusCheck
}

func DefaultWAFDeploymentChecker(conn *gofastly.Client) func(wafID string, version int) (*gofastly.WAFVersion, error) {
	checkDeploymentStatus := func(wafID string, version int) (*gofastly.WAFVersion, error) {
		resp, err := conn.GetWAFVersion(&gofastly.GetWAFVersionInput{
			WAFID:            wafID,
			WAFVersionNumber: version,
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return checkDeploymentStatus
}

func (c *WAFDeploymentChecker) waitForDeployment(ctx context.Context, wafID string, latestVersion *gofastly.WAFVersion) error {
	createStateConf := &resource.StateChangeConf{
		Pending: []string{
			gofastly.WAFVersionDeploymentStatusPending,
			gofastly.WAFVersionDeploymentStatusInProgress,
		},
		Target: []string{
			gofastly.WAFVersionDeploymentStatusCompleted,
		},
		Refresh: func() (interface{}, string, error) {
			res, err := c.Check(wafID, latestVersion.Number)
			if err != nil {
				return nil, "", err
			}
			if res.LastDeploymentStatus == gofastly.WAFVersionDeploymentStatusFailed {
				return res, res.LastDeploymentStatus, fmt.Errorf("WAF deployment failed. Error message: %v", res.Error)
			}
			return res, res.LastDeploymentStatus, nil
		},
		Timeout:                   c.Timeout,
		Delay:                     c.Delay,
		MinTimeout:                c.MinTimeout,
		ContinuousTargetOccurence: 5,
		NotFoundChecks:            1,
	}

	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("Error waiting for WAF Version (%s) to be updated: %v", wafID, err)
	}
	return nil
}
