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
package cmd_test

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"testing"
	"time"

	"github.com/DanielRivasMD/Lilith/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func setupMocks() {
	cmd.FindHomeFn = func(bool) (string, error) {
		return "/mock/home", nil
	}
	cmd.CreateDirFn = func(_ string, _ bool) error {
		return nil
	}
	cmd.SpawnWatcherFn = func(_ *cmd.DaemonMeta) (int, error) {
		return 9999, nil
	}
	cmd.SaveMetaFn = func(_ *cmd.DaemonMeta) error {
		return nil
	}
	cmd.ListMetaFilesFn = func() []string {
		return []string{}
	}
	cmd.LoadMetaFn = func(_ string) cmd.DaemonMeta {
		return cmd.DaemonMeta{}
	}
	cmd.IsDaemonActiveFn = func(_ *cmd.DaemonMeta) bool {
		return false
	}
	cmd.NowFn = func() time.Time {
		return time.Date(2025, 8, 9, 12, 0, 0, 0, time.UTC)
	}
}

func TestPreInvoke_ConfigExists(t *testing.T) {
	setupMocks()

	v := viper.New()
	v.Set("workflows.test.script", "run.sh")
	v.Set("workflows.test.watch", "~/mock")
	// Inject mocked config if needed

	cmd.ConfigName = "test"
	cmdObj := &cobra.Command{}
	cmdObj.Flags().String("name", "", "")
	cmdObj.Flags().String("group", "", "")
	cmdObj.Flags().String("log", "", "")
	cmdObj.Flags().String("watch", "", "")
	cmdObj.Flags().String("script", "", "")
	cmdObj.Flags().String("config", "", "")

	err := cmd.PreInvoke(cmdObj, []string{})
	assert.NoError(t, err)
}

func TestRunInvoke_Success(t *testing.T) {
	setupMocks()

	cmd.WatchDir = "~/mock"
	cmd.ScriptPath = "run.sh"
	cmd.LogName = "testd"
	cmd.DaemonName = "testd"
	cmd.GroupName = "alpha"

	cmd.RunInvoke(&cobra.Command{}, []string{})
}

////////////////////////////////////////////////////////////////////////////////////////////////////
