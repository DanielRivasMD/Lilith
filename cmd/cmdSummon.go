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
	"os"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var summonFlags struct {
	follow bool
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func SummonCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("summon", runSummon,
		domovoi.WithArgs(cobra.ExactArgs(1)),
	))

	cmd.Flags().BoolVarP(&summonFlags.follow, "follow", "f", false, "Continuously watch the log file")

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSummon(cmd *cobra.Command, args []string) {
	const op = "lilith.summon"
	name := args[0]

	meta := loadMeta(name)

	if summonFlags.follow {
		horus.CheckErr(
			domovoi.ExecCmd("tail", "-f", meta.LogPath),
			horus.WithOp(op),
			horus.WithMessage("streaming log"),
		)
	} else {
		pager := os.Getenv("PAGER")
		if pager == "" {
			pager = "less"
		}
		horus.CheckErr(
			domovoi.ExecCmd(pager, "--paging", "always", meta.LogPath),
			horus.WithOp(op),
			horus.WithMessage("paging log"),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
