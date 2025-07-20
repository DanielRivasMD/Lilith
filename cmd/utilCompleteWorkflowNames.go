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
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

////////////////////////////////////////////////////////////////////////////////////////////////////
