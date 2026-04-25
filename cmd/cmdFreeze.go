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

var freezeFlags struct {
	all   bool
	group string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func FreezeCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("freeze", runFreeze,
		domovoi.WithValidArgsFunction(completeDaemonNames),
	))

	cmd.Flags().StringVarP(&freezeFlags.group, "group", "", "", "Freeze all daemons belonging to a specific group")
	cmd.Flags().BoolVarP(&freezeFlags.all, "all", "", false, "Freeze all running daemons")

	horus.CheckErr(cmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("freeze.init"), horus.WithMessage("registering config completion"))

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runFreeze(cmd *cobra.Command, args []string) {
	const op = "lilith.freeze"

	switch {
	case freezeFlags.all:
		freezeAllDaemons()
	case freezeFlags.group != "":
		freezeGroupDaemons(freezeFlags.group)
	case len(args) == 1:
		freezeDaemon(args[0])
	default:
		horus.CheckErr(
			errors.New(""),
			horus.WithOp(op),
			horus.WithMessage("daemon / flag"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string {
				return "missing " + horus.OneLineErr(he.Message)
			}),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
