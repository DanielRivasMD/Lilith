package cmd_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/DanielRivasMD/Lilith/cmd"
	"github.com/spf13/cobra"
)

func setupFreezeMocks(t *testing.T) {
	t.Helper()

	// Save original seams for cleanup
	origLoadMetaFn := cmd.LoadMetaFn
	origMustListFn := cmd.MustListDaemonMetaFilesFn
	origMustLoadMetaFn := cmd.MustLoadMetaFn
	origMatchesGroupFn := cmd.MatchesGroupFn
	origSendSignalFn := cmd.SendSignalFn

	// Restore them after test
	t.Cleanup(func() {
		cmd.LoadMetaFn = origLoadMetaFn
		cmd.MustListDaemonMetaFilesFn = origMustListFn
		cmd.MustLoadMetaFn = origMustLoadMetaFn
		cmd.MatchesGroupFn = origMatchesGroupFn
		cmd.SendSignalFn = origSendSignalFn
	})

	// Mock single daemon metadata (used with --name)
	cmd.LoadMetaFn = func(name string) (*cmd.DaemonMeta, error) {
		return &cmd.DaemonMeta{
			Name:      name,
			Group:     "mock-group",
			PID:       os.Getpid(),
			InvokedAt: time.Now(),
		}, nil
	}

	// Mock group/all daemon metadata files
	cmd.MustListDaemonMetaFilesFn = func() []string {
		return []string{"daemonA.json", "daemonB.json"}
	}

	cmd.MustLoadMetaFn = func(path string) *cmd.DaemonMeta {
		return &cmd.DaemonMeta{
			Name:      filepath.Base(path),
			Group:     "mock-group",
			PID:       os.Getpid(),
			InvokedAt: time.Now(),
		}
	}

	// Match daemons by group
	cmd.MatchesGroupFn = func(path, group string) bool {
		return group == "mock-group"
	}

	// Simulate signal sending
	cmd.SendSignalFn = func(pid int, sig syscall.Signal) error {
		return nil
	}
}

// Helper to capture output from stdout
func captureOutput(t *testing.T, run func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	run()

	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func Test_runFreeze_SingleDaemon(t *testing.T) {
	setupFreezeMocks(t)

	output := captureOutput(t, func() {
		cmdObj := &cobra.Command{}
		cmd.RunFreeze(cmdObj, []string{"daemonX"})
	})

	if !strings.Contains(output, "froze daemon") || !strings.Contains(output, "daemonX") {
		t.Errorf("Expected freeze confirmation for daemonX:\n%s", output)
	}
}

func Test_runFreeze_GroupFlag(t *testing.T) {
	setupFreezeMocks(t)

	output := captureOutput(t, func() {
		cmdObj := &cobra.Command{}
		cmdObj.Flags().String("group", "mock-group", "")
		_ = cmdObj.Flags().Set("group", "mock-group")
		cmd.RunFreeze(cmdObj, []string{})
	})

	if !strings.Contains(output, "daemonA") || !strings.Contains(output, "daemonB") {
		t.Errorf("Expected group freeze output for daemonA and daemonB:\n%s", output)
	}
}

func Test_runFreeze_AllFlag(t *testing.T) {
	setupFreezeMocks(t)

	output := captureOutput(t, func() {
		cmdObj := &cobra.Command{}
		cmdObj.Flags().Bool("all", false, "")
		_ = cmdObj.Flags().Set("all", "true")
		cmd.RunFreeze(cmdObj, []string{})
	})

	if !strings.Contains(output, "froze daemon") || !strings.Contains(output, "daemonA") {
		t.Errorf("Expected global freeze output for daemonA:\n%s", output)
	}
}
