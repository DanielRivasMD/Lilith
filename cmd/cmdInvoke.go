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
	Use:     "invoke " + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<config>")),
	Short:   "Start daemon",
	Long:    helpInvoke,
	Example: exampleInvoke,

	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeWorkflowNames,

	PreRun: preInvoke,
	Run:    runInvoke,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	configName string // workflow key
	daemonName string // instance name, defaults to configName
	watchPath  string
	scriptPath string
	logName    string
	groupName  string // derived from TOML filename
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&daemonName, "daemon", "", "", "Daemon instance name (defaults to config key)")
	invokeCmd.Flags().StringVarP(&groupName, "group", "", "default", "Watcher group name (overrides TOML). Default value: `default`")
	invokeCmd.Flags().StringVarP(&watchPath, "watch", "", "", "Directory to watch (required in manual mode)")
	invokeCmd.Flags().StringVarP(&scriptPath, "script", "", "", "Script to execute on change (required in manual mode)")
	invokeCmd.Flags().StringVarP(&logName, "log", "", "", "Name for log file (no `.log` extension; required in manual mode)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func preInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke.pre"

	if len(args) == 1 {
		// CONFIG MODE: pull everything from TOML
		if verbose {
			fmt.Println("Running on Config mode...")
		}

		// declare workflow
		configName = args[0]

		// discover matching workflow file
		files, err := domovoi.ReadDir(configDir, verbose)
		horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("reading config dir"))
		var foundV *viper.Viper
		var configFileUsed string
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".toml") {
				continue
			}
			path := filepath.Join(configDir, f.Name())
			v := viper.New()
			v.SetConfigFile(path)
			if err := v.ReadInConfig(); err != nil {
				continue
			}
			if v.IsSet("workflows." + configName) {
				foundV = v
				configFileUsed = path
				break
			}
		}
		if foundV == nil {
			horus.CheckErr(
				errors.New(""),
				horus.WithMessage(fmt.Sprintf("workflow %s not found", configName)),
				horus.WithFormatter(func(he *horus.Herror) string { return errorFmt(he.Message) }),
			)
		}

		// defaults
		if daemonName == "" {
			daemonName = configName
			horus.CheckErr(cmd.Flags().Set("daemon", daemonName), horus.WithOp(op), horus.WithMessage("setting default --daemon"))
		}

		// bind watch & script from TOML
		wf := foundV.Sub("workflows." + configName)
		bindFlag(cmd, "watch", wf)
		bindFlag(cmd, "script", wf)

		// group default
		if !cmd.Flags().Changed("group") {
			base := filepath.Base(configFileUsed)
			groupName = strings.TrimSuffix(base, filepath.Ext(base))
			horus.CheckErr(cmd.Flags().Set("group", groupName), horus.WithOp(op), horus.WithMessage("setting default --group"))
		}

		// log default
		if !cmd.Flags().Changed("log") {
			logName = configName
			horus.CheckErr(cmd.Flags().Set("log", logName), horus.WithOp(op), horus.WithMessage("setting default --log"))
		}

	} else {

		// MANUAL MODE: require explicit flags
		if verbose {
			fmt.Println("Running on Manual mode...")
		}

		horus.CheckEmpty(
			watchPath,
			"",
			horus.WithMessage("`--watch` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			scriptPath,
			"",
			horus.WithMessage("`--script` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			logName,
			"",
			horus.WithMessage("`--log` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke"

	// format paths
	watchPath = strings.Replace(watchPath, "~", home, 1)
	scriptPath = strings.Replace(scriptPath, "~", home, 1)
	logPath := filepath.Join(logDir, logName+".log")

	// declare meta
	meta := &daemonMeta{
		Name:       daemonName,
		Group:      groupName,
		WatchDir:   watchPath,
		ScriptPath: scriptPath,
		LogPath:    logPath,
		InvokedAt:  time.Now(),
	}

	// check running daemons
	for _, path := range listDaemonMetaFiles() {
		existingMeta := loadMeta(path)
		if existingMeta.WatchDir == watchPath && isDaemonActive(existingMeta) {
			horus.CheckErr(
				errors.New(""),
				horus.WithOp(op),
				horus.WithMessage(existingMeta.Name),
				horus.WithExitCode(2),
				horus.WithFormatter(func(he *horus.Herror) string {
					return "daemon " + errorFmt(he.Message) + " already running"
				}),
			)
		}
	}

	// launch watch
	meta.PID = spawnWatcher(meta)

	// record meta
	saveMeta(meta)

	// log meta
	fmt.Printf(
		"invoked daemon %s group %s PID %s\n",
		chalk.Green.Color(daemonName),
		chalk.Green.Color(groupName),
		chalk.Green.Color(strconv.Itoa(meta.PID)),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
