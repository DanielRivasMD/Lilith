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
	const op = "lilith.genesis"
	createSubdirs(configDirs, rootFlags.verbose, op)
	if genesisFlags.output != "" {
		generateConfig(generateToml())
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
