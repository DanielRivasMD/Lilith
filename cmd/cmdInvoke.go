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
	daemonName   string
	daemonGroup  string
	daemonWatch  string
	daemonScript string
	daemonLog    string
	group        string
	all          bool
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func InvokeCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("invoke", runInvoke,
		domovoi.WithValidArgsFunction(completeWorkflowNames),
	))

	// Manual mode flags
	cmd.Flags().StringVarP(&invokeFlags.daemonName, "daemon-name", "", "", "Daemon instance name (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.daemonGroup, "daemon-group", "", "default", "Group name for this daemon (overrides TOML). Default: `default`")
	cmd.Flags().StringVarP(&invokeFlags.daemonWatch, "daemon-watch", "", "", "Directory to watch (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.daemonScript, "daemon-script", "", "", "Script to execute on change (required in manual mode)")
	cmd.Flags().StringVarP(&invokeFlags.daemonLog, "daemon-log", "", "", "Name for log file (no .log extension; required in manual mode)")

	// Group / all modes
	cmd.Flags().StringVarP(&invokeFlags.group, "group", "g", "", "Invoke all workflows in this group (config file name without .toml)")
	cmd.Flags().BoolVarP(&invokeFlags.all, "all", "a", false, "Invoke all workflows from all config files")

	// Completion for --group flag (config file names)
	horus.CheckErr(cmd.RegisterFlagCompletionFunc("group", completeConfigGroups), horus.WithOp("invoke.init"), horus.WithMessage("registering group completion"))

	// PreRun only needed for manual mode validation
	cmd.PreRun = preInvokeManual

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// preInvokeManual checks that required flags are present in manual mode
func preInvokeManual(cmd *cobra.Command, args []string) {
	// If we are in group/all mode or config mode (has args), skip validation
	if invokeFlags.group != "" || invokeFlags.all || len(args) == 1 {
		return
	}

	// Manual mode: require daemon, watch, script, log
	if rootFlags.verbose {
		fmt.Println("Running on Manual mode...")
	}
	horus.CheckEmpty(
		invokeFlags.daemonName,
		"",
		horus.WithMessage("`--daemon` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
	horus.CheckEmpty(
		invokeFlags.daemonWatch,
		"",
		horus.WithMessage("`--daemon-watch` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
	horus.CheckEmpty(
		invokeFlags.daemonScript,
		"",
		horus.WithMessage("`--daemon-script` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
	horus.CheckEmpty(
		invokeFlags.daemonLog,
		"",
		horus.WithMessage("`--log` is required"),
		horus.WithExitCode(2),
		horus.WithFormatter(func(he *horus.Herror) string { return chalk.Red.Color(he.Message) }),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runInvoke(cmd *cobra.Command, args []string) {
	const op = "lilith.invoke"

	// Mode 1: --all
	if invokeFlags.all {
		invokeAllWorkflows()
		return
	}

	// Mode 2: --group <name>
	if invokeFlags.group != "" {
		invokeGroupWorkflows(invokeFlags.group)
		return
	}

	// Mode 3: config mode (one or more workflow names as arguments)
	if len(args) >= 1 {
		// Collect all workflow names from arguments (allow multiple)
		workflowNames := args
		invokeNamedWorkflows(workflowNames)
		return
	}

	// Mode 4: manual mode (all flags provided)
	invokeManual()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// invokeAllWorkflows reads every .toml in config dir and invokes all workflows
func invokeAllWorkflows() {
	configDir := configDirs.config
	entries, err := os.ReadDir(configDir)
	if err != nil {
		horus.CheckErr(err, horus.WithOp("invoke.all"), horus.WithMessage("reading config directory"))
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		groupName := strings.TrimSuffix(entry.Name(), ".toml")
		invokeGroupWorkflows(groupName)
	}
}

// invokeGroupWorkflows invokes all workflows inside a specific config file (group)
func invokeGroupWorkflows(groupName string) {
	configPath := filepath.Join(configDirs.config, groupName+".toml")
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Warning: cannot read group config %s: %v\n", configPath, err)
		return
	}
	workflows := v.GetStringMap("workflows")
	if len(workflows) == 0 {
		fmt.Printf("No workflows found in group %q\n", groupName)
		return
	}
	for wfName := range workflows {
		spawnWorkflowFromViper(v, wfName, groupName)
	}
}

// invokeNamedWorkflows finds all workflows with the given names across all config files and invokes them
func invokeNamedWorkflows(workflowNames []string) {
	// Map from workflow name to list of (group, viper, wfSub)
	type wfInfo struct {
		group string
		viper *viper.Viper
		wfSub *viper.Viper
	}
	// First, scan all config files and collect matching workflows
	configDir := configDirs.config
	entries, err := os.ReadDir(configDir)
	if err != nil {
		horus.CheckErr(err, horus.WithOp("invoke.named"), horus.WithMessage("reading config directory"))
	}
	// Build a set of desired workflow names for faster lookup
	wanted := make(map[string]bool)
	for _, name := range workflowNames {
		wanted[name] = true
	}
	matches := make(map[string][]wfInfo) // workflow name -> list of occurrences

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		groupName := strings.TrimSuffix(entry.Name(), ".toml")
		path := filepath.Join(configDir, entry.Name())
		v := viper.New()
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			continue
		}
		for wfName := range v.GetStringMap("workflows") {
			if wanted[wfName] {
				matches[wfName] = append(matches[wfName], wfInfo{
					group: groupName,
					viper: v,
					wfSub: v.Sub("workflows." + wfName),
				})
			}
		}
	}

	// Now spawn each match
	for _, wfName := range workflowNames {
		list, ok := matches[wfName]
		if !ok {
			fmt.Printf("Workflow %q not found in any config file\n", wfName)
			continue
		}
		for _, info := range list {
			spawnWorkflowFromViper(info.viper, wfName, info.group)
		}
	}
}

// spawnWorkflowFromViper creates and spawns a daemon from a viper workflow definition
func spawnWorkflowFromViper(v *viper.Viper, wfName, groupName string) {
	wf := v.Sub("workflows." + wfName)
	if wf == nil {
		return
	}
	watchDir := wf.GetString("watch")
	scriptPath := wf.GetString("script")
	if watchDir == "" || scriptPath == "" {
		fmt.Printf("Skipping %s/%s: missing watch or script\n", groupName, wfName)
		return
	}
	// Expand tilde
	watchDir = strings.Replace(watchDir, "~", configDirs.home, 1)
	scriptPath = strings.Replace(scriptPath, "~", configDirs.home, 1)

	// Determine daemon name: if "daemon" key exists, use it; else group-workflow
	daemonName := wf.GetString("daemon")
	if daemonName == "" {
		daemonName = groupName + "-" + wfName
	}
	// Log name: if "log" key exists, use it; else same as daemonName
	logName := wf.GetString("log")
	if logName == "" {
		logName = daemonName
	}
	// Group: if "group" key exists, use it; else the config file group
	group := wf.GetString("group")
	if group == "" {
		group = groupName
	}

	meta := &daemonMeta{
		Daemon:     daemonName,
		Group:      group,
		WatchDir:   watchDir,
		ScriptPath: scriptPath,
		LogPath:    filepath.Join(configDirs.log, logName+".log"),
		InvokedAt:  time.Now(),
	}

	// Check for already running daemon with same name+group
	for _, path := range listDaemonMetaFiles() {
		existing := loadMeta(path)
		if existing.Daemon == meta.Daemon && existing.Group == meta.Group && isDaemonActive(existing) {
			fmt.Printf("Daemon %s (group %s) already running, skipping\n", meta.Daemon, meta.Group)
			return
		}
	}

	meta.PID = spawnWatcher(meta)
	saveMeta(meta)
	fmt.Printf("%s invoked daemon %s group %s PID %s\n",
		chalk.Green.Color("OK:"), meta.Daemon, meta.Group, chalk.Green.Color(strconv.Itoa(meta.PID)))
}

// invokeManual handles the manual mode (all flags provided)
func invokeManual() {
	const op = "lilith.invoke.manual"

	// Expand tilde
	invokeFlags.daemonWatch = strings.Replace(invokeFlags.daemonWatch, "~", configDirs.home, 1)
	invokeFlags.daemonScript = strings.Replace(invokeFlags.daemonScript, "~", configDirs.home, 1)

	meta := &daemonMeta{
		Daemon:     invokeFlags.daemonName,
		Group:      invokeFlags.daemonGroup,
		WatchDir:   invokeFlags.daemonWatch,
		ScriptPath: invokeFlags.daemonScript,
		LogPath:    filepath.Join(configDirs.log, invokeFlags.daemonLog+".log"),
		InvokedAt:  time.Now(),
	}

	// Check for already running daemon with same name+group
	for _, path := range listDaemonMetaFiles() {
		existingMeta := loadMeta(path)
		if existingMeta.Daemon == invokeFlags.daemonName && existingMeta.Group == invokeFlags.daemonGroup && isDaemonActive(existingMeta) {
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
		chalk.Green.Color(invokeFlags.daemonName),
		chalk.Green.Color(invokeFlags.daemonGroup),
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

// completeConfigGroups provides tab completion for --group flag (config file names)
func completeConfigGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	entries, err := os.ReadDir(configDirs.config)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	var groups []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".toml")
		if strings.HasPrefix(name, toComplete) {
			groups = append(groups, name)
		}
	}
	return groups, cobra.ShellCompDirectiveNoFileComp
}

////////////////////////////////////////////////////////////////////////////////////////////////////
