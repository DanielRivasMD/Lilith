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
	"strings"
	"syscall"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var tallyCmd = &cobra.Command{
	Use:   "tally",
	Short: "List all daemons started by Lilith, with invocation time",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Italic.TextStyle(chalk.Blue.Color("lilith")) + ` lists all daemons you have invoked, shows their group, PID, time they were started, and whether they’re still running`,
	Example: chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("tally")),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.tally"

		// 1) Read the daemon directory
		dir := getDaemonDir()
		entries, err := domovoi.ReadDir(dir, verbose)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("reading daemon directory"))

		// 2) Print header
		fmt.Printf(
			"%-20s %-15s %-6s %-20s %s\n",
			"NAME", "GROUP", "PID", "INVOKED", "STATUS",
		)

		// 3) Iterate over metadata files
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

			// 4) Load metadata
			meta, err := loadMeta(name)
			if err != nil {
				// skip entries that fail to parse
				horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)))
				continue
			}

			// 5) Check process status
			status := chalk.Red.Color("stopped")
			if p, err := os.FindProcess(meta.PID); err == nil {
				if err = p.Signal(syscall.Signal(0)); err == nil {
					status = chalk.Green.Color("running")
				}
			}

			// 6) Format invoked timestamp
			invoked := meta.InvokedAt.Format("2006-01-02 15:04:05")

			// 7) Print row
			fmt.Printf(
				"%-20s %-15s %-6d %-20s %s\n",
				meta.Name, meta.Group, meta.PID, invoked, status,
			)
		}
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(tallyCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
