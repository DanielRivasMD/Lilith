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
	verbose bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostic output")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpRoot = chalk.Bold.TextStyle(chalk.Green.Color("Daniel Rivas ")) +
	chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) +
	chalk.Dim.TextStyle(chalk.Cyan.Color("\n\nmaster of daemons"))

var exampleRoot = chalk.White.Color("lilith") + ` ` + chalk.Bold.TextStyle(chalk.White.Color("help"))

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

// getDaemonDir returns ~/.lilith/daemon
func getDaemonDir() string {
	return filepath.Join(os.Getenv("HOME"), ".lilith", "daemon")
}

// saveMeta writes meta to ~/.lilith/daemon/<name>.json
func saveMeta(meta *daemonMeta) error {
	dir := getDaemonDir()
	// TODO: error handler => `horus`
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, meta.Name+".json"), data, 0644)
}

// loadMeta reads ~/.lilith/daemon/<name>.json
func loadMeta(name string) (*daemonMeta, error) {
	path := filepath.Join(getDaemonDir(), name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m daemonMeta
	return &m, json.Unmarshal(data, &m)
}

// spawnWatcher starts watchexec, redirects logs, returns its PID
func spawnWatcher(meta *daemonMeta) (int, error) {
	// TODO: error handler => `horus`
	if err := os.MkdirAll(filepath.Dir(meta.LogPath), 0755); err != nil {
		return 0, err
	}

	// TODO: exec cmd => domovoi?
	cmd := exec.Command("watchexec",
		"--watch", meta.WatchDir,
		"--",
		"bash", meta.ScriptPath,
	)
	f, err := os.OpenFile(meta.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Start(); err != nil {
		f.Close()
		return 0, err
	}
	pid := cmd.Process.Pid

	if err := cmd.Process.Release(); err != nil {
		return pid, err
	}
	return pid, nil
}

// bindFlag copies a Viper value into a flag variable if the flag was not set
func bindFlag(cmd *cobra.Command, flagName string, dest *string, cfg *viper.Viper) {
	if !cmd.Flags().Changed(flagName) && cfg.IsSet(flagName) {
		*dest = cfg.GetString(flagName)
		cmd.Flags().Set(flagName, *dest)
	}
}

func mustExpand(val, label string) string {
	const op = "expand.path"
	expanded, err := expandPathFn(val)
	horus.CheckErr(err, horus.WithOp(op), horus.WithCategory("env_error"), horus.WithMessage(fmt.Sprintf("expanding %s path", label)))
	return expanded
}

// expandPath replaces a leading "~" with $HOME and then does os.ExpandEnv.
func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~"+string(filepath.Separator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, p[2:]) // drop the "~/" and re-join
	}
	return os.ExpandEnv(p), nil
}

// completeWorkflowNames scans ~/.lilith/config/*.toml for [workflows.<name>] keys.
func completeWorkflowNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	home, err := os.UserHomeDir()
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
	dir := getDaemonDir()
	fis, err := os.ReadDir(dir)
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

func mustListDaemonMetaFiles() []string {
	dir := getDaemonDir()
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	horus.CheckErr(err, horus.WithOp("daemon.list"))
	return matches
}

func nameFrom(path string) string {
	return filepath.Base(path[:len(path)-len(".json")])
}

func matchesGroup(metaPath, expectedGroup string) bool {
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

func completeWorkflowGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return availableGroups(), cobra.ShellCompDirectiveDefault
}

func availableGroups() []string {
	files, err := filepath.Glob("/Users/drivas/.lilith/daemon/*.json")
	if err != nil {
		return nil
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

	var result []string
	for g := range groups {
		result = append(result, g)
	}
	return result
}

func mustLoadMeta(path string) daemonMeta {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading metadata from %s: %v\n", path, err)
		os.Exit(1)
	}

	var meta daemonMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON in %s: %v\n", path, err)
		os.Exit(1)
	}

	return meta
}

func sendSignal(pid int, sig syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("could not find process %d: %w", pid, err)
	}
	return proc.Signal(sig)
}

func mustSpawnWatcher(meta daemonMeta) int {
	const op = "lilith.mustSpawnWatcher"
	pid, err := spawnWatcher(&meta)
	horus.CheckErr(err, horus.WithOp(op), horus.WithMessage(fmt.Sprintf("spawning %q", meta.Name)))
	return pid
}

////////////////////////////////////////////////////////////////////////////////////////////////////
