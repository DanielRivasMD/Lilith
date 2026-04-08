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

var slayFlags struct {
	all   bool
	group string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func SlayCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("slay", runSlay,
		domovoi.WithValidArgsFunction(completeDaemonNames),
	))

	cmd.Flags().BoolVarP(&slayFlags.all, "all", "", false, "Slay all daemons")
	cmd.Flags().StringVarP(&slayFlags.group, "group", "", "", "Slay all daemons in a specific group")

	horus.CheckErr(cmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("slay.init"), horus.WithMessage("registering config completion"))

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSlay(cmd *cobra.Command, args []string) {
	const op = "lilith.slay"

	switch {
	case slayFlags.all:
		slayAllDaemons()
	case slayFlags.group != "":
		slayGroupDaemons(slayFlags.group)
	case len(args) == 1:
		slayDaemon(args[0])
	default:
		horus.CheckErr(
			errors.New(""),
			horus.WithOp(op),
			horus.WithMessage("daemon / flag"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string {
				return "missing " + horus.OneLineErr(he.Message)
			}),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func slayDaemon(daemonMeta string) {
	const op = "lilith.slay"
	meta := loadMeta(daemonMeta)

	sendDaemonSignal(meta.PID, syscall.SIGTERM)

	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(daemonMeta, rootFlags.verbose)(resolveMetaPath(daemonMeta))
			return err
		}(),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("removing metadata file"),
	)

	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(meta.LogPath, rootFlags.verbose)(meta.LogPath)
			return err
		}(),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("removing log file"),
	)

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
