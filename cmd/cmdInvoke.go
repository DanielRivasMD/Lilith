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
	"os/exec"
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

var invokeFlags struct {
	daemon   string
	group    string
	watch    string
	script   string
	log      string
	allGroup string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func InvokeCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("invoke", runInvoke,
		domovoi.WithValidArgsFunction(completeWorkflowNames),
	))

	cmd.Flags().StringVarP(&invokeFlags.daemon, "daemon", "", "", "Daemon instance name (defaults to config key)")
	cmd.Flags().StringVarP(&invokeFlags.group, "group", "", "default", "Watcher group name (overrides TOML). Default value: `default`")
	cmd.Flags().StringVarP(&invokeFlags.watch, "watch", "", "", "Directory to watch (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.script, "script", "", "", "Script to execute on change (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.log, "log", "", "", "Name for log file (no `.log` extension; required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.allGroup, "all-group", "G", "", "Invoke all daemons in the specified group (config file name)")

	cmd.PreRun = preInvoke
	horus.CheckErr(cmd.RegisterFlagCompletionFunc("group", completeWorkflowGroups), horus.WithOp("invoke.init"), horus.WithMessage("registering config completion"))

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func preInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke.pre"

	if invokeFlags.allGroup != "" {
		return
	}

	if len(args) == 1 {
		// CONFIG MODE
		if rootFlags.verbose {
			fmt.Println("Running on Config mode...")
		}
		configName := args[0]

		files, err := domovoi.ReadDir(configDirs.config, rootFlags.verbose)
		horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage("reading config dir"))
		var foundV *viper.Viper
		var configFileUsed string
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".toml") {
				continue
			}
			path := filepath.Join(configDirs.config, f.Name())
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
				horus.WithFormatter(func(he *horus.Herror) string { return horus.OneLineErr(he.Message) }),
			)
		}

		groupPrefix := strings.TrimSuffix(filepath.Base(configFileUsed), filepath.Ext(configFileUsed))

		if invokeFlags.daemon == "" {
			invokeFlags.daemon = groupPrefix + "-" + configName
			horus.CheckErr(cmd.Flags().Set("daemon", invokeFlags.daemon), horus.WithOp(op), horus.WithMessage("setting default --daemon with group prefix"))
		}

		wf := foundV.Sub("workflows." + configName)
		bindFlag(cmd, "watch", wf)
		bindFlag(cmd, "script", wf)

		if !cmd.Flags().Changed("group") {
			invokeFlags.group = groupPrefix
			horus.CheckErr(cmd.Flags().Set("group", invokeFlags.group), horus.WithOp(op), horus.WithMessage("setting default --group"))
		}

		if !cmd.Flags().Changed("log") {
			invokeFlags.log = groupPrefix + "-" + configName
			horus.CheckErr(cmd.Flags().Set("log", invokeFlags.log), horus.WithOp(op), horus.WithMessage("setting default --log with group prefix"))
		}
	} else {
		// MANUAL MODE
		if rootFlags.verbose {
			fmt.Println("Running on Manual mode...")
		}
		horus.CheckEmpty(
			invokeFlags.daemon,
			"",
			horus.WithMessage("`--daemon` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			invokeFlags.watch,
			"",
			horus.WithMessage("`--watch` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			invokeFlags.script,
			"",
			horus.WithMessage("`--script` is required"),
			horus.WithExitCode(2),
			horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
		)
		horus.CheckEmpty(
			invokeFlags.log,
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

	// GROUP MODE: invoke all workflows in a group config file
	if invokeFlags.allGroup != "" {
		groupName := invokeFlags.allGroup
		configPath := filepath.Join(configDirs.config, groupName+".toml")
		v := viper.New()
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("reading group config %s", configPath)))
		}
		workflows := v.GetStringMap("workflows")
		if len(workflows) == 0 {
			fmt.Printf("No workflows found in group %q\n", groupName)
			return
		}
		for wfName := range workflows {
			// Build daemon name and log name from group + workflow
			daemonName := groupName + "-" + wfName
			logName := daemonName
			wf := v.Sub("workflows." + wfName)
			watchDir := wf.GetString("watch")
			scriptPath := wf.GetString("script")
			if watchDir == "" || scriptPath == "" {
				fmt.Printf("Skipping %s: missing watch or script\n", wfName)
				continue
			}
			// Expand tilde
			watchDir = strings.Replace(watchDir, "~", configDirs.home, 1)
			scriptPath = strings.Replace(scriptPath, "~", configDirs.home, 1)

			meta := &daemonMeta{
				Daemon:     daemonName,
				Group:      groupName,
				WatchDir:   watchDir,
				ScriptPath: scriptPath,
				LogPath:    filepath.Join(configDirs.log, logName+".log"),
				InvokedAt:  time.Now(),
			}

			// Check for already running daemon with same name+group
			duplicate := false
			for _, path := range listDaemonMetaFiles() {
				existing := loadMeta(path)
				if existing.Daemon == meta.Daemon && existing.Group == meta.Group && isDaemonActive(existing) {
					fmt.Printf("Daemon %s already running, skipping\n", meta.Daemon)
					duplicate = true
					break
				}
			}
			if duplicate {
				continue
			}

			meta.PID = spawnWatcher(meta)
			saveMeta(meta)
			fmt.Printf("%s invoked daemon %s group %s PID %s\n",
				chalk.Green.Color("OK:"), meta.Daemon, meta.Group, chalk.Green.Color(strconv.Itoa(meta.PID)))
		}
		return
	}

	// SINGLE INVOCATION (manual or config mode)
	// expand tilde
	invokeFlags.watch = strings.Replace(invokeFlags.watch, "~", configDirs.home, 1)
	invokeFlags.script = strings.Replace(invokeFlags.script, "~", configDirs.home, 1)

	meta := &daemonMeta{
		Daemon:     invokeFlags.daemon,
		Group:      invokeFlags.group,
		WatchDir:   invokeFlags.watch,
		ScriptPath: invokeFlags.script,
		LogPath:    filepath.Join(configDirs.log, invokeFlags.log+".log"),
		InvokedAt:  time.Now(),
	}

	// check for already running daemon with same name+group
	for _, path := range listDaemonMetaFiles() {
		existingMeta := loadMeta(path)
		if existingMeta.Daemon == invokeFlags.daemon && existingMeta.Group == invokeFlags.group && isDaemonActive(existingMeta) {
			horus.CheckErr(
				errors.New(""),
				horus.WithOp(op),
				horus.WithMessage(existingMeta.Daemon),
				horus.WithExitCode(2),
				horus.WithFormatter(func(he *horus.Herror) string {
					return "daemon " + horus.OneLineErr(he.Message) + " already running"
				}),
			)
		}
	}

	meta.PID = spawnWatcher(meta)
	saveMeta(meta)

	fmt.Printf(
		"%s invoked daemon %s group %s PID %s\n",
		chalk.Green.Color("OK:"),
		chalk.Green.Color(invokeFlags.daemon),
		chalk.Green.Color(invokeFlags.group),
		chalk.Green.Color(strconv.Itoa(meta.PID)),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func spawnWatcher(meta *daemonMeta) int {
	const op = "daemon.spawnWatcher"

	horus.CheckErr(
		domovoi.CreateDir(configDirs.log, rootFlags.verbose),
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage("creating log directory"),
		horus.WithDetails(map[string]any{
			"logDir": configDirs.log,
		}),
	)

	cmd := exec.Command(
		"watchexec",
		"--watch", meta.WatchDir,
		"--no-vcs-ignore",
		"--",
		"bash", meta.ScriptPath,
	)

	f, err := os.OpenFile(
		meta.LogPath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o644,
	)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage("opening log file"),
		horus.WithDetails(map[string]any{
			"logPath": meta.LogPath,
		}),
	)
	defer f.Close()

	cmd.Stdout = f
	cmd.Stderr = f

	horus.CheckErr(
		cmd.Start(),
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage("starting watcher process"),
		horus.WithDetails(map[string]any{
			"watchDir":   meta.WatchDir,
			"scriptPath": meta.ScriptPath,
		}),
	)

	pid := cmd.Process.Pid

	horus.CheckErr(
		cmd.Process.Release(),
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage("releasing watcher process"),
		horus.WithDetails(map[string]any{
			"pid": pid,
		}),
	)

	return pid
}

////////////////////////////////////////////////////////////////////////////////////////////////////
