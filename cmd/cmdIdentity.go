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

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// Global declarations
var ()

////////////////////////////////////////////////////////////////////////////////////////////////////

var identityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Reveal Lilith’s mythic origin",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Italic.TextStyle(chalk.Blue.Color("lilith")) + ` identity summons the first woman of legend—Lilith. Across millennia she has stood as a symbol of independence, mystery, and the night’s whisper.`,
	Example: chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("identity")),

	////////////////////////////////////////////////////////////////////////////////////////////////////
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(chalk.Magenta.Color("Lilith:"))
		fmt.Println("  • In ancient Mesopotamia she appears as a wind spirit, free and untamed.")
		fmt.Println("  • In early Jewish lore she is Adam’s first wife, who walked away rather than submit.")
		fmt.Println("  • Through medieval tales she became a demon of the night—yet also a feminist icon.")
		fmt.Println("  • Today she embodies autonomy, the power of the untold story, and the strength of dusk.")
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(identityCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
