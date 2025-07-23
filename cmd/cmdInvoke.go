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

var (
	daemonName string // instance name, defaults to configName
	configName string // workflow key
	watchDir   string
	scriptPath string
	logName    string
	groupName  string // derived from TOML filename
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Start a new watcher daemon",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Italic.TextStyle(chalk.Blue.Color("lilith")) + ` read a named workflow from ~/.lilith/config/*.toml files, spawn a background watcher process for the specified directory, and execute the configured script on each change. Metadata is persisted for inspection or summoning the daemon`,
	Example: chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("invoke")) + " " +
		chalk.Italic.TextStyle(chalk.White.Color("--config")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")) + "\n" +
		chalk.White.Color("lilith") + " " +
		chalk.Bold.TextStyle(chalk.White.Color("invoke")) + " " +
		chalk.Italic.TextStyle(chalk.White.Color("--name")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")) + " " +
		chalk.Italic.TextStyle(chalk.White.Color("--watch")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("~/src/helix")) + " " +
		chalk.Italic.TextStyle(chalk.White.Color("--script")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix.sh")) + " " +
		chalk.Italic.TextStyle(chalk.White.Color("--log")) + " " +
		chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix")),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	PreRunE: func(cmd *cobra.Command, args []string) error {
		const op = "lilith.invoke.pre"

		// 1) Find home and read config dir
		home, err := domovoi.FindHome(verbose)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("getting home directory"))
		cfgDir := filepath.Join(home, ".lilith", "config")

		// 2) Load matching TOML
		var (
			foundV      *viper.Viper
			cfgFileUsed string
		)
		fis, err := domovoi.ReadDir(cfgDir, verbose)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("reading config dir"))

		for _, fi := range fis {
			if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".toml") {
				continue
			}
			path := filepath.Join(cfgDir, fi.Name())
			v := viper.New()
			v.SetConfigFile(path)
			if err := v.ReadInConfig(); err != nil {
				continue
			}
			if v.IsSet("workflows." + configName) {
				foundV = v
				cfgFileUsed = path
				break
			}
		}

		// TODO: let `horus` watch over this error
		if foundV == nil {
			return fmt.Errorf("workflow %q not found in %s/*.toml", configName, cfgDir)
		}

		// 3) Default daemonName ← configName if none provided
		if daemonName == "" {
			daemonName = configName
			if err := cmd.Flags().Set("name", daemonName); err != nil {
				horus.CheckErr(
					err,
					horus.WithOp(op),
					horus.WithMessage("setting default --name from config"),
				)
			}
		}

		// 4) Derive groupName from TOML filename
		base := filepath.Base(cfgFileUsed)                       // e.g. "forge.toml"
		groupName = strings.TrimSuffix(base, filepath.Ext(base)) // e.g. "forge"
		if err := cmd.Flags().Set("group", groupName); err != nil {
			horus.CheckErr(
				err,
				horus.WithOp(op),
				horus.WithMessage("setting default --group from TOML filename"),
			)
		}

		// 5) Bind watch & script flags
		wf := foundV.Sub("workflows." + configName)
		bindFlag(cmd, "watch", &watchDir, wf)
		bindFlag(cmd, "script", &scriptPath, wf)

		// 6) Auto‐assign logName from workflow key
		if !cmd.Flags().Changed("log") {
			logName = configName
			if err := cmd.Flags().Set("log", logName); err != nil {
				horus.CheckErr(
					err,
					horus.WithOp(op),
					horus.WithMessage("setting default --log from workflow key"),
				)
			}
		}

		return nil
	},

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.invoke"

		// 7) Validate required flags
		horus.CheckEmpty(
			watchDir,
			"`--watch` is required",
			horus.WithOp(op),
			horus.WithMessage("provide a directory to watch"),
		)
		horus.CheckEmpty(
			scriptPath,
			"`--script` is required",
			horus.WithOp(op),
			horus.WithMessage("provide a script to run"),
		)
		horus.CheckEmpty(
			logName,
			"`--log` is required",
			horus.WithOp(op),
			horus.WithMessage("provide a log name"),
		)

		// 8) Expand env vars / tilde
		horus.CheckErr(
			func() error {
				var err error
				watchDir, err = expandPath(watchDir)
				return err
			}(),
			horus.WithOp(op),
			horus.WithMessage("expanding watch path"),
		)
		horus.CheckErr(
			func() error {
				var err error
				scriptPath, err = expandPath(scriptPath)
				return err
			}(),
			horus.WithOp(op),
			horus.WithMessage("expanding script path"),
		)

		// 9) Ensure ~/.lilith/logs exists
		home, err := domovoi.FindHome(verbose)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("getting home directory"))
		logDir := filepath.Join(home, ".lilith", "logs")
		horus.CheckErr(
			domovoi.CreateDir(logDir, verbose),
			horus.WithOp(op),
			horus.WithMessage(fmt.Sprintf("creating %q", logDir)),
		)
		logPath := filepath.Join(logDir, logName+".log")

		// 10) Build & persist metadata
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

		// 11) Done
		fmt.Printf(
			"%s invoked daemon %q (group=%q) with PID %d\n",
			chalk.Green.Color("OK:"),
			daemonName,
			groupName,
			pid,
		)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&daemonName, "name", "n", "", "Unique daemon name (defaults to --config)")
	invokeCmd.Flags().StringVarP(&configName, "config", "c", "", "Workflow to apply")
	invokeCmd.Flags().StringVarP(&groupName, "group", "g", "", "Watcher group name (overrides TOML)")
	invokeCmd.Flags().StringVarP(&watchDir, "watch", "w", "", "Directory to watch")
	invokeCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "Script to execute on change")
	invokeCmd.Flags().StringVarP(&logName, "log", "l", "", "Name for log file (no `.log` extension)")

	if err := invokeCmd.RegisterFlagCompletionFunc("config", completeWorkflowNames); err != nil {
		horus.CheckErr(err, horus.WithOp("invoke.init"), horus.WithMessage("registering config completion"))
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
