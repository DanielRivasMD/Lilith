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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var invokeCmd = &cobra.Command{
	Use:     "invoke",
	Short:   "Start a new watcher daemon",
	Long:    helpInvoke,
	Example: exampleInvoke,

	PreRunE: preInvoke,
	Run:     runInvoke,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	configName string // workflow key
	daemonName string // instance name, defaults to configName
	watchDir   string
	scriptPath string
	logName    string
	groupName  string // derived from TOML filename
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&configName, "config", "c", "", "Workflow to apply")
	invokeCmd.Flags().StringVarP(&daemonName, "name", "n", "", "Unique daemon name (defaults to --config)")
	invokeCmd.Flags().StringVarP(&groupName, "group", "g", "", "Watcher group name (overrides TOML)")
	invokeCmd.Flags().StringVarP(&watchDir, "watch", "w", "", "Directory to watch")
	invokeCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "Script to execute on change")
	invokeCmd.Flags().StringVarP(&logName, "log", "l", "", "Name for log file (no `.log` extension)")

	horus.CheckErr(invokeCmd.RegisterFlagCompletionFunc("config", completeWorkflowNames), horus.WithOp("invoke.init"), horus.WithMessage("registering config completion"))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpInvoke = chalk.Bold.TextStyle(chalk.Green.Color("Daniel Rivas ")) +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) +
	chalk.Dim.TextStyle(chalk.Cyan.Color("\n\nspawn daemon process for the specified directory & execute the configured script on change"+
		"\nmetadata is persistent for summoning the daemon"))

var exampleInvoke = chalk.White.Color("lilith") + " " +
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
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("helix"))

////////////////////////////////////////////////////////////////////////////////////////////////////

func preInvoke(cmd *cobra.Command, args []string) error {
	const op = "lilith.invoke.pre"

	// 1) Find home and read config dir
	home, err := domovoi.FindHome(verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("getting home directory"))
	cfgDir := filepath.Join(home, ".lilith", "config")

	// 2) Load matching TOML
	var (
		foundV      *viper.Viper
		cfgFileUsed string
	)
	fis, err := domovoi.ReadDir(cfgDir, verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("reading config dir"))

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

	// 3) Handle missing workflow with horus
	if foundV == nil {
		horus.CheckErr(
			fmt.Errorf("workflow %q not found in %s/*.toml", configName, cfgDir),
			horus.WithOp(op),
			horus.WithMessage("could not find named workflow in config directory"),
			horus.WithCategory("config_error"),
		)
	}

	// 4) Default daemonName ← configName if none provided
	if daemonName == "" {
		daemonName = configName
		if err := cmd.Flags().Set("name", daemonName); err != nil {
			horus.CheckErr(
				err,
				horus.WithOp(op),
				horus.WithMessage("setting default --name from config"),
				horus.WithCategory("config_error"),
			)
		}
	}

	// 5) Derive groupName from TOML filename
	base := filepath.Base(cfgFileUsed)                       // e.g. "forge.toml"
	groupName = strings.TrimSuffix(base, filepath.Ext(base)) // e.g. "forge"
	if err := cmd.Flags().Set("group", groupName); err != nil {
		horus.CheckErr(
			err,
			horus.WithOp(op),
			horus.WithMessage("setting default --group from TOML filename"),
			horus.WithCategory("config_error"),
		)
	}

	// 6) Bind watch & script flags
	wf := foundV.Sub("workflows." + configName)
	bindFlag(cmd, "watch", &watchDir, wf)
	bindFlag(cmd, "script", &scriptPath, wf)

	// 7) Auto‐assign logName from workflow key
	if !cmd.Flags().Changed("log") {
		logName = configName
		if err := cmd.Flags().Set("log", logName); err != nil {
			horus.CheckErr(
				err,
				horus.WithOp(op),
				horus.WithMessage("setting default --log from workflow key"),
				horus.WithCategory("config_error"),
			)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke"

	// 7) Validate required flags
	horus.CheckEmpty(
		watchDir,
		"`--watch` is required",
		horus.WithOp(op),
		horus.WithMessage("provide a directory to watch"),
		horus.WithCategory("spawn_error"),
	)
	horus.CheckEmpty(
		scriptPath,
		"`--script` is required",
		horus.WithOp(op),
		horus.WithMessage("provide a script to run"),
		horus.WithCategory("spawn_error"),
	)
	horus.CheckEmpty(
		logName,
		"`--log` is required",
		horus.WithOp(op),
		horus.WithMessage("provide a log name"),
		horus.WithCategory("spawn_error"),
	)

	// 8) Expand env vars / tilde
	watchDir = mustExpand(watchDir, "--watch")
	scriptPath = mustExpand(scriptPath, "--script")

	// 9) Ensure ~/.lilith/logs exists
	home, err := domovoi.FindHome(verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("getting home directory"))
	logDir := filepath.Join(home, ".lilith", "logs")
	horus.CheckErr(
		domovoi.CreateDir(logDir, verbose),
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("creating %q", logDir)),
		horus.WithCategory("env_error"),
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

	for _, path := range mustListDaemonMetaFiles() {
		existing := mustLoadMeta(path)
		if existing.WatchDir == watchDir && isDaemonActive(&existing) {
			exitAlreadyRunning(existing.Name)
			// unreachable
		}
	}

	pid, err := spawnWatcher(meta)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("starting watcher"))
	meta.PID = pid
	horus.CheckErr(saveMeta(meta), horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("writing metadata"))

	// 11) Done
	fmt.Printf(
		"%s invoked daemon %q (group=%q) with PID %d\n",
		chalk.Green.Color("OK:"),
		daemonName,
		groupName,
		pid,
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func isDaemonActive(meta *daemonMeta) bool {
	if meta.PID <= 0 {
		return false
	}
	proc, err := os.FindProcess(meta.PID)
	if err != nil {
		return false
	}
	// Unix-like: signal 0 checks existence; EPERM implies it's running but not signalable.
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return errors.Is(err, syscall.EPERM)
	}
	return true
}

func mustExpand(val, label string) string {
	const op = "expand path"
	expanded, err := expandPath(val)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage(fmt.Sprintf("expanding %s path", label)))
	return expanded
}

// Formatter and helper
var OneLineRedFormatter horus.FormatterFunc = func(he *horus.Herror) string {
	return chalk.Red.Color(he.Message)
}

func exitAlreadyRunning(name string) {
	horus.CheckErr(
		fmt.Errorf("daemon already running"),
		horus.WithOp("invoke"),
		horus.WithMessage(fmt.Sprintf("daemon %q is already running", name)),
		horus.WithCategory("spawn_error"),
		horus.WithExitCode(2),
		horus.WithFormatter(OneLineRedFormatter),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
