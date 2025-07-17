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
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	daemonName string
	watchDir   string
	scriptPath string
	groupName  string
	logName    string
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Start a new watcher daemon",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if daemonName != "" && viper.IsSet("workflows."+daemonName) {
			wf := viper.Sub("workflows." + daemonName)
			if wf == nil {
				return fmt.Errorf("workflow %q is not a table in config", daemonName)
			}
			bindFlag(cmd, "watch", &watchDir, wf)
			bindFlag(cmd, "script", &scriptPath, wf)
			bindFlag(cmd, "group", &groupName, wf)
			bindFlag(cmd, "log", &logName, wf)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.invoke"

		// 1) Validate required flags before using horus
		if daemonName == "" {
			horus.CheckErr(
				fmt.Errorf("`--name` is required"),
				horus.WithOp(op),
				horus.WithMessage("you must specify a unique daemon name"),
			)
		}
		if watchDir == "" {
			horus.CheckErr(
				fmt.Errorf("`--watch` is required"),
				horus.WithOp(op),
				horus.WithMessage("you must specify a directory to watch"),
			)
		}
		if scriptPath == "" {
			horus.CheckErr(
				fmt.Errorf("`--script` is required"),
				horus.WithOp(op),
				horus.WithMessage("you must specify a script to run"),
			)
		}

		// 2) Expand environment variables and tildes
		watchDir = os.ExpandEnv(watchDir)
		scriptPath = os.ExpandEnv(scriptPath)

		// 3) Prepare log directory under ~/.lilith/logs/
		home, err := os.UserHomeDir()
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("fetching home directory"))

		logDir := filepath.Join(home, ".lilith", "logs")
		horus.CheckErr(
			domovoi.CreateDir(logDir),
			horus.WithOp(op),
			horus.WithMessage(fmt.Sprintf("creating log directory %q", logDir)),
		)

		// 4) Construct full log path
		logPath := filepath.Join(logDir, logName+".log")

		// 5) Build and persist metadata
		meta := &daemonMeta{
			Name:       daemonName,
			Group:      groupName,
			WatchDir:   watchDir,
			ScriptPath: scriptPath,
			LogPath:    logPath,
			InvokedAt:  time.Now(),
		}

		// 6) Spawn the watcher process
		pid, err := spawnWatcher(meta)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("starting watcher"))
		meta.PID = pid

		// 7) Save metadata to disk
		horus.CheckErr(saveMeta(meta), horus.WithOp(op), horus.WithMessage("writing metadata"))

		// 8) Final success message
		fmt.Printf(
			"%s invoked daemon %q (group=%q) with PID %d\n",
			chalk.Green.Color("OK:"), daemonName, groupName, pid,
		)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&daemonName, "name", "n", "", "Unique daemon name")
	invokeCmd.Flags().StringVarP(&watchDir, "watch", "w", viper.GetString("watch"), "Directory to watch")
	invokeCmd.Flags().StringVarP(&scriptPath, "script", "s", viper.GetString("script"), "Script to execute on change")
	invokeCmd.Flags().StringVarP(&groupName, "group", "g", viper.GetString("group"), "Watcher group name")
	invokeCmd.Flags().StringVarP(&logName, "log", "l", viper.GetString("log"), "Name for log file (without extension)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
