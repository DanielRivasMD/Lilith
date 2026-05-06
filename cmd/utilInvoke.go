/*
Copyright © 2026 Daniel Rivas <danielrivasmd@gmail.com>

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
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: consider refactoring to eliminate redundant code => tilde expansion into function, run once functions into generics to cognate watchers
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

func invokeNamedWorkflows(workflowNames []string) {
	type wfInfo struct {
		group string
		viper *viper.Viper
		wfSub *viper.Viper
	}
	configDir := configDirs.config
	entries, err := os.ReadDir(configDir)
	if err != nil {
		horus.CheckErr(err, horus.WithOp("invoke.named"), horus.WithMessage("reading config directory"))
	}
	wanted := make(map[string]bool)
	for _, name := range workflowNames {
		wanted[name] = true
	}
	matches := make(map[string][]wfInfo)

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
	watchDir = strings.Replace(watchDir, "~", configDirs.home, 1)
	scriptPath = strings.Replace(scriptPath, "~", configDirs.home, 1)

	daemonName := wf.GetString("daemon")
	if daemonName == "" {
		daemonName = groupName + "-" + wfName
	}
	logName := wf.GetString("log")
	if logName == "" {
		logName = daemonName
	}
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

func invokeManual() {
	const op = "lilith.invoke.manual"

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

	for _, path := range listDaemonMetaFiles() {
		existingMeta := loadMeta(path)
		if existingMeta.Daemon == invokeFlags.daemonName && existingMeta.Group == invokeFlags.daemonGroup && isDaemonActive(existingMeta) {
			horus.CheckErr(
				errors.New("daemon already running: "),
				horus.WithMessage(existingMeta.Daemon),
				horus.WithExitCode(2),
				horus.WithFormatter(func(he *horus.Herror) string {
					return horus.OneLineErr(he.Message)
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

func runOnce(args []string) {
	if invokeFlags.all {
		runAllWorkflowsOnce()
		return
	}
	if invokeFlags.group != "" {
		runGroupWorkflowsOnce(invokeFlags.group)
		return
	}
	if len(args) >= 1 {
		runNamedWorkflowsOnce(args)
		return
	}
	runManualOnce()
}

func runAllWorkflowsOnce() {
	configDir := configDirs.config
	entries, err := os.ReadDir(configDir)
	if err != nil {
		horus.CheckErr(err, horus.WithOp("once.all"), horus.WithMessage("reading config directory"))
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		groupName := strings.TrimSuffix(entry.Name(), ".toml")
		runGroupWorkflowsOnce(groupName)
	}
}

func runGroupWorkflowsOnce(groupName string) {
	configPath := filepath.Join(configDirs.config, groupName+".toml")
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Warning: cannot read group config %s: %v\n", configPath, err)
		return
	}
	workflows := v.GetStringMap("workflows")
	for wfName := range workflows {
		runSingleWorkflowOnce(v, wfName, groupName)
	}
}

func runNamedWorkflowsOnce(workflowNames []string) {
	wanted := make(map[string]bool)
	for _, name := range workflowNames {
		wanted[name] = true
	}
	configDir := configDirs.config
	entries, err := os.ReadDir(configDir)
	if err != nil {
		horus.CheckErr(err, horus.WithOp("once.named"), horus.WithMessage("reading config directory"))
	}
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
				runSingleWorkflowOnce(v, wfName, groupName)
			}
		}
	}
}

func runSingleWorkflowOnce(v *viper.Viper, wfName, groupName string) {
	wf := v.Sub("workflows." + wfName)
	if wf == nil {
		return
	}
	scriptPath := wf.GetString("script")
	if scriptPath == "" {
		fmt.Printf("Skipping %s/%s: missing script\n", groupName, wfName)
		return
	}
	scriptPath = strings.Replace(scriptPath, "~", configDirs.home, 1)

	logName := wf.GetString("log")
	if logName == "" {
		logName = groupName + "-" + wfName
	}
	logPath := filepath.Join(configDirs.log, logName+".log")

	horus.CheckErr(domovoi.CreateDir(configDirs.log, rootFlags.verbose),
		horus.WithOp("once"), horus.WithMessage("creating log directory"))

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	horus.CheckErr(err, horus.WithOp("once"), horus.WithMessage("opening log file"))
	defer f.Close()

	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = f
	cmd.Stderr = f

	if rootFlags.verbose {
		fmt.Printf("Running %s/%s once, logging to %s\n", groupName, wfName, logPath)
	}
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running %s/%s: %v\n", groupName, wfName, err)
	} else {
		fmt.Printf("Successfully ran %s/%s once\n", groupName, wfName)
	}
}

func runManualOnce() {
	invokeFlags.daemonScript = strings.Replace(invokeFlags.daemonScript, "~", configDirs.home, 1)

	logPath := filepath.Join(configDirs.log, invokeFlags.daemonLog+".log")
	horus.CheckErr(domovoi.CreateDir(configDirs.log, rootFlags.verbose),
		horus.WithOp("once.manual"), horus.WithMessage("creating log directory"))

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	horus.CheckErr(err, horus.WithOp("once.manual"), horus.WithMessage("opening log file"))
	defer f.Close()

	cmd := exec.Command("bash", invokeFlags.daemonScript)
	cmd.Stdout = f
	cmd.Stderr = f

	if rootFlags.verbose {
		fmt.Printf("Running script %s once, logging to %s\n", invokeFlags.daemonScript, logPath)
	}
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running script: %v\n", err)
	} else {
		fmt.Println("Successfully ran script once")
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
