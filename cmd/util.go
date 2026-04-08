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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// onelineErr returns a bold red error string (used in some formatters)
func onelineErr(er string) string {
	return chalk.Bold.TextStyle(chalk.Red.Color(er))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// bindFlag sets a flag from a viper config if not already changed
func bindFlag(cmd *cobra.Command, flagName string, cfg *viper.Viper) {
	const op = "daemon.bindFlag"
	flags := cmd.Flags()

	if flags.Changed(flagName) || !cfg.IsSet(flagName) {
		return
	}

	f := flags.Lookup(flagName)
	if f == nil {
		return
	}

	var raw string
	switch f.Value.Type() {
	case "string":
		raw = cfg.GetString(flagName)
	case "int":
		raw = strconv.Itoa(cfg.GetInt(flagName))
	case "bool":
		raw = strconv.FormatBool(cfg.GetBool(flagName))
	default:
		raw = cfg.GetString(flagName)
	}

	if err := flags.Set(flagName, raw); err != nil {
		horus.CheckErr(
			horus.NewCategorizedHerror(
				op,
				"flag_error",
				fmt.Sprintf("setting %q from config", flagName),
				err,
				map[string]any{
					"flag":  flagName,
					"value": raw,
				},
			),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// saveMeta writes meta to ~/.lilith/daemon/<name>.json
func saveMeta(meta *daemonMeta) {
	const op = "daemon.saveMeta"

	horus.CheckErr(
		domovoi.CreateDir(configDirs.daemon, rootFlags.verbose),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("creating daemon directory"),
		horus.WithDetails(map[string]any{
			"dir": configDirs.daemon,
		}),
	)

	// Convert time to string for JSON
	type metaExport struct {
		Daemon     string `json:"name"`
		Group      string `json:"group"`
		WatchDir   string `json:"watchDir"`
		ScriptPath string `json:"scriptPath"`
		LogPath    string `json:"logPath"`
		PID        int    `json:"pid"`
		InvokedAt  string `json:"invokedAt"`
	}
	export := metaExport{
		Daemon:     meta.Daemon,
		Group:      meta.Group,
		WatchDir:   meta.WatchDir,
		ScriptPath: meta.ScriptPath,
		LogPath:    meta.LogPath,
		PID:        meta.PID,
		InvokedAt:  meta.InvokedAt.Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(export, "", "  ")
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("encode_error"),
		horus.WithMessage("marshaling metadata"),
		horus.WithDetails(map[string]any{
			"name": meta.Daemon,
		}),
	)

	path := filepath.Join(configDirs.daemon, meta.Daemon+".json")
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

	var export struct {
		Daemon     string `json:"name"`
		Group      string `json:"group"`
		WatchDir   string `json:"watchDir"`
		ScriptPath string `json:"scriptPath"`
		LogPath    string `json:"logPath"`
		PID        int    `json:"pid"`
		InvokedAt  string `json:"invokedAt"`
	}
	horus.CheckErr(
		json.Unmarshal(data, &export),
		horus.WithOp(op),
		horus.WithCategory("decode_error"),
		horus.WithMessage("unmarshaling metadata"),
		horus.WithDetails(map[string]any{
			"path": path,
			"name": name,
		}),
	)

	invoked, _ := time.Parse(time.RFC3339, export.InvokedAt) // ignore error, default to zero
	return &daemonMeta{
		Daemon:     export.Daemon,
		Group:      export.Group,
		WatchDir:   export.WatchDir,
		ScriptPath: export.ScriptPath,
		LogPath:    export.LogPath,
		PID:        export.PID,
		InvokedAt:  invoked,
	}
}

// resolveMetaPath returns the full path to a daemon metadata file
func resolveMetaPath(name string) string {
	if fi, err := os.Stat(name); err == nil && !fi.IsDir() {
		return name
	}
	if filepath.Ext(name) != ".json" {
		name = name + ".json"
	}
	return filepath.Join(configDirs.daemon, name)
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
	daemonPattern := filepath.Join(configDirs.daemon, "*.json")
	daemonMatches, err := filepath.Glob(daemonPattern)
	horus.CheckErr(
		err,
		horus.WithOp(op),
		horus.WithCategory("daemon_error"),
		horus.WithMessage("listing daemon metadata files"),
		horus.WithDetails(map[string]any{
			"pattern": daemonPattern,
		}),
	)
	return daemonMatches
}

func matchDaemonGroup(metaPath, expectedGroup string) bool {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return false
	}
	var meta struct {
		Group string `json:"group"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}
	return meta.Group == expectedGroup
}

////////////////////////////////////////////////////////////////////////////////////////////////////

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
	proc.Signal(sig) // ignore error if process already gone
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// completion functions
func completeWorkflowNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fis, err := os.ReadDir(configDirs.config)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	seen := map[string]struct{}{}
	for _, fi := range fis {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".toml") {
			continue
		}
		path := filepath.Join(configDirs.config, fi.Name())
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

func completeDaemonNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fis, err := os.ReadDir(configDirs.daemon)
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

func completeWorkflowGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	files, err := filepath.Glob(filepath.Join(configDirs.daemon, "*.json"))
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
