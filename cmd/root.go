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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

var rootCmd = &cobra.Command{
	Use:     "lilith",
	Long:    helpRoot,
	Example: exampleRoot,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	home    string
	verbose bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostics")
	cobra.OnInitialize(initHome)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// daemonMeta holds persistent info about process
type daemonMeta struct {
	Name       string    `json:"name"`
	Group      string    `json:"group"`
	WatchDir   string    `json:"watchDir"`
	ScriptPath string    `json:"scriptPath"`
	LogPath    string    `json:"logPath"`
	PID        int       `json:"pid"`
	InvokedAt  time.Time `json:"invokedAt"`
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	lilithDir = filepath.Join(home, ".lilith")
	configDir = filepath.Join(lilithDir, "config")
	logDir    = filepath.Join(lilithDir, "logs")
	daemonDir = filepath.Join(lilithDir, "daemon")
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func initHome() {
	var err error
	home, err = domovoi.FindHome(verbose)
	horus.CheckErr(err,
		horus.WithCategory("env_error"),
		horus.WithMessage("getting home directory"),
	)
}

func errorFmt(er string) string {
	return chalk.Bold.TextStyle(chalk.Red.Color(er))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// bindFlag will take a pointer dest *T, pull the T out of cfg if
// the flag was not changed and cfg has that key, store it in *dest
// and then call cmd.Flags().Set(flagName, fmt.Sprint(val)) so that
// cobra/pflag also knows about it
//
//	getVal is typically v.GetString, v.GetInt, v.GetBool, etc
func bindFlag[T any](
	cmd *cobra.Command,
	flagName string,
	dest *T,
	cfg *viper.Viper,
	getVal func(*viper.Viper, string) T,
) {
	const op = "viper.bindFlag"

	flags := cmd.Flags()
	if flags.Changed(flagName) || !cfg.IsSet(flagName) {
		return
	}

	val := getVal(cfg, flagName)
	*dest = val

	str := fmt.Sprint(val)
	if err := flags.Set(flagName, str); err != nil {
		horus.CheckErr(
			horus.NewCategorizedHerror(
				op,
				"viper_error",
				fmt.Sprintf("setting %q from config", flagName),
				err,
				map[string]any{"flag": flagName, "value": str},
			),
		)
	}
}

func bindString(cmd *cobra.Command, flagName string, dest *string, cfg *viper.Viper) {
	bindFlag(cmd, flagName, dest, cfg, (*viper.Viper).GetString)
}
func bindInt(cmd *cobra.Command, flagName string, dest *int, cfg *viper.Viper) {
	bindFlag(cmd, flagName, dest, cfg, (*viper.Viper).GetInt)
}
func bindBool(cmd *cobra.Command, flagName string, dest *bool, cfg *viper.Viper) {
	bindFlag(cmd, flagName, dest, cfg, (*viper.Viper).GetBool)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// saveMeta writes meta to ~/.lilith/daemon/<name>.json
func saveMeta(meta *daemonMeta) {
	const op = "daemon.saveMeta"

	if err := domovoi.CreateDir(daemonDir, verbose); err != nil {
		horus.Wrap(err, op, "creating daemon directory")
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		horus.NewCategorizedHerror(
			op, "encode_error", "marshaling metadata", err,
			map[string]any{"name": meta.Name},
		)
	}

	path := filepath.Join(daemonDir, meta.Name+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		horus.NewCategorizedHerror(
			op, "env_error", "writing metadata file", err,
			map[string]any{"path": path},
		)
	}

}

// loadMeta reads ~/.lilith/daemon/<name>.json
func loadMeta(name string) *daemonMeta {
	const op = "daemon.loadMeta"
	path := filepath.Join(daemonDir, name+".json")

	data, err := os.ReadFile(path)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("env_error"),
		horus.WithMessage("reading metadata file"),
		horus.WithDetails(map[string]any{
			"path": path,
			"name": name,
		}),
	)

	var meta daemonMeta
	horus.CheckErr(
		json.Unmarshal(data, &meta),
		horus.WithOp(op),
		horus.WithCategory("decode_error"),
		horus.WithMessage("unmarshaling metadata"),
		horus.WithDetails(map[string]any{
			"path": path,
			"name": name,
		}),
	)

	return &meta
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
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return errors.Is(err, syscall.EPERM)
	}
	return true
}

func listDaemonMetaFiles() []string {
	op := "daemon.list"
	daemonPattern := filepath.Join(daemonDir, "*.json")
	matches, err := filepath.Glob(daemonPattern)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithMessage("listing daemon metadata files"),
		horus.WithDetails(map[string]any{"pattern": daemonPattern}),
	)
	return matches
}

func getDaemonName(path string) string {
	return filepath.Base(path[:len(path)-len(".json")])
}

func matchDaemonGroup(metaPath, expectedGroup string) bool {
	// Try to load JSON metadata
	data, err := os.ReadFile(metaPath)
	if err != nil {
		// optionally log or ignore
		return false
	}

	var meta struct {
		Group string `json:"group"`
	}

	if err := json.Unmarshal(data, &meta); err != nil {
		// if unmarshal fails, ignore this file
		return false
	}

	return meta.Group == expectedGroup
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// spawnWatcher starts watchexec, redirects logs, returns its PID
func spawnWatcher(meta *daemonMeta) int {
	const op = "daemon.spawnWatcher"
	logDir := filepath.Dir(meta.LogPath)

	if err := domovoi.CreateDir(logDir, false); err != nil {
		horus.Wrap(err, op, "creating log directory")
	}

	cmd := exec.Command("watchexec",
		"--watch", meta.WatchDir,
		"--",
		"bash", meta.ScriptPath,
	)

	f, err := os.OpenFile(meta.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		horus.NewCategorizedHerror(
			op, "env_error", "opening log file", err,
			map[string]any{"logPath": meta.LogPath},
		)
	}
	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Start(); err != nil {
		_ = f.Close()
		horus.NewCategorizedHerror(
			op, "spawn_error", "starting watcher process", err,
			map[string]any{"watch": meta.WatchDir, "script": meta.ScriptPath},
		)
	}
	pid := cmd.Process.Pid

	if err := cmd.Process.Release(); err != nil {
		return pid
	}

	return pid
}

func sendDaemonSignal(pid int, sig syscall.Signal) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Errorf("could not find process %d: %w", pid, err)
	}
	proc.Signal(sig)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// completeWorkflowNames scans ~/.lilith/config/*.toml for [workflows.<name>] keys
func completeWorkflowNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	home, err := domovoi.FindHome(false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	cfgDir := filepath.Join(home, ".lilith", "config")
	fis, err := os.ReadDir(cfgDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	seen := map[string]struct{}{}
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
		for wf := range v.GetStringMap("workflows") {
			if strings.HasPrefix(wf, toComplete) {
				seen[wf] = struct{}{}
			}
		}
	}

	var out []string
	for wf := range seen {
		out = append(out, wf)
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// completeDaemonNames offers tab‐completion based on ~/.lilith/daemons/*.json
func completeDaemonNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fis, err := os.ReadDir(daemonDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var out []string
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		name := strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name()))
		if strings.HasPrefix(name, toComplete) {
			out = append(out, name)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// completeWorkflowGroups offer tab-completion based on on ~/.lilith/daemons/*.json
func completeWorkflowGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	files, err := filepath.Glob(filepath.Join(daemonDir, "*.json"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	groups := map[string]bool{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var meta struct {
			Group string `json:"group"`
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		if meta.Group != "" {
			groups[meta.Group] = true
		}
	}

	var availableGroups []string
	for g := range groups {
		availableGroups = append(availableGroups, g)
	}
	return availableGroups, cobra.ShellCompDirectiveDefault
}

////////////////////////////////////////////////////////////////////////////////////////////////////
