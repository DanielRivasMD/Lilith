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
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rekindleCmd = &cobra.Command{
	Use:   "rekindle " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short: "Restart a stopped daemon",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Italic.TextStyle(chalk.Blue.Color("lilith")) + ` restarts a stopped watcher daemon using its saved metadata`,
	Example: chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("rekindle")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeDaemonNames,

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.rekindle"
		name := args[0]

		// load existing metadata
		meta, err := loadMeta(name)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)))

		// spawn new watcher
		pid, err := spawnWatcher(meta)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("restarting watcher"))

		// update and save metadata
		meta.PID = pid
		meta.InvokedAt = time.Now()
		horus.CheckErr(saveMeta(meta), horus.WithOp(op), horus.WithMessage("updating metadata"))

		fmt.Printf("%s rekindled %q with PID %d\n",
			chalk.Green.Color("OK:"), name, pid)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(rekindleCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
