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
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	verbose bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// daemonMeta holds persistent info about a watcher process
type daemonMeta struct {
	Name       string    `json:"name"`
	Group      string    `json:"group"`
	WatchDir   string    `json:"watchDir"`
	ScriptPath string    `json:"scriptPath"`
	LogPath    string    `json:"logPath"`
	PID        int       `json:"pid"`
	InvokedAt  time.Time `json:"invokedAt"`
}

// getDaemonDir returns ~/.lilith/daemon
func getDaemonDir() string {
	return filepath.Join(os.Getenv("HOME"), ".lilith", "daemon")
}

// saveMeta writes meta to ~/.lilith/daemon/<name>.json
func saveMeta(meta *daemonMeta) error {
	dir := getDaemonDir()
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
	if err := os.MkdirAll(filepath.Dir(meta.LogPath), 0755); err != nil {
		return 0, err
	}

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

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use: "lilith",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Blue.Color("lilith") + `, manage background watcher daemon
`,
	Example: chalk.White.Color("lilith") + ` ` + chalk.Bold.TextStyle(chalk.White.Color("help")),
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostic output")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
