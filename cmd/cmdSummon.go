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
	"os"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var summonCmd = &cobra.Command{
	Use:     "summon " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short:   "View daemon logs",
	Long:    helpSummon,
	Example: exampleSummon,

	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeDaemonNames,

	Run: runSummon,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	follow bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(summonCmd)
	summonCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Continuously watch the log file")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpSummon = chalk.Bold.TextStyle(chalk.Green.Color("Daniel Rivas ")) +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) +
	chalk.Dim.TextStyle(chalk.Cyan.Color("\n\ndisplay a daemon's log output - pass --follow to stream in real time"))

var exampleSummon = chalk.White.Color("lilith") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("summon")) + " " +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")) + " " +
	chalk.Italic.TextStyle(chalk.White.Color("--follow"))

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSummon(cmd *cobra.Command, args []string) {
	const op = "lilith.summon"
	name := args[0]

	meta, err := loadMeta(name)
	horus.CheckErr(err,
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)),
	)

	if follow {
		horus.CheckErr(
			domovoi.ExecCmd("tail", "-f", meta.LogPath),
			horus.WithOp(op),
			horus.WithMessage("streaming log"),
		)
	} else {
		pager := os.Getenv("PAGER")
		if pager == "" {
			pager = "less"
		}
		horus.CheckErr(
			domovoi.ExecCmd(pager, "--paging", "always", meta.LogPath),
			horus.WithOp(op),
			horus.WithMessage("paging log"),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
