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
	"strings"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// invocation flags + derived defaults
var (
	daemonName string
	watchDir   string
	scriptPath string
	logName    string
	groupName  string
	configName string
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// invokeCmd starts a new watcher daemon, pulling defaults from ~/.lilith/forge.toml
var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Start a new watcher daemon",

	////////////////////////////////////////////////////////////////////////////////////////////////////

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// 1) Load ~/.lilith/forge.{toml,yaml,json}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		v := viper.New()
		v.AddConfigPath(filepath.Join(home, ".lilith"))
		v.SetConfigName(configName)
		if err := v.ReadInConfig(); err != nil {
			if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
				return err
			}
		}

		// 2) Default group ← the config filename
		if used := v.ConfigFileUsed(); used != "" {
			base := filepath.Base(used) // e.g. "forge.toml"
			groupName = strings.TrimSuffix(base, filepath.Ext(base))
		} else {
			groupName = configName
		}

		// 3) Bind top–level keys from forge.toml if flags unset
		bindFlag(cmd, "group", &groupName, v)
		bindFlag(cmd, "watch", &watchDir, v)
		bindFlag(cmd, "script", &scriptPath, v)
		bindFlag(cmd, "log", &logName, v)

		// 4) Per–workflow overrides under [workflows.<daemonName>]
		if daemonName != "" && v.IsSet("workflows."+daemonName) {
			wf := v.Sub("workflows." + daemonName)
			if wf == nil {
				return fmt.Errorf("workflow %q is not defined", daemonName)
			}
			bindFlag(cmd, "group", &groupName, wf)
			bindFlag(cmd, "watch", &watchDir, wf)
			bindFlag(cmd, "script", &scriptPath, wf)
			bindFlag(cmd, "log", &logName, wf)
		}

		return nil
	},

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.invoke"

		// 5) Validate required inputs
		if daemonName == "" {
			horus.CheckErr(
				fmt.Errorf("`--name` is required"),
				horus.WithOp(op), horus.WithMessage("provide a unique daemon name"),
			)
		}
		if watchDir == "" {
			horus.CheckErr(
				fmt.Errorf("`--watch` is required"),
				horus.WithOp(op), horus.WithMessage("provide a directory to watch"),
			)
		}
		if scriptPath == "" {
			horus.CheckErr(
				fmt.Errorf("`--script` is required"),
				horus.WithOp(op), horus.WithMessage("provide a script to run"),
			)
		}
		if logName == "" {
			horus.CheckErr(
				fmt.Errorf("`--log` is required"),
				horus.WithOp(op), horus.WithMessage("provide a log name"),
			)
		}

		// 6) Expand ~ and $ENV
		watchDir = os.ExpandEnv(watchDir)
		scriptPath = os.ExpandEnv(scriptPath)

		// 7) Ensure ~/.lilith/logs exists
		home, err := os.UserHomeDir()
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("getting home directory"))

		logDir := filepath.Join(home, ".lilith", "logs")
		horus.CheckErr(
			domovoi.CreateDir(logDir),
			horus.WithOp(op), horus.WithMessage(fmt.Sprintf("creating %q", logDir)),
		)

		logPath := filepath.Join(logDir, logName+".log")

		// 8) Build & persist metadata
		meta := &daemonMeta{
			Name:       daemonName,
			Group:      groupName,
			WatchDir:   watchDir,
			ScriptPath: scriptPath,
			LogPath:    logPath,
			InvokedAt:  time.Now(),
		}

		pid, err := spawnWatcher(meta)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("starting watcher"))
		meta.PID = pid

		horus.CheckErr(saveMeta(meta), horus.WithOp(op), horus.WithMessage("writing metadata"))

		// 9) Final message
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
	invokeCmd.Flags().StringVarP(&groupName, "group", "g", "", "Watcher group name (overrides config default)")
	invokeCmd.Flags().StringVarP(&watchDir, "watch", "w", "", "Directory to watch")
	invokeCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "Script to execute on change")
	invokeCmd.Flags().StringVarP(&logName, "log", "l", "", "Name for log file (no `.log` extension)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
