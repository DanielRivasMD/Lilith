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

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var installCmd = &cobra.Command{
	Use:     "install",
	Hidden:  true,
	Short:   "",
	Long:    helpInstall,
	Example: exampleInstall,

	ValidArgs: []string{"full", "config", "dirs"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),

	Run: runInstall,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var ()

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(installCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpInstall = formatHelp(
	"Daniel Rivas ",
	"<danielrivasmd@gmail.com>",
	"",
)

var exampleInstall = formatExample(
	"lilith",
	[]string{"install"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func runInstall(cmd *cobra.Command, args []string) {

	switch args[0] {
	case "full":
		createConfig(verbose)
		createDirs(verbose)
	case "config":
		createConfig(verbose)
	case "dirs":
		createDirs(verbose)
	default:
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func createConfig(verbose bool) {
	const op = "install.create_config"

	// Determine full path: ~/.lilith/config/config.toml
	home, err := domovoi.FindHome(verbose)
	horus.CheckErr(err,
		horus.WithOp(op),
		horus.WithMessage("locating $HOME"),
	)
	dir := filepath.Join(home, ".lilith", "config")
	path := filepath.Join(dir, "config.toml")
	horus.CheckErr(
		domovoi.CreateDir(dir, verbose),
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("creating directory %q", dir)),
	)

	// 1) Create the file if it doesn't exist
	createAction := domovoi.CreateFile(path, verbose)
	created, err := createAction(path)
	horus.CheckErr(err,
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("ensuring file %q exists", path)),
	)

	// 2) If newly created, write the default TOML
	if created {
		if verbose {
			fmt.Printf("Writing default configuration into %s\n", path)
		}
		defaultToml := `[workflows.dummy]
watch = "~/Downloads"
script = "echo 'downloaded'"` + "\n"
		if writeErr := os.WriteFile(path, []byte(defaultToml), 0644); writeErr != nil {
			horus.CheckErr(writeErr,
				horus.WithOp(op),
				horus.WithMessage(fmt.Sprintf("writing default config to %q", path)),
			)
		}
	} else if verbose {
		fmt.Printf("Configuration file already present: %s\n", path)
	}
}

func createDirs(verbose bool) {
	const op = "install.create_dirs"

	// 1) Find home
	home, err := domovoi.FindHome(verbose)
	horus.CheckErr(err,
		horus.WithOp(op),
		horus.WithMessage("locating $HOME"),
	)

	// 2) Create ~/.lilith and its subdirs
	base := filepath.Join(home, ".lilith")
	for _, sub := range []string{"config", "daemon", "logs"} {
		dir := filepath.Join(base, sub)
		horus.CheckErr(
			domovoi.CreateDir(dir, verbose),
			horus.WithOp(op),
			horus.WithMessage(fmt.Sprintf("creating directory %q", dir)),
		)
	}

	if verbose {
		fmt.Println(chalk.Green.Color("OK:") + " initialized ~/.lilith directories")
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
