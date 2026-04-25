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

func TallyCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	return horus.Must(d.MakeCmd("tally", runTally))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTally(cmd *cobra.Command, args []string) {
	const op = "lilith.tally"

	tallyBorder := strings.Repeat("=", 80)
	tallyInter := strings.Repeat("-", 80)
	colWidths := []int{20, 15, 10, 25, 10}

	entries, err := domovoi.ReadDir(configDirs.daemon, rootFlags.verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("reading daemon directory"))

	fmt.Println(tallyBorder)

	headers := []string{
		chalk.Cyan.Color("NAME"),
		chalk.Cyan.Color("GROUP"),
		chalk.Cyan.Color("PID"),
		chalk.Cyan.Color("INVOKED"),
		chalk.Cyan.Color("STATUS"),
	}
	var headerRow string
	for i, h := range headers {
		headerRow += padRight(h, colWidths[i])
	}
	fmt.Println(headerRow)

	fmt.Println(tallyInter)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		meta := loadMeta(name)

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

		invoked := meta.InvokedAt.Format("2006-01-02 15:04:05")

		row := padRight(meta.Daemon, colWidths[0]) +
			padRight(meta.Group, colWidths[1]) +
			padRight(strconv.Itoa(meta.PID), colWidths[2]) +
			padRight(invoked, colWidths[3]) +
			padRight(status, colWidths[4])
		fmt.Println(row)
	}
	fmt.Println(tallyBorder)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
