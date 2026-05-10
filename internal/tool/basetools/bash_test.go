package basetools

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"Amadeus/internal/skill"
)

func TestLoadRegistersCmdTool(t *testing.T) {
	tools := Load(skill.Config{})

	names := make(map[string]bool, len(tools))
	for _, item := range tools {
		info, err := item.Info(context.Background())
		if err != nil {
			t.Fatalf("load tool info: %v", err)
		}
		names[info.Name] = true
	}

	for _, name := range []string{"bash", "cmd", "load_skill"} {
		if !names[name] {
			t.Fatalf("expected tool %q to be registered; got %v", name, names)
		}
	}
}

func TestCmdToolMetadata(t *testing.T) {
	info, err := GetCmdTool().Info(context.Background())
	if err != nil {
		t.Fatalf("load cmd tool info: %v", err)
	}

	if info.Name != "cmd" {
		t.Fatalf("expected cmd tool name, got %q", info.Name)
	}
	if !strings.Contains(info.Desc, "cmd.exe") {
		t.Fatalf("expected cmd description to mention cmd.exe, got %q", info.Desc)
	}
}

func TestRunCmdReturnsUnsupportedPlatformError(t *testing.T) {
	originalGOOS := runtimeGOOS
	runtimeGOOS = "linux"
	t.Cleanup(func() {
		runtimeGOOS = originalGOOS
	})

	_, err := runCmd(context.Background(), map[string]interface{}{
		"command": "echo should-not-run",
	})
	if err == nil {
		t.Fatal("expected unsupported platform error")
	}
	if !strings.Contains(err.Error(), "only supported on Windows") {
		t.Fatalf("expected unsupported platform error, got %v", err)
	}
}

func TestRunCmdExecutesCommandOnWindows(t *testing.T) {
	if runtimeGOOS != "windows" {
		t.Skip("cmd execution is Windows-only")
	}

	output, err := runCmd(context.Background(), map[string]interface{}{
		"command": "echo amadeus-cmd",
	})
	if err != nil {
		t.Fatalf("run cmd: %v\n%s", err, output)
	}

	if !strings.Contains(output, "command: echo amadeus-cmd") {
		t.Fatalf("expected command in output, got:\n%s", output)
	}
	if !strings.Contains(output, "exit_code: 0") {
		t.Fatalf("expected exit code in output, got:\n%s", output)
	}
	if !strings.Contains(output, "amadeus-cmd") {
		t.Fatalf("expected stdout in output, got:\n%s", output)
	}
}

func TestRunCmdUsesWorkdirOnWindows(t *testing.T) {
	if runtimeGOOS != "windows" {
		t.Skip("cmd execution is Windows-only")
	}

	workdir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}

	output, err := runCmd(context.Background(), map[string]interface{}{
		"command": "cd",
		"workdir": workdir,
	})
	if err != nil {
		t.Fatalf("run cmd: %v\n%s", err, output)
	}

	if !strings.Contains(output, "workdir: "+workdir) {
		t.Fatalf("expected workdir in output, got:\n%s", output)
	}
	if !strings.Contains(strings.ToLower(output), strings.ToLower(workdir)) {
		t.Fatalf("expected command to run in workdir, got:\n%s", output)
	}
}

func TestRunCmdRejectsInvalidTimeout(t *testing.T) {
	if runtimeGOOS != "windows" {
		t.Skip("cmd timeout validation is reached after Windows platform validation")
	}

	_, err := runCmd(context.Background(), map[string]interface{}{
		"command":         "echo amadeus-cmd",
		"timeout_seconds": 0,
	})
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}
	if !strings.Contains(err.Error(), "timeout_seconds") {
		t.Fatalf("expected timeout validation error, got %v", err)
	}
}
