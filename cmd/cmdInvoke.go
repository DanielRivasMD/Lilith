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
	"strconv"
	"strings"
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
	Short:   "Start daemon",
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

	home, err := findHomeFn(verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("getting home directory"))
	cfgDir := filepath.Join(home, ".lilith", "config")

	var (
		foundV      *viper.Viper
		cfgFileUsed string
	)
	fis, err := readDirFn(cfgDir, verbose)
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

	if foundV == nil {
		horus.CheckErr(
			fmt.Errorf("workflow %q not found in %s/*.toml", configName, cfgDir),
			horus.WithOp(op),
			horus.WithMessage("could not find named workflow in config directory"),
			horus.WithCategory("config_error"),
		)
	}

	if daemonName == "" {
		daemonName = configName
		horus.CheckErr(
			cmd.Flags().Set("name", daemonName),
			horus.WithOp(op),
			horus.WithMessage("setting default --name from config"),
			horus.WithCategory("config_error"),
		)
	}

	base := filepath.Base(cfgFileUsed)
	groupName = strings.TrimSuffix(base, filepath.Ext(base))
	horus.CheckErr(
		cmd.Flags().Set("group", groupName),
		horus.WithOp(op),
		horus.WithMessage("setting default --group from TOML filename"),
		horus.WithCategory("config_error"),
	)

	wf := foundV.Sub("workflows." + configName)
	bindFlag(cmd, "watch", &watchDir, wf)
	bindFlag(cmd, "script", &scriptPath, wf)

	if !cmd.Flags().Changed("log") {
		logName = configName
		horus.CheckErr(
			cmd.Flags().Set("log", logName),
			horus.WithOp(op),
			horus.WithMessage("setting default --log from workflow key"),
			horus.WithCategory("config_error"),
		)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke"

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

	watchDir = mustExpand(watchDir, "--watch")
	scriptPath = mustExpand(scriptPath, "--script")

	home, err := findHomeFn(verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("getting home directory"))
	logDir := filepath.Join(home, ".lilith", "logs")
	horus.CheckErr(
		createDirFn(logDir, verbose),
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("creating %q", logDir)),
		horus.WithCategory("env_error"),
	)
	logPath := filepath.Join(logDir, logName+".log")

	meta := &daemonMeta{
		Name:       daemonName,
		Group:      groupName,
		WatchDir:   watchDir,
		ScriptPath: scriptPath,
		LogPath:    logPath,
		InvokedAt:  nowFn(),
	}

	for _, path := range listMetaFilesFn() {
		existing := loadMetaFn(path)
		if existing.WatchDir == watchDir && isDaemonActiveFn(&existing) {
			horus.CheckErr(
				fmt.Errorf("daemon already running"),
				horus.WithMessage(existing.Name),
				horus.WithExitCode(2),
				horus.WithFormatter(func(he *horus.Herror) string {
					return "daemon " + chalk.Red.Color(he.Message) + " already running"
				}),
			)
		}
	}

	pid, err := spawnWatcherFn(meta)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("starting watcher"))
	meta.PID = pid

	horus.CheckErr(saveMetaFn(meta), horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("writing metadata"))

	fmt.Printf(
		"invoked daemon %s group %s PID %s\n",
		chalk.Green.Color(daemonName),
		chalk.Green.Color(groupName),
		chalk.Green.Color(strconv.Itoa(pid)),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// seams for testing (default to real funcs)
var (
	spawnWatcherFn   = spawnWatcher
	saveMetaFn       = saveMeta
	listMetaFilesFn  = mustListDaemonMetaFiles
	loadMetaFn       = mustLoadMeta
	isDaemonActiveFn = isDaemonActive
	expandPathFn     = expandPath
	findHomeFn       = domovoi.FindHome
	createDirFn      = domovoi.CreateDir
	readDirFn        = domovoi.ReadDir
	nowFn            = time.Now
)

////////////////////////////////////////////////////////////////////////////////////////////////////
