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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var tallyCmd = &cobra.Command{
	Use:     "tally",
	Short:   "List active daemons",
	Long:    helpTally,
	Example: exampleTally,

	Run: RunTally,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(tallyCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpTally = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"List all daemons invoked, showing group, PID, start time, and current status",
)

var exampleTally = formatExample(
	"lilith",
	[]string{"tally"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func RunTally(cmd *cobra.Command, args []string) {
	const op = "lilith.tally"

	// 1) Read the daemon directory
	dir := GetDaemonDir()
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

		// 5) Determine process status via `ps` (detect T=stopped/paused)
		status := chalk.Red.Color("dead")
		stateOut, err := exec.Command("ps", "-o", "state=", "-p", strconv.Itoa(meta.PID)).Output()
		if err == nil {
			state := strings.TrimSpace(string(stateOut))
			switch {
			case strings.HasPrefix(state, "T"):
				status = chalk.Yellow.Color("limbo")
			default:
				status = chalk.Green.Color("alive")
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
}

////////////////////////////////////////////////////////////////////////////////////////////////////
