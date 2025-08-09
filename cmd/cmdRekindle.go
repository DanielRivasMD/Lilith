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
	"path/filepath"
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rekindleCmd = &cobra.Command{
	Use:     "rekindle " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short:   "Resurrect daemon in limbo",
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

var helpRekindle = chalk.Bold.TextStyle(chalk.Green.Color("Daniel Rivas ")) +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) +
	chalk.Dim.TextStyle(chalk.Cyan.Color("\n\nrestart daemons in limbo using persisted metadata"))

var exampleRekindle = chalk.White.Color("lilith") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("rekindle")) + " " +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")) + "\n" +
	chalk.White.Color("lilith") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("rekindle")) + " " +
	chalk.Italic.TextStyle(chalk.White.Color("--group")) + " " +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<forge>")) + "\n" +
	chalk.White.Color("lilith") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("rekindle")) + " " +
	chalk.Italic.TextStyle(chalk.White.Color("--all"))

////////////////////////////////////////////////////////////////////////////////////////////////////

func runRekindle(cmd *cobra.Command, args []string) {
	const op = "lilith.rekindle"

	switch {
	case rekindleAll:
		rekindleAllDaemons()
		return

	case rekindleGroup != "":
		rekindleGroupDaemons(rekindleGroup)
		return

	case len(args) == 1:
		name := args[0]
		meta := mustLoadMeta(filepath.Join(GetDaemonDir(), name))
		pid := mustSpawnWatcher(meta)
		meta.PID = pid
		meta.InvokedAt = time.Now()
		horus.CheckErr(SaveMeta(&meta), horus.WithOp(op), horus.WithMessage("updating metadata"))
		fmt.Printf("%s rekindled %q with PID %d\n", chalk.Green.Color("OK:"), name, pid)
		return

	default:
		horus.CheckErr(horus.NewCategorizedHerror(op, "validation", "missing daemon name or flag", nil, nil))
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func rekindleAllDaemons() {
	for _, path := range MustListDaemonMetaFiles() {
		meta := mustLoadMeta(path)
		pid := mustSpawnWatcher(meta)
		meta.PID = pid
		meta.InvokedAt = time.Now()
		horus.CheckErr(SaveMeta(&meta))
		fmt.Printf("%s rekindled %q with PID %d\n", chalk.Green.Color("OK:"), meta.Name, pid)
	}
}

func rekindleGroupDaemons(group string) {
	for _, path := range MustListDaemonMetaFiles() {
		if matchesGroup(path, group) {
			meta := mustLoadMeta(path)
			pid := mustSpawnWatcher(meta)
			meta.PID = pid
			meta.InvokedAt = time.Now()
			horus.CheckErr(SaveMeta(&meta))
			fmt.Printf("%s rekindled %q with PID %d\n", chalk.Green.Color("OK:"), meta.Name, pid)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
