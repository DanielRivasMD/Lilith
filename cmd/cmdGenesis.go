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

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var genesisFlags struct {
	output string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func GenesisCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("genesis", runGenesis))

	cmd.Flags().StringVarP(&genesisFlags.output, "output", "o", "", "Path to write example config (default = stdout)")

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runGenesis(cmd *cobra.Command, args []string) {
	createSubdirs(dirs, rootFlags.verbose)
	generateConfig(generateToml())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func createSubdirs(d configDirs, verbose bool) {
	const op = "genesis.createSubDirs"

	toCreate := []struct {
		label, path string
	}{
		{"lilith root", d.lilith},
		{"config", d.config},
		{"logs", d.log},
		{"daemon", d.daemon},
	}

	for _, dir := range toCreate {
		horus.CheckErr(
			domovoi.CreateDir(dir.path, verbose),
			horus.WithOp(op),
			horus.WithCategory("io_error"),
			horus.WithMessage(fmt.Sprintf("creating %s directory", dir.label)),
			horus.WithDetails(map[string]any{
				"path": dir.path,
			}),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func generateToml() string {
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

func generateConfig(example string) {
	op := "genesis.generateConfig"

	if genesisFlags.output == "" {
		fmt.Print(example)
		return
	}

	horus.CheckErr(
		os.WriteFile(genesisFlags.output, []byte(example), 0o644),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage(fmt.Sprintf("writing example to %q", genesisFlags.output)),
	)

	fmt.Printf("Example config written to %s\n", genesisFlags.output)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
