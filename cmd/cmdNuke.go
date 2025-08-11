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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var nukeCmd = &cobra.Command{
	Use:     "nuke",
	Short:   "",
	Long:    helpNuke,
	Example: exampleNuke,

	Run: runNuke,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var ()

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(nukeCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpNuke = chalk.Bold.TextStyle(chalk.Green.Color("Daniel Rivas ")) +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) +
	chalk.Dim.TextStyle(chalk.Cyan.Color("\n\n"))

var exampleNuke = chalk.White.Color("") + " " +
	chalk.Bold.TextStyle(chalk.White.Color("nuke")) + " "

////////////////////////////////////////////////////////////////////////////////////////////////////

func runNuke(cmd *cobra.Command, args []string) {
	const op = "lilith.nuke"

	daemonDir := filepath.Join(home, ".lilith", "daemon")
	daemonFiles, err := domovoi.ReadDir(daemonDir, verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("Error reading daemon directory"))

	for _, daemon := range daemonFiles {
		fmt.Println(daemon)

		if daemon.IsDir() || !strings.HasSuffix(daemon.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(daemonDir, daemon.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", filePath, err)
			continue
		}

		var data struct {
			PID int `json:"pid"`
		}
		if err := json.Unmarshal(content, &data); err != nil {
			fmt.Printf("Invalid JSON in %s: %v\n", filePath, err)
			continue
		}
		fmt.Println(data)

		proc, err := os.FindProcess(data.PID)
		if err == nil {
			if err := proc.Kill(); err == nil && verbose {
				fmt.Printf("Killed active process (PID %d)\n", data.PID)
			}
		}
		os.Remove(filepath.Join(daemonDir, daemon.Name()))
		// domovoi.RemoveFile(filepath.Join(daemonDir, daemon.Name()), verbose)

	}

	logDir := filepath.Join(home, ".lilith", "logs")
	logEntries, err := domovoi.ReadDir(logDir, verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("Error reading  directory"))

	for _, logEntry := range logEntries {
		fmt.Println(filepath.Join(logDir, logEntry.Name()))
		fmt.Println(logEntry.Name())

		os.Remove(filepath.Join(logDir, logEntry.Name()))
		// domovoi.RemoveFile(filepath.Join(logDir, logEntry.Name()), verbose)
	}

	if verbose {
		fmt.Println("Nuke completed.")
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
