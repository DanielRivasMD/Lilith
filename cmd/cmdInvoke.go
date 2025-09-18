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
	paths configPaths
)

type configPaths struct {
	config string
	daemon string
	watch  string
	script string
	log    string
	group  string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&paths.daemon, "daemon", "", "", "Daemon instance name (defaults to config key)")
	invokeCmd.Flags().StringVarP(&paths.group, "group", "", "default", "Watcher group name (overrides TOML). Default value: `default`")
	invokeCmd.Flags().StringVarP(&paths.watch, "watch", "", "", "Directory to watch (required in manual mode)")
	invokeCmd.Flags().StringVarP(&paths.script, "script", "", "", "Script to execute on change (required in manual mode)")
	invokeCmd.Flags().StringVarP(&paths.log, "log", "", "", "Name for log file (no `.log` extension; required in manual mode)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func preInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke.pre"

	if len(args) == 1 {
		// CONFIG MODE: pull everything from TOML
		if flags.verbose {
			fmt.Println("Running on Config mode...")
		}

		// declare workflow
		paths.config = args[0]

		// discover matching workflow file
		files, err := domovoi.ReadDir(dirs.config, flags.verbose)
		horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("reading config dir"))
		var foundV *viper.Viper
		var configFileUsed string
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".toml") {
				continue
			}
			path := filepath.Join(dirs.config, f.Name())
			v := viper.New()
			v.SetConfigFile(path)
			if err := v.ReadInConfig(); err != nil {
				continue
			}
			if v.IsSet("workflows." + paths.config) {
				foundV = v
				configFileUsed = path
				break
			}
		}
		if foundV == nil {
			horus.CheckErr(
				errors.New(""),
				horus.WithMessage(fmt.Sprintf("workflow %s not found", paths.config)),
				horus.WithFormatter(func(he *horus.Herror) string { return onelineErr(he.Message) }),
			)
		}

		// defaults
		if paths.daemon == "" {
			paths.daemon = paths.config
			horus.CheckErr(cmd.Flags().Set("daemon", paths.daemon), horus.WithOp(op), horus.WithMessage("setting default --daemon"))
		}

		// bind watch & script from TOML
		wf := foundV.Sub("workflows." + paths.config)
		bindFlag(cmd, "watch", wf)
		bindFlag(cmd, "script", wf)

		// group default
		if !cmd.Flags().Changed("group") {
			base := filepath.Base(configFileUsed)
			paths.group = strings.TrimSuffix(base, filepath.Ext(base))
			horus.CheckErr(cmd.Flags().Set("group", paths.group), horus.WithOp(op), horus.WithMessage("setting default --group"))
		}

		// log default
		if !cmd.Flags().Changed("log") {
			paths.log = paths.config
			horus.CheckErr(cmd.Flags().Set("log", paths.log), horus.WithOp(op), horus.WithMessage("setting default --log"))
		}

	} else {

		// MANUAL MODE: require explicit flags
		if flags.verbose {
			fmt.Println("Running on Manual mode...")
		}

		horus.CheckEmpty(
			paths.daemon,
			"",
			horus.WithMessage("`--daemon` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			paths.watch,
			"",
			horus.WithMessage("`--watch` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			paths.script,
			"",
			horus.WithMessage("`--script` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			paths.log,
			"",
			horus.WithMessage("`--log` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: catch erros if command does not launch succesfully
func runInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke"

	// format paths
	paths.watch = strings.Replace(paths.watch, "~", dirs.home, 1)
	paths.script = strings.Replace(paths.script, "~", dirs.home, 1)

	// TODO: bind directly from paths?
	// declare meta
	meta := &daemonMeta{
		Daemon:     paths.daemon,
		Group:      paths.group,
		WatchDir:   paths.watch,
		ScriptPath: paths.script,
		LogPath:    filepath.Join(dirs.log, paths.log+".log"),
		InvokedAt:  time.Now(),
	}

	// check running daemons
	for _, path := range listDaemonMetaFiles() {
		existingMeta := loadMeta(path)
		if existingMeta.Daemon == paths.daemon && existingMeta.Group == paths.group && isDaemonActive(existingMeta) {
			horus.CheckErr(
				errors.New(""),
				horus.WithOp(op),
				horus.WithMessage(existingMeta.Daemon),
				horus.WithExitCode(2),
				horus.WithFormatter(func(he *horus.Herror) string {
					return "daemon " + onelineErr(he.Message) + " already running"
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
		chalk.Green.Color(paths.daemon),
		chalk.Green.Color(paths.group),
		chalk.Green.Color(strconv.Itoa(meta.PID)),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
