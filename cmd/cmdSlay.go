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

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var slayCmd = &cobra.Command{
	Use:     "slay " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short:   "Kill daemon",
	Long:    helpSlay,
	Example: exampleSlay,

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeDaemonNames,

	Run: runSlay,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(slayCmd)

	slayCmd.Flags().BoolVar(&flags.slayAll, "all", false, "Slay all daemons")
	slayCmd.Flags().StringVar(&flags.slayGroup, "group", "", "Slay all daemons in a specific group")

	horus.CheckErr(slayCmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("slay.init"), horus.WithMessage("registering config completion"))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSlay(cmd *cobra.Command, args []string) {
	const op = "lilith.slay"

	switch {
	case flags.slayAll:
		slayAllDaemons()
	case flags.slayGroup != "":
		slayGroupDaemons(flags.slayGroup)
	case len(args) == 1:
		slayDaemon(args[0])
	default:
		horus.CheckErr(
			errors.New(""),
			horus.WithOp(op),
			horus.WithMessage("daemon / flag"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string {
				return "missing " + onelineErr(he.Message)
			}),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: if process finish, clean up
func slayDaemon(daemonMeta string) {
	const op = "lilith.slay"

	// load metadata
	meta := loadMeta(daemonMeta)

	// try terminating process, but proceed if already gone
	sendDaemonSignal(meta.PID, syscall.SIGTERM)

	// remove metadata JSON file
	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(daemonMeta, flags.verbose)(resolveMetaPath(daemonMeta))
			return err
		}(),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("removing metadata file"),
	)

	// remove log file
	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(meta.LogPath, flags.verbose)(meta.LogPath)
			return err
		}(),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("removing log file"),
	)

	// log meta
	fmt.Printf("%s slayed daemon %q\n", chalk.Green.Color("OK:"), meta.Daemon)
}

func slayGroupDaemons(group string) {
	for _, daemonMeta := range listDaemonMetaFiles() {
		if matchDaemonGroup(daemonMeta, group) {
			slayDaemon(daemonMeta)
		}
	}
}

func slayAllDaemons() {
	for _, daemonMeta := range listDaemonMetaFiles() {
		slayDaemon(daemonMeta)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
