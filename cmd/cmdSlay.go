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
	"errors"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var slayFlags struct {
	all   bool
	group string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func SlayCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("slay", runSlay,
		domovoi.WithValidArgsFunction(completeDaemonNames),
	))

	cmd.Flags().BoolVarP(&slayFlags.all, "all", "", false, "Slay all daemons")
	cmd.Flags().StringVarP(&slayFlags.group, "group", "", "", "Slay all daemons in a specific group")

	horus.CheckErr(cmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("slay.init"), horus.WithMessage("registering config completion"))

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSlay(cmd *cobra.Command, args []string) {
	const op = "lilith.slay"

	switch {
	case slayFlags.all:
		slayAllDaemons()
	case slayFlags.group != "":
		slayGroupDaemons(slayFlags.group)
	case len(args) == 1:
		slayDaemon(args[0])
	default:
		horus.CheckErr(
			errors.New("missing: "),
			horus.WithMessage("daemon / flag"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string {
				return horus.OneLineErr(he.Message)
			}),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
