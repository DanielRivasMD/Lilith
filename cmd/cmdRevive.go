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
	"os"
	"syscall"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var reviveFlags struct {
	all   bool
	group string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func ReviveCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("revive", runRevive,
		domovoi.WithValidArgsFunction(completeDaemonNames),
	))

	cmd.Flags().BoolVarP(&reviveFlags.all, "all", "", false, "Revive all dead daemons")
	cmd.Flags().StringVarP(&reviveFlags.group, "group", "", "", "Revive all daemons in a specific group")

	horus.CheckErr(cmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("revive.init"), horus.WithMessage("registering config completion"))

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runRevive(cmd *cobra.Command, args []string) {
	const op = "lilith.revive"

	switch {
	case reviveFlags.all:
		reviveAllDaemons()
	case reviveFlags.group != "":
		reviveGroupDaemons(reviveFlags.group)
	case len(args) == 1:
		reviveDaemon(args[0])
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

func reviveDaemon(daemonMeta string) {
	meta := loadMeta(daemonMeta)

	if proc, err := os.FindProcess(meta.PID); err == nil {
		if err = proc.Signal(syscall.Signal(0)); err == nil {
			sendDaemonSignal(meta.PID, syscall.SIGCONT)
			meta.InvokedAt = time.Now()
			saveMeta(meta)
			fmt.Printf("%s resumed %q PID %d\n",
				chalk.Green.Color("OK:"),
				meta.Daemon,
				meta.PID,
			)
			return
		}
	}

	newPID := spawnWatcher(meta)
	meta.PID = newPID
	meta.InvokedAt = time.Now()
	saveMeta(meta)

	fmt.Printf("%s revived %q new PID %d\n",
		chalk.Green.Color("OK:"),
		meta.Daemon,
		newPID,
	)
}

func reviveGroupDaemons(group string) {
	for _, daemonMeta := range listDaemonMetaFiles() {
		if matchDaemonGroup(daemonMeta, group) {
			reviveDaemon(daemonMeta)
		}
	}
}

func reviveAllDaemons() {
	for _, daemonMeta := range listDaemonMetaFiles() {
		reviveDaemon(daemonMeta)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
