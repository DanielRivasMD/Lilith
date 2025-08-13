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

	PreRunE: PreInvoke,
	Run:     RunInvoke,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	ConfigName string // workflow key
	DaemonName string // instance name, defaults to configName
	WatchDir   string
	ScriptPath string
	LogName    string
	GroupName  string // derived from TOML filename
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&ConfigName, "config", "c", "", "Workflow to apply")
	invokeCmd.Flags().StringVarP(&DaemonName, "name", "n", "", "Unique daemon name (defaults to --config)")
	invokeCmd.Flags().StringVarP(&GroupName, "group", "g", "", "Watcher group name (overrides TOML)")
	invokeCmd.Flags().StringVarP(&WatchDir, "watch", "w", "", "Directory to watch")
	invokeCmd.Flags().StringVarP(&ScriptPath, "script", "s", "", "Script to execute on change")
	invokeCmd.Flags().StringVarP(&LogName, "log", "l", "", "Name for log file (no `.log` extension)")

	horus.CheckErr(invokeCmd.RegisterFlagCompletionFunc("config", completeWorkflowNames), horus.WithOp("invoke.init"), horus.WithMessage("registering config completion"))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpInvoke = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Spawn daemon process for the specified directory & execute the configured script on change\n"+
		"Metadata is persistent for summoning the daemon",
)

var exampleInvoke = formatExample(
	"lilith",
	[]string{"invoke", "--config", "helix"},
	[]string{
		"invoke", "--name", "helix",
		"--watch", "~/src/helix",
		"--script", "helix.sh",
		"--log", "helix",
	},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func PreInvoke(cmd *cobra.Command, args []string) error {
	const op = "lilith.invoke.pre"

	home, err := domovoi.FindHome(verbose)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("env_error"),
		horus.WithMessage("getting home directory"),
	)
	cfgDir := filepath.Join(home, ".lilith", "config")

	var (
		foundV      *viper.Viper
		cfgFileUsed string
	)
	fis, err := domovoi.ReadDir(cfgDir, verbose)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("env_error"),
		horus.WithMessage("reading config dir"),
	)

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
		if v.IsSet("workflows." + ConfigName) {
			foundV = v
			cfgFileUsed = path
			break
		}
	}

	if foundV == nil {
		horus.CheckErr(
			fmt.Errorf("workflow %q not found in %s/*.toml", ConfigName, cfgDir),
			horus.WithOp(op),
			horus.WithMessage("could not find named workflow in config directory"),
			horus.WithCategory("config_error"),
		)
	}

	if DaemonName == "" {
		DaemonName = ConfigName
		horus.CheckErr(
			cmd.Flags().Set("name", DaemonName),
			horus.WithOp(op),
			horus.WithMessage("setting default --name from config"),
			horus.WithCategory("config_error"),
		)
	}

	base := filepath.Base(cfgFileUsed)
	GroupName = strings.TrimSuffix(base, filepath.Ext(base))
	horus.CheckErr(
		cmd.Flags().Set("group", GroupName),
		horus.WithOp(op),
		horus.WithMessage("setting default --group from TOML filename"),
		horus.WithCategory("config_error"),
	)

	wf := foundV.Sub("workflows." + ConfigName)
	BindFlag(cmd, "watch", &WatchDir, wf)
	BindFlag(cmd, "script", &ScriptPath, wf)

	if !cmd.Flags().Changed("log") {
		LogName = ConfigName
		horus.CheckErr(
			cmd.Flags().Set("log", LogName),
			horus.WithOp(op),
			horus.WithMessage("setting default --log from workflow key"),
			horus.WithCategory("config_error"),
		)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func RunInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke"

	horus.CheckEmpty(
		WatchDir,
		"`--watch` is required",
		horus.WithOp(op),
		horus.WithMessage("provide a directory to watch"),
		horus.WithCategory("spawn_error"),
	)
	horus.CheckEmpty(
		ScriptPath,
		"`--script` is required",
		horus.WithOp(op),
		horus.WithMessage("provide a script to run"),
		horus.WithCategory("spawn_error"),
	)
	horus.CheckEmpty(
		LogName,
		"`--log` is required",
		horus.WithOp(op),
		horus.WithMessage("provide a log name"),
		horus.WithCategory("spawn_error"),
	)

	WatchDir = mustExpand(WatchDir, "--watch")
	ScriptPath = mustExpand(ScriptPath, "--script")

	home, err := domovoi.FindHome(verbose)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("getting home directory"))
	logDir := filepath.Join(home, ".lilith", "logs")
	horus.CheckErr(
		domovoi.CreateDir(logDir, verbose),
		horus.WithOp(op),
		horus.WithMessage(fmt.Sprintf("creating %q", logDir)),
		horus.WithCategory("env_error"),
	)
	logPath := filepath.Join(logDir, LogName+".log")

	meta := &DaemonMeta{
		Name:       DaemonName,
		Group:      GroupName,
		WatchDir:   WatchDir,
		ScriptPath: ScriptPath,
		LogPath:    logPath,
		InvokedAt:  time.Now(),
	}

	for _, path := range mustListDaemonMetaFiles() {
		existing := mustLoadMeta(path)
		if existing.WatchDir == WatchDir && isDaemonActive(existing) {
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

	pid, err := spawnWatcher(meta)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("env_error"),
		horus.WithMessage("starting watcher"),
	)
	meta.PID = pid

	horus.CheckErr(
		saveMeta(meta),
		horus.WithOp(op),
		horus.WithCategory("env_error"),
		horus.WithMessage("writing metadata"),
	)

	fmt.Printf(
		"invoked daemon %s group %s PID %s\n",
		chalk.Green.Color(DaemonName),
		chalk.Green.Color(GroupName),
		chalk.Green.Color(strconv.Itoa(pid)),
	)
	// BUG: cannot execute daemons passed on the command line
}

////////////////////////////////////////////////////////////////////////////////////////////////////
