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
		return
	case group != "":
		freezeGroupDaemons(group)
		return
	default:
		// Single daemon freeze
		freezeDaemon(args[0])
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func freezeGroupDaemons(group string) {
	files := listDaemonMetaFiles()
	for _, path := range files {
		if matchDaemonGroup(path, group) {
			freezeDaemon(path)
		}
	}
}

func freezeAllDaemons() {
	files := listDaemonMetaFiles()
	for _, path := range files {
		freezeDaemon(path)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func freezeDaemon(name string) {
	meta := loadMeta(name)
	sendDaemonSignal(meta.PID, syscall.SIGSTOP)
	fmt.Printf("%s froze daemon %q\n", chalk.Green.Color("OK:"), meta.Name)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
