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
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var slayCmd = &cobra.Command{
	Use:     "slay " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short:   "Stop & clean up daemons",
	Long:    helpSlay,
	Example: exampleSlay,

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeDaemonNames,

	Run: runSlay,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	slayAll   bool
	slayGroup string
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(slayCmd)

	slayCmd.Flags().BoolVar(&slayAll, "all", false, "Slay all daemons")
	slayCmd.Flags().StringVar(&slayGroup, "group", "", "Slay all daemons in a specific group")

	horus.CheckErr(
		slayCmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups),
		horus.WithOp("slay.init"),
		horus.WithMessage("registering config completion"),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpSlay = chalk.Bold.TextStyle(chalk.Green.Color("Daniel Rivas ")) +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) +
	chalk.Dim.TextStyle(chalk.Cyan.Color("\n\ngracefully stop alive daemons, removing their metadata and logs to allow clean reinvocation"))

var exampleSlay = chalk.White.Color("lilith") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("slay")) + " " +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")) + "\n" +
	chalk.White.Color("lilith") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("slay")) + " " +
	chalk.Italic.TextStyle(chalk.White.Color("--group")) + " " +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<forge>")) + "\n" +
	chalk.White.Color("lilith") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("slay")) + " " +
	chalk.Italic.TextStyle(chalk.White.Color("--all"))

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSlay(cmd *cobra.Command, args []string) {
	const op = "lilith.slay"

	switch {
	case slayAll:
		slayAllDaemons()
	case slayGroup != "":
		slayGroupDaemons(slayGroup)
	case len(args) == 1:
		slaySingleDaemon(args[0])
	default:
		// TODO: refactor error message as one liner
		horus.CheckErr(
			horus.NewCategorizedHerror(
				op,
				"validation",
				"must provide a daemon name or --all / --group",
				nil,
				nil,
			),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func slaySingleDaemon(name string) {
	const op = "lilith.slay"

	// 1) Load metadata
	meta, err := loadMeta(name)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)),
	)

	// 2) Try terminating the process, but proceed if it’s already gone
	horus.CheckErr(
		terminate(meta.PID),
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("terminating PID %d", meta.PID)),
	)

	// 3) Remove the metadata JSON file
	metaFile := filepath.Join(GetDaemonDir(), name+".json")
	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(metaFile, verbose)(metaFile)
			return err
		}(),
		horus.WithOp(op),
		horus.WithMessage("removing metadata file"),
	)

	// 4) Remove the log file
	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(meta.LogPath, verbose)(meta.LogPath)
			return err
		}(),
		horus.WithOp(op),
		horus.WithMessage("removing log file"),
	)

	// 5) Final confirmation
	fmt.Printf("%s slayed daemon %q\n", chalk.Green.Color("OK:"), name)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// terminate sends SIGTERM to pid. Returns nil if the process is already gone.
func terminate(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// On Unix: ESRCH == “no such process”
		// os.ErrProcessDone == “process already finished”
		// On Windows: Signal isn't really supported, so ignore all errors
		switch {
		case errors.Is(err, syscall.ESRCH),
			errors.Is(err, os.ErrProcessDone),
			runtime.GOOS == "windows":
			return nil
		default:
			return fmt.Errorf("signal SIGTERM to %d: %w", pid, err)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func slayAllDaemons() {
	files := mustListDaemonMetaFiles()
	for _, file := range files {
		name := nameFrom(file)
		slaySingleDaemon(name)
	}
}

func slayGroupDaemons(group string) {
	files := mustListDaemonMetaFiles()
	for _, metaPath := range files {
		if matchesGroup(metaPath, group) {
			name := nameFrom(metaPath)
			slaySingleDaemon(name)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
