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
var (
	lilithDir string
	configDir string
	logDir    string
	daemonDir string
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostics")
	cobra.OnInitialize(initConfigPaths)

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

func initConfigPaths() {
	var err error
	home, err = domovoi.FindHome(verbose)
	horus.CheckErr(err, horus.WithCategory("init_error"), horus.WithMessage("getting home directory"))
	lilithDir = filepath.Join(home, ".lilith")
	configDir = filepath.Join(lilithDir, "config")
	logDir = filepath.Join(lilithDir, "logs")
	daemonDir = filepath.Join(lilithDir, "daemon")
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
	// ensure the daemon directory exists
	if err := domovoi.CreateDir(daemonDir, verbose); err != nil {
		horus.CheckErr(
			err,
			horus.WithOp(op),
			horus.WithCategory("io_error"),
			horus.WithMessage("creating daemon directory"),
			horus.WithDetails(map[string]any{
				"dir": daemonDir,
			}),
		)
	}

	// marshal the metadata
	data, err := json.MarshalIndent(meta, "", "  ")
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("encode_error"),
		horus.WithMessage("marshaling metadata"),
		horus.WithDetails(map[string]any{
			"name": meta.Name,
		}),
	)

	// write the file
	path := filepath.Join(daemonDir, meta.Name+".json")
	horus.CheckErr(
		os.WriteFile(path, data, 0o644),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("writing metadata file"),
		horus.WithDetails(map[string]any{
			"path": path,
		}),
	)
}

// loadMeta reads ~/.lilith/daemon/<name>.json
func loadMeta(name string) *daemonMeta {
	const op = "daemon.loadMeta"
	path := resolveMetaPath(name)

	// read the file
	data, err := os.ReadFile(path)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("reading metadata file"),
		horus.WithDetails(map[string]any{
			"path": path,
			"name": name,
		}),
	)

	// unmarshal into struct
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

// resolveMetaPath will turn any of these into the correct file to read:
//
//	"foo"          → /home/me/.lilith/daemon/foo.json
//	"foo.json"     → /home/me/.lilith/daemon/foo.json
//	"/tmp/bar.json" → /tmp/bar.json
//	"sub/dir/baz"  → sub/dir/baz.json (relative path)
//
// check if the literal name exists first, otherwise falls back to daemonDir
func resolveMetaPath(name string) string {
	// 1) if the user passed an absolute or relative path that actually exists, use it
	if fi, err := os.Stat(name); err == nil && !fi.IsDir() {
		return name
	}

	// 2) ensure we have a .json extension
	if filepath.Ext(name) != ".json" {
		name = name + ".json"
	}

	// 3) join with the default daemonDir
	return filepath.Join(daemonDir, name)
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
		horus.WithCategory("daemon_error"),
		horus.WithMessage("listing daemon metadata files"),
		horus.WithDetails(map[string]any{
			"pattern": daemonPattern,
		}),
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

	// ensure the log directory exists
	horus.CheckErr(
		domovoi.CreateDir(logDir, verbose),
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage("creating log directory"),
		horus.WithDetails(map[string]any{
			"logDir": logDir,
		}),
	)

	// build watcher command
	cmd := exec.Command(
		"watchexec",
		"--watch", meta.WatchDir,
		"--",
		"bash", meta.ScriptPath,
	)

	// open (or create) the log file
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

	// start the watcher process
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

	// detach from parent
	horus.CheckErr(
		cmd.Process.Release(),
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage("releasing watcher process"),
		horus.WithDetails(map[string]any{
			"pid": cmd.Process.Pid,
		}),
	)

	return cmd.Process.Pid
}

// sendDaemonSignal finds the process and sends it the given signal
func sendDaemonSignal(pid int, sig syscall.Signal) {
	const op = "daemon.sendSignal"

	proc, err := os.FindProcess(pid)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage(fmt.Sprintf("finding process %d", pid)),
		horus.WithDetails(map[string]any{
			"pid": pid,
		}),
	)

	horus.CheckErr(
		proc.Signal(sig),
		horus.WithOp(op),
		horus.WithCategory("spawn_error"),
		horus.WithMessage(fmt.Sprintf("sending signal %s to pid %d", sig, pid)),
		horus.WithDetails(map[string]any{
			"pid": pid,
			"sig": sig,
		}),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// completeWorkflowNames scans ~/.lilith/config/*.toml for [workflows.<name>] keys
func completeWorkflowNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fis, err := os.ReadDir(configDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	seen := map[string]struct{}{}
	for _, fi := range fis {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".toml") {
			continue
		}
		path := filepath.Join(configDir, fi.Name())
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
