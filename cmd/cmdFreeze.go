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
	"errors"
	"fmt"
	"syscall"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var freezeCmd = &cobra.Command{
	Use:     "freeze " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short:   "Pause daemon",
	Long:    helpFreeze,
	Example: exampleFreeze,

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeDaemonNames,

	Run: runFreeze,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(freezeCmd)

	freezeCmd.Flags().String("group", "", "Freeze all daemons belonging to a specific group")
	freezeCmd.Flags().Bool("all", false, "Freeze all running daemons")

	horus.CheckErr(freezeCmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("freeze.init"), horus.WithMessage("registering config completion"))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runFreeze(cmd *cobra.Command, args []string) {
	const op = "lilith.freeze"

	group, _ := cmd.Flags().GetString("group")
	all, _ := cmd.Flags().GetBool("all")

	switch {
	case all:
		freezeAllDaemons()
	case group != "":
		freezeGroupDaemons(group)
	case len(args) == 1:
		freezeDaemon(args[0])
	default:
		horus.CheckErr(
			errors.New(""),
			horus.WithOp(op),
			horus.WithMessage("daemon / flag"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string {
				return "missing " + errorFmt(he.Message)
			}),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func freezeDaemon(daemonMeta string) {
	meta := loadMeta(daemonMeta)
	sendDaemonSignal(meta.PID, syscall.SIGSTOP)
	fmt.Printf("%s froze daemon %q\n", chalk.Green.Color("OK:"), meta.Name)
}

func freezeGroupDaemons(group string) {
	for _, daemonMeta := range listDaemonMetaFiles() {
		if matchDaemonGroup(daemonMeta, group) {
			freezeDaemon(daemonMeta)
		}
	}
}

func freezeAllDaemons() {
	for _, daemonMeta := range listDaemonMetaFiles() {
		freezeDaemon(daemonMeta)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
