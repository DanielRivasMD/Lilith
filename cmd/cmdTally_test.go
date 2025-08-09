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

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DanielRivasMD/Lilith/cmd"
)

type dummyEntry struct {
	name string
}

func (d dummyEntry) Name() string               { return d.name }
func (d dummyEntry) IsDir() bool                { return false }
func (d dummyEntry) Type() os.FileMode          { return 0 }
func (d dummyEntry) Info() (os.FileInfo, error) { return nil, nil }

func Test_runTally_basic(t *testing.T) {
	// Setup dummy daemon directory
	tmp := t.TempDir()

	// Write two daemon .json files
	writeMetaFile(t, tmp, "daemon1", cmd.DaemonMeta{
		Name:      "daemon1",
		Group:     "alpha",
		PID:       os.Getpid(), // current process is guaranteed alive
		InvokedAt: time.Date(2025, 8, 9, 12, 30, 0, 0, time.UTC),
	})
	writeMetaFile(t, tmp, "daemon2", cmd.DaemonMeta{
		Name:      "daemon2",
		Group:     "beta",
		PID:       999999, // unlikely PID, to simulate dead
		InvokedAt: time.Date(2025, 8, 9, 12, 45, 0, 0, time.UTC),
	})

	// Override GetDaemonDir to point to temp
	cmd.GetDaemonDir = func() string { return tmp }

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.RunTally(nil, nil)

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()

	if !strings.Contains(output, "daemon1") || !strings.Contains(output, "daemon2") {
		t.Errorf("Expected output to include both daemons:\n%s", output)
	}
	if !strings.Contains(output, "alive") || !strings.Contains(output, "dead") {
		t.Errorf("Expected status markers in output:\n%s", output)
	}
}

func writeMetaFile(t *testing.T, dir, name string, meta cmd.DaemonMeta) {
	t.Helper()
	path := filepath.Join(dir, name+".json")
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal meta: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write meta file: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
