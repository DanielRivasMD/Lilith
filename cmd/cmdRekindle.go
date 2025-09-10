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

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rekindleCmd = &cobra.Command{
	Use:     "rekindle " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short:   "Resurrect daemon from limbo",
	Long:    helpRekindle,
	Example: exampleRekindle,

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeDaemonNames,

	Run: runRekindle,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	rekindleGroup string
	rekindleAll   bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(rekindleCmd)

	rekindleCmd.Flags().BoolVar(&rekindleAll, "all", false, "Rekindle all dead daemons")
	rekindleCmd.Flags().StringVar(&rekindleGroup, "group", "", "Rekindle all daemons in a specific group")

	horus.CheckErr(rekindleCmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("rekindle.init"), horus.WithMessage("registering config completion"))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runRekindle(cmd *cobra.Command, args []string) {
	const op = "lilith.rekindle"

	switch {
	case rekindleAll:
		rekindleAllDaemons()
	case rekindleGroup != "":
		rekindleGroupDaemons(rekindleGroup)
	case len(args) == 1:
		rekindleDaemon(args[0])
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

func rekindleDaemon(daemonMeta string) {
	meta := loadMeta(daemonMeta)

	// try to find & ping the existing process
	if proc, err := os.FindProcess(meta.PID); err == nil {
		if err = proc.Signal(syscall.Signal(0)); err == nil {
			// process alive => send SIGCONT
			sendDaemonSignal(meta.PID, syscall.SIGCONT)

			// update only timestamp
			meta.InvokedAt = time.Now()
			saveMeta(meta)

			fmt.Printf("%s resumed %q PID %d\n",
				chalk.Green.Color("OK:"),
				meta.Name,
				meta.PID,
			)
			return
		}
	}

	// no existing process => spawn fresh watcher
	newPID := spawnWatcher(meta)

	// update metadata
	meta.PID = newPID
	meta.InvokedAt = time.Now()
	saveMeta(meta)

	fmt.Printf("%s rekindled %q new PID %d\n",
		chalk.Green.Color("OK:"),
		meta.Name,
		newPID,
	)
}

func rekindleGroupDaemons(group string) {
	for _, daemonMeta := range listDaemonMetaFiles() {
		if matchDaemonGroup(daemonMeta, group) {
			rekindleDaemon(daemonMeta)
		}
	}
}

func rekindleAllDaemons() {
	for _, daemonMeta := range listDaemonMetaFiles() {
		rekindleDaemon(daemonMeta)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
