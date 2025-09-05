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
	"strings"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var dumpconfigCmd = &cobra.Command{
	Use:     "dump-config",
	Short:   "Generate an example TOML config file",
	Long:    helpDumpConfig,
	Example: exampleDumpConfig,

	Run: runDumpConfig,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	dumpOutput string
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(dumpconfigCmd)

	dumpconfigCmd.Flags().StringVarP(&dumpOutput, "output", "o", "", "Path to write example config (default = stdout)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runDumpConfig(cmd *cobra.Command, args []string) {
	const op = "lilith.dumpConfig"

	example := generateExampleConfig()

	if dumpOutput == "" {
		fmt.Print(example)
		return
	}

	horus.CheckErr(
		os.WriteFile(dumpOutput, []byte(example), 0o644),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage(fmt.Sprintf("writing example to %q", dumpOutput)),
	)

	fmt.Printf("Example config written to %s\n", dumpOutput)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func generateExampleConfig() string {
	lines := []string{
		"# Lilith configuration file",
		"# Define workflows under the [workflows.<name>] table.",
		"",
		"# Minimal example workflow named \"demo\":",
		"[workflows.demo]",
		"watch  = \"/path/to/your/project\"   # directory to watch",
		"script = \"build.sh\"                # script to run on changes",
		"",
		"# Optional settings:",
		"# daemon = \"demo-daemon\"            # unique watcher name",
		"# group  = \"default\"                # watcher group",
		"# log    = \"demo\"                   # base name for the .log file",
		"",
		"# Add more workflows simply by adding new blocks:",
		"# [workflows.other]",
		"# watch  = \"/another/path\"",
		"# script = \"deploy.sh\"",
		"# daemon = \"other-daemon\"",
		"# group  = \"deploy-group\"",
		"# log    = \"other\"",
		"",
		"# Save this snippet as ~/.lilith/config/example.toml",
	}
	return strings.Join(lines, "\n") + "\n"
}

////////////////////////////////////////////////////////////////////////////////////////////////////
