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
	"syscall"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var freezeCmd = &cobra.Command{
	Use:   "freeze " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("[daemon]")),
	Short: "Pause a daemon",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Italic.TextStyle(chalk.Blue.Color("lilith")) + ` send SIGSTOP to an alive daemon, pausing its execution until you explicitly resume it via OS tools.`,
	Example: chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("freeze")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeDaemonNames,

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.freeze"

		group, _ := cmd.Flags().GetString("group")
		all, _ := cmd.Flags().GetBool("all")

		switch {
		case all:
			freezeAllDaemons()
			return
		case group != "":
			freezeGroupDaemons(group)
			return
		default:
			// Single daemon freeze
			name := args[0]

			// 1) Load metadata
			meta, err := loadMeta(name)
			horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("loading metadata for %q", name)))

			// 2) Find and pause process
			proc, err := os.FindProcess(meta.PID)
			horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("finding process"))
			horus.CheckErr(proc.Signal(syscall.SIGSTOP), horus.WithOp(op), horus.WithMessage("sending SIGSTOP"))

			// 3) Confirmation
			fmt.Printf("%s froze daemon %q\n", chalk.Green.Color("OK:"), name)
		}
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(freezeCmd)

	freezeCmd.Flags().String("group", "", "Freeze all daemons belonging to a specific group")
	freezeCmd.Flags().Bool("all", false, "Freeze all running daemons")

	_ = freezeCmd.RegisterFlagCompletionFunc("group", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return availableGroups(), cobra.ShellCompDirectiveDefault
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func freezeGroupDaemons(group string) {
	files := mustListDaemonMetaFiles()
	for _, path := range files {
		if matchesGroup(path, group) {
			meta := mustLoadMeta(path)
			_ = sendSignal(meta.PID, syscall.SIGSTOP)
			fmt.Printf("%s froze daemon %q\n", chalk.Green.Color("OK:"), meta.Name)
		}
	}
}

func freezeAllDaemons() {
	files := mustListDaemonMetaFiles()
	for _, path := range files {
		meta := mustLoadMeta(path)
		_ = sendSignal(meta.PID, syscall.SIGSTOP)
		fmt.Printf("%s froze daemon %q\n", chalk.Green.Color("OK:"), meta.Name)
	}
}

func mustLoadMeta(path string) daemonMeta {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading metadata from %s: %v\n", path, err)
		os.Exit(1)
	}

	var meta daemonMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON in %s: %v\n", path, err)
		os.Exit(1)
	}

	return meta
}

func sendSignal(pid int, sig syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("could not find process %d: %w", pid, err)
	}
	return proc.Signal(sig)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
