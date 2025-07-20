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
	"syscall"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var freezeCmd = &cobra.Command{
	Use:   "freeze [name]",
	Short: "Pause a daemon",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Italic.TextStyle(chalk.Blue.Color("lilith")) + ` freeze sends SIGSTOP to a running daemon, pausing its execution until you explicitly resume it via OS tools.`,
	Example: chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("freeze")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeDaemonNames,

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.freeze"
		name := args[0]

		// 1) Load metadata
		meta, err := loadMeta(name)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)))

		// 2) Find and pause process
		proc, err := os.FindProcess(meta.PID)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("finding process"))
		horus.CheckErr(proc.Signal(syscall.SIGSTOP), horus.WithOp(op), horus.WithMessage("sending SIGSTOP"))

		// 3) Confirmation
		fmt.Printf("%s froze daemon %q\n", chalk.Green.Color("OK:"), name)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(freezeCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
