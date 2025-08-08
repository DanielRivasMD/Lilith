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
	"syscall"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	slayAll   bool
	slayGroup string
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var slayCmd = &cobra.Command{
	Use:   "slay " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short: "Stop & clean up one or more daemons",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` +
		chalk.Blue.Color(chalk.Italic.TextStyle("lilith")) + ` gracefully stop one or many alive daemons, remove their metadata files and logs, allowing fresh invocation later
`,
	Example: chalk.White.Color("lilith") + ` ` + chalk.Bold.TextStyle(chalk.White.Color("slay")) + ` ` +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")) + `
` + chalk.White.Color("lilith") + ` ` + chalk.Bold.TextStyle(chalk.White.Color("slay")) + ` ` +
		chalk.Italic.TextStyle("--all") + `
` + chalk.White.Color("lilith") + ` ` + chalk.Bold.TextStyle(chalk.White.Color("slay")) + ` ` +
		chalk.Italic.TextStyle("--group") + ` ` + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<editors>")),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeDaemonNames,

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: runSlay,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	slayCmd.Flags().BoolVar(&slayAll, "all", false, "Slay all daemons")
	slayCmd.Flags().StringVar(&slayGroup, "group", "", "Slay all daemons in a specific group")
	rootCmd.AddCommand(slayCmd)
}

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
		horus.CheckErr(
			horus.NewCategorizedHerror(op, "validation", "must provide a daemon name or --all / --group", nil, nil),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func slaySingleDaemon(name string) {

	const op = "lilith.slay"

	// 1) Load metadata
	meta, err := loadMeta(name)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)))

	// 2) Signal the process to terminate
	proc, err := os.FindProcess(meta.PID)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("finding process"))
	horus.CheckErr(proc.Signal(syscall.SIGTERM), horus.WithOp(op), horus.WithMessage("sending SIGTERM"))

	// 3) Remove the metadata JSON file
	metaFile := filepath.Join(getDaemonDir(), name+".json")
	_, err = domovoi.RemoveFile(metaFile, verbose)(metaFile)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("removing metadata file"))

	// 4) Remove the log file
	_, err = domovoi.RemoveFile(meta.LogPath, verbose)(meta.LogPath)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("removing log file"))

	// 5) Final confirmation
	fmt.Printf("%s slayed daemon %q\n", chalk.Green.Color("OK:"), name)
}

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

func mustListDaemonMetaFiles() []string {
	dir := getDaemonDir()
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	horus.CheckErr(err, horus.WithOp("daemon.list"))
	return matches
}

func nameFrom(path string) string {
	return filepath.Base(path[:len(path)-len(".json")])
}

func matchesGroup(metaPath, expectedGroup string) bool {
	// Try to load JSON metadata
	data, err := os.ReadFile(metaPath)
	if err != nil {
		// optionally log or ignore
		return false
	}

	var meta struct {
		Group string `json:"group"`
	}

	if err := json.Unmarshal(data, &meta); err != nil {
		// if unmarshal fails, ignore this file
		return false
	}

	return meta.Group == expectedGroup
}

////////////////////////////////////////////////////////////////////////////////////////////////////
