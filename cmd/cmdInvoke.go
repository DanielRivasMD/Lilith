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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// invocation flags + derived values
var (
	daemonName string // instance name, defaults to configName
	configName string // workflow key
	watchDir   string
	scriptPath string
	logName    string

	groupName string // derived from TOML filename
)

////////////////////////////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////////////////////////////

// invokeCmd starts a watcher using settings from ~/.lilith/config/*.toml
var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Start a new watcher daemon",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// 1) Load every TOML in ~/.lilith/config
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cfgDir := filepath.Join(home, ".lilith", "config")
		var (
			foundV      *viper.Viper
			cfgFileUsed string
		)

		fis, err := os.ReadDir(cfgDir)
		if err != nil {
			return err
		}
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
			if v.IsSet("workflows." + configName) {
				foundV = v
				cfgFileUsed = path
				break
			}
		}
		if foundV == nil {
			return fmt.Errorf("workflow %q not found in %s/*.toml", configName, cfgDir)
		}

		// 2) Default daemonName ← configName if none provided
		if daemonName == "" {
			daemonName = configName
			cmd.Flags().Set("name", daemonName)
		}

		// 3) Derive groupName from the TOML file basename
		base := filepath.Base(cfgFileUsed)                       // e.g. "forge.toml"
		groupName = strings.TrimSuffix(base, filepath.Ext(base)) // e.g. "forge"
		cmd.Flags().Set("group", groupName)

		// 4) Bind flags from that workflow block
		wf := foundV.Sub("workflows." + configName)
		bindFlag(cmd, "watch", &watchDir, wf)
		bindFlag(cmd, "script", &scriptPath, wf)
		bindFlag(cmd, "log", &logName, wf)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		const op = "lilith.invoke"

		// 5) Validate
		// (name now always set after PreRunE)
		if watchDir == "" {
			horus.CheckErr(
				fmt.Errorf("`--watch` is required"),
				horus.WithOp(op), horus.WithMessage("provide a directory to watch"),
			)
		}
		if scriptPath == "" {
			horus.CheckErr(
				fmt.Errorf("`--script` is required"),
				horus.WithOp(op), horus.WithMessage("provide a script to run"),
			)
		}
		if logName == "" {
			horus.CheckErr(
				fmt.Errorf("`--log` is required"),
				horus.WithOp(op), horus.WithMessage("provide a log name"),
			)
		}

		// 6) Expand env vars / tilde
		var err error
		watchDir, err = expandPath(watchDir)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("expanding watch path"))

		scriptPath, err = expandPath(scriptPath)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("expanding script path"))

		// 7) Ensure ~/.lilith/logs exists
		home, err := os.UserHomeDir()
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("getting home directory"))

		logDir := filepath.Join(home, ".lilith", "logs")
		horus.CheckErr(
			domovoi.CreateDir(logDir),
			horus.WithOp(op), horus.WithMessage(fmt.Sprintf("creating %q", logDir)),
		)
		logPath := filepath.Join(logDir, logName+".log")

		// 8) Build & persist metadata
		meta := &daemonMeta{
			Name:       daemonName,
			Group:      groupName,
			WatchDir:   watchDir,
			ScriptPath: scriptPath,
			LogPath:    logPath,
			InvokedAt:  time.Now(),
		}

		pid, err := spawnWatcher(meta)
		horus.CheckErr(err, horus.WithOp(op), horus.WithMessage("starting watcher"))
		meta.PID = pid
		horus.CheckErr(saveMeta(meta), horus.WithOp(op), horus.WithMessage("writing metadata"))

		// 9) Done
		fmt.Printf(
			"%s invoked daemon %q (group=%q) with PID %d\n",
			chalk.Green.Color("OK:"), daemonName, groupName, pid,
		)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(invokeCmd)

	invokeCmd.Flags().StringVarP(&daemonName, "name", "n", "", "Unique daemon name (defaults to --config)")
	invokeCmd.Flags().StringVarP(&configName, "config", "c", "", "Workflow to apply")
	invokeCmd.Flags().StringVarP(&groupName, "group", "g", "", "Watcher group name (overrides TOML)")
	invokeCmd.Flags().StringVarP(&watchDir, "watch", "w", "", "Directory to watch")
	invokeCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "Script to execute on change")
	invokeCmd.Flags().StringVarP(&logName, "log", "l", "", "Name for log file (no `.log` extension)")

	invokeCmd.RegisterFlagCompletionFunc("config", completeWorkflowNames)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
