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
	"path/filepath"
	"syscall"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var slayCmd = &cobra.Command{
	Use:   "slay [name]",
	Short: "Stop and clean up a daemon by name",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Italic.TextStyle(chalk.White.Color("lilith")) + ` slay gracefully stops a running daemon, removes its metadata file and its log file, allowing you to start fresh later.`,
	Example: chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("slay")) + " " +
		chalk.Cyan.Color("helix"),
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeDaemonNames,

	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.slay"
		name := args[0]

		// 1) Load metadata
		meta, err := loadMeta(name)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)))

		// 2) Signal the process to terminate
		proc, err := os.FindProcess(meta.PID)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("finding process"))
		horus.CheckErr(proc.Signal(syscall.SIGTERM), horus.WithOp(op), horus.WithMessage("sending SIGTERM"))

		// 3) Remove the metadata JSON file
		metaFile := filepath.Join(getDaemonDir(), name+".json")
		_, err = domovoi.RemoveFile(metaFile)(metaFile)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("removing metadata file"))

		// 4) Remove the log file
		_, err = domovoi.RemoveFile(meta.LogPath)(meta.LogPath)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("removing log file"))

		// 5) Final confirmation
		fmt.Printf("%s slayed daemon %q\n", chalk.Green.Color("OK:"), name)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(slayCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
