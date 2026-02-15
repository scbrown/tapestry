package cli

import (
	"bytes"
	"testing"
)

func TestRootCmd_NoArgs(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newServeCmd(), newConfigCmd(), newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{})

	if err := root.Execute(); err != nil {
		t.Fatalf("root command failed: %v", err)
	}
}

func TestRootCmd_Version(t *testing.T) {
	root := newRootCmd("1.2.3")
	root.AddCommand(newServeCmd(), newConfigCmd(), newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("version flag failed: %v", err)
	}

	got := buf.String()
	if want := "1.2.3"; !bytes.Contains([]byte(got), []byte(want)) {
		t.Errorf("version output = %q, want to contain %q", got, want)
	}
}

func TestRootCmd_Help(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newServeCmd(), newConfigCmd(), newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("help flag failed: %v", err)
	}

	got := buf.String()
	for _, want := range []string{"serve", "config", "workspace"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Errorf("help output missing %q subcommand", want)
		}
	}
}
