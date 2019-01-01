package command

import (
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/state"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statemgr"

	backendLocal "github.com/hashicorp/terraform/backend/local"
)

// StateMeta is the meta struct that should be embedded in state subcommands.
type StateMeta struct {
	Meta
}

// State returns the state for this meta. This gets the appropriate state from
// the backend, but changes the way that backups are done. This configures
// backups to be timestamped rather than just the original state path plus a
// backup path.
func (c *StateMeta) State() (state.State, error) {
	var realState state.State
	backupPath := c.backupPath
	stateOutPath := c.statePath

	// use the specified state
	if c.statePath != "" {
		realState = statemgr.NewFilesystem(c.statePath)
	} else {
		// Load the backend
		b, backendDiags := c.Backend(nil)
		if backendDiags.HasErrors() {
			return nil, backendDiags.Err()
		}

		workspace := c.Workspace()
		// Get the state
		s, err := b.StateMgr(workspace)
		if err != nil {
			return nil, err
		}

		// Get a local backend
		localRaw, backendDiags := c.Backend(&BackendOpts{ForceLocal: true})
		if backendDiags.HasErrors() {
			// This should never fail
			panic(backendDiags.Err())
		}
		localB := localRaw.(*backendLocal.Local)
		_, stateOutPath, _ = localB.StatePaths(workspace)
		if err != nil {
			return nil, err
		}

		realState = s
	}

	// We always backup state commands, so set the back if none was specified
	// (the default is "-", but some tests bypass the flag parsing).
	if backupPath == "-" || backupPath == "" {
		// Determine the backup path. stateOutPath is set to the resulting
		// file where state is written (cached in the case of remote state)
		backupPath = fmt.Sprintf(
			"%s.%d%s",
			stateOutPath,
			time.Now().UTC().Unix(),
			DefaultBackupExtension)
	}

	// If the backend is local (which it should always be, given our asserting
	// of it above) we can now enable backups for it.
	if lb, ok := realState.(*statemgr.Filesystem); ok {
		lb.SetBackupPath(backupPath)
	}

	return realState, nil
}

func (c *StateMeta) filter(state *states.State, args []string) ([]*states.FilterResult, error) {
	var results []*states.FilterResult

	filter := &states.Filter{State: state}
	for _, arg := range args {
		filtered, err := filter.Filter(arg)
		if err != nil {
			return nil, err
		}

	filtered:
		for _, result := range filtered {
			switch result.Address.(type) {
			case addrs.ModuleInstance:
				for _, result := range filtered {
					if _, ok := result.Address.(addrs.ModuleInstance); ok {
						results = append(results, result)
					}
				}
				break filtered
			case addrs.AbsResource:
				for _, result := range filtered {
					if _, ok := result.Address.(addrs.AbsResource); ok {
						results = append(results, result)
					}
				}
				break filtered
			case addrs.AbsResourceInstance:
				results = append(results, result)
			}
		}
	}

	// Sort the results
	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]

		// If the length is different, sort on the length so that the
		// best match is the first result.
		if len(a.Address.String()) != len(b.Address.String()) {
			return len(a.Address.String()) < len(b.Address.String())
		}

		// If the addresses are different it is just lexographic sorting
		if a.Address.String() != b.Address.String() {
			return a.Address.String() < b.Address.String()
		}

		// Addresses are the same, which means it matters on the type
		return a.SortedType() < b.SortedType()
	})

	return results, nil
}
