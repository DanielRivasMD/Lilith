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

	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DanielRivasMD/Lilith/cmd"
	"github.com/DanielRivasMD/domovoi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func TestSaveAndLoadMeta(t *testing.T) {
	meta := &cmd.DaemonMeta{
		Name:       "testd",
		Group:      "alpha",
		WatchDir:   "/tmp",
		ScriptPath: "/bin/true",
		LogPath:    "/tmp/testd.log",
		PID:        123,
		InvokedAt:  time.Now(),
	}
	dir := cmd.GetDaemonDir()
	_ = os.Remove(filepath.Join(dir, meta.Name+".json")) // cleanup

	err := cmd.SaveMetaFn(meta)
	assert.NoError(t, err)

	read, err := cmd.LoadMetaFn("testd")
	assert.NoError(t, err)
	assert.Equal(t, meta.Name, read.Name)
}

func TestExpandPath(t *testing.T) {
	home, err := domovoi.FindHome(false)
	assert.NoError(t, err)

	path := "~/dummy"
	expanded, err := cmd.ExpandPathFn(path)
	assert.NoError(t, err)
	assert.True(t, filepath.HasPrefix(expanded, home))
}

func TestBindFlag(t *testing.T) {
	v := viper.New()
	v.Set("magic", "42")
	var result string

	cmdObj := &cobra.Command{}
	cmdObj.Flags().String("magic", "", "")
	cmd.BindFlag(cmdObj, "magic", &result, v)

	assert.Equal(t, "42", result)
}

func TestMustListDaemonMetaFiles(t *testing.T) {
	files := cmd.MustListDaemonMetaFilesFn()
	assert.NotNil(t, files)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
