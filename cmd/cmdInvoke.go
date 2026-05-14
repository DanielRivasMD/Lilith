/*
Copyright © 2025 Daniel Rivas <danielrivasmd@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var invokeFlags struct {
	daemonName   string
	daemonGroup  string
	daemonWatch  string
	daemonScript string
	daemonLog    string
	group        string
	all          bool
	once         bool
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: clean daemons run once
func InvokeCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("invoke", runInvoke,
		domovoi.WithValidArgsFunction(completeWorkflowNames),
	))

	// DOC: manual is mutually exclusive with toml file mode
	// manual mode flags
	cmd.Flags().StringVarP(&invokeFlags.daemonName, "daemon-name", "", "", "Daemon instance name (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.daemonGroup, "daemon-group", "", "default", "Group name for this daemon (overrides TOML). Default: `default`")
	cmd.Flags().StringVarP(&invokeFlags.daemonWatch, "daemon-watch", "", "", "Directory to watch (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.daemonScript, "daemon-script", "", "", "Script to execute on change (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.daemonLog, "daemon-log", "", "", "Name for log file (no .log extension; required in manual mode)")

	// group / all / once modes
	cmd.Flags().StringVarP(&invokeFlags.group, "group", "", "", "Invoke all workflows in this group (config file name without .toml)")
	cmd.Flags().BoolVarP(&invokeFlags.all, "all", "", false, "Invoke all workflows from all config files")
	cmd.Flags().BoolVarP(&invokeFlags.once, "once", "", false, "Run script once and exit (no watching, no persistent daemon)")

	// completion for --group flag
	horus.CheckErr(cmd.RegisterFlagCompletionFunc("group", completeConfigGroups), horus.WithOp("invoke.init"), horus.WithMessage("registering group completion"))

	cmd.PreRun = preInvokeManual

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func preInvokeManual(cmd *cobra.Command, args []string) {
	if invokeFlags.group != "" || invokeFlags.all || len(args) == 1 {
		return
	}

	// manual mode: require daemon, watch, script, log
	if rootFlags.verbose {
		fmt.Println("Running on Manual mode...")
	}
	horus.CheckEmpty(
		invokeFlags.daemonName,
		"",
		horus.WithMessage("`--daemon` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
	horus.CheckEmpty(
		invokeFlags.daemonWatch,
		"",
		horus.WithMessage("`--daemon-watch` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
	horus.CheckEmpty(
		invokeFlags.daemonScript,
		"",
		horus.WithMessage("`--daemon-script` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
	horus.CheckEmpty(
		invokeFlags.daemonLog,
		"",
		horus.WithMessage("`--log` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: use op in invoke functions
func runInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke"

	// mode 0: --once
	if invokeFlags.once {
		runOnce(args)
		return
	}

	// mode 1: --all
	if invokeFlags.all {
		invokeAllWorkflows()
		return
	}

	// mode 2: --group <name>
	if invokeFlags.group != "" {
		invokeGroupWorkflows(invokeFlags.group)
		return
	}

	// mode 3: config mode
	if len(args) >= 1 {
		workflowNames := args
		invokeNamedWorkflows(workflowNames)
		return
	}

	// mode 4: manual mode
	invokeManual()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
