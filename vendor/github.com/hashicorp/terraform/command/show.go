package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/plans/planfile"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/hashicorp/terraform/tfdiags"

	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/command/jsonplan"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/states"
)

// ShowCommand is a Command implementation that reads and outputs the
// contents of a Terraform plan or state file.
type ShowCommand struct {
	Meta
}

func (c *ShowCommand) Run(args []string) int {
	args, err := c.Meta.process(args, false)
	if err != nil {
		return 1
	}

	cmdFlags := c.Meta.defaultFlagSet("show")
	var jsonOutput bool
	cmdFlags.BoolVar(&jsonOutput, "json", false, "produce JSON output (only available when showing a planfile)")
	cmdFlags.Usage = func() { c.Ui.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	args = cmdFlags.Args()
	if len(args) > 2 {
		c.Ui.Error(
			"The show command expects at most two arguments.\n The path to a " +
				"Terraform state or plan file, and optionally -json for json output.\n")
		cmdFlags.Usage()
		return 1
	}

	var diags tfdiags.Diagnostics

	// Load the backend
	b, backendDiags := c.Backend(nil)
	diags = diags.Append(backendDiags)
	if backendDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// We require a local backend
	local, ok := b.(backend.Local)
	if !ok {
		c.showDiagnostics(diags) // in case of any warnings in here
		c.Ui.Error(ErrUnsupportedLocalOp)
		return 1
	}

	// the show command expects the config dir to always be the cwd
	cwd, err := os.Getwd()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting cwd: %s", err))
		return 1
	}

	// Determine if a planfile was passed to the command
	var planFile *planfile.Reader
	if len(args) > 0 {
		// We will handle error checking later on - this is just required to
		// load the local context if the given path is successfully read as
		// a planfile.
		planFile, _ = c.PlanFile(args[0])
	}

	// Build the operation
	opReq := c.Operation(b)
	opReq.ConfigDir = cwd
	opReq.PlanFile = planFile
	opReq.ConfigLoader, err = c.initConfigLoader()
	if err != nil {
		diags = diags.Append(err)
		c.showDiagnostics(diags)
		return 1
	}

	// Get the context
	ctx, _, ctxDiags := local.Context(opReq)
	diags = diags.Append(ctxDiags)
	if ctxDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// Get the schemas from the context
	schemas := ctx.Schemas()

	var planErr, stateErr error
	var plan *plans.Plan
	var state *states.State

	// if a path was provided, try to read it as a path to a planfile
	// if that fails, try to read the cli argument as a path to a statefile
	if len(args) > 0 {
		path := args[0]
		plan, planErr = getPlanFromPath(path)
		if planErr != nil {
			// json output is only supported for plans
			if jsonOutput == true {
				c.Ui.Error("Error: JSON output not available for state")
				return 1
			}
			state, stateErr = getStateFromPath(path)
			if stateErr != nil {
				c.Ui.Error(fmt.Sprintf(
					"Terraform couldn't read the given file as a state or plan file.\n"+
						"The errors while attempting to read the file as each format are\n"+
						"shown below.\n\n"+
						"State read error: %s\n\nPlan read error: %s",
					stateErr,
					planErr))
				return 1
			}
		}
	}

	if state == nil {
		env := c.Workspace()
		state, stateErr = getStateFromEnv(b, env)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	}

	// This is an odd-looking check, because it's ok if we have a plan and an
	// empty state, and we've already validated that any command-line arguments
	// have been read successfully
	if plan == nil && state == nil {
		c.Ui.Output("No state.")
		return 0
	}

	if plan != nil {
		if jsonOutput == true {
			config := ctx.Config()
			jsonPlan, err := jsonplan.Marshal(config, plan, state, schemas)
			if err != nil {
				c.Ui.Error(fmt.Sprintf("Failed to marshal plan to json: %s", err))
				return 1
			}
			c.Ui.Output(string(jsonPlan))
			return 0
		}
		dispPlan := format.NewPlan(plan.Changes)
		c.Ui.Output(dispPlan.Format(c.Colorize()))
		return 0
	}

	c.Ui.Output(format.State(&format.StateOpts{
		State:   state,
		Color:   c.Colorize(),
		Schemas: schemas,
	}))
	return 0
}

func (c *ShowCommand) Help() string {
	helpText := `
Usage: terraform show [options] [path]

  Reads and outputs a Terraform state or plan file in a human-readable
  form. If no path is specified, the current state will be shown.

Options:

  -no-color           If specified, output won't contain any color.
  -json				  If specified, output the Terraform plan in a machine-
						readable form. Only available for plan files.

`
	return strings.TrimSpace(helpText)
}

func (c *ShowCommand) Synopsis() string {
	return "Inspect Terraform state or plan"
}

// getPlanFromPath returns a plan if the user-supplied path points to a planfile.
// If both plan and error are nil, the path is likely a directory.
// An error could suggest that the given path points to a statefile.
func getPlanFromPath(path string) (*plans.Plan, error) {
	pr, err := planfile.Open(path)
	if err != nil {
		return nil, err
	}
	plan, err := pr.ReadPlan()
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// getStateFromPath returns a State if the user-supplied path points to a statefile.
func getStateFromPath(path string) (*states.State, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error loading statefile: %s", err)
	}
	defer f.Close()

	var stateFile *statefile.File
	stateFile, err = statefile.Read(f)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s as a statefile: %s", path, err)
	}
	return stateFile.State, nil
}

// getStateFromEnv returns the State for the current workspace, if available.
func getStateFromEnv(b backend.Backend, env string) (*states.State, error) {
	// Get the state
	stateStore, err := b.StateMgr(env)
	if err != nil {
		return nil, fmt.Errorf("Failed to load state manager: %s", err)
	}

	if err := stateStore.RefreshState(); err != nil {
		return nil, fmt.Errorf("Failed to load state: %s", err)
	}

	state := stateStore.State()
	return state, nil
}
