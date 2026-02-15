package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestWorkspaceCmd_List(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("workspace list failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "No workspaces configured") {
		t.Errorf("workspace list output = %q, want empty state message", got)
	}
}

func TestWorkspaceCmd_Add(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "add", "homelab", "/home/braino/gt"})

	if err := root.Execute(); err != nil {
		t.Fatalf("workspace add failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "homelab") {
		t.Errorf("workspace add output = %q, want workspace name", got)
	}
}

func TestWorkspaceCmd_AddMissingArgs(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "add", "homelab"})

	err := root.Execute()
	if err == nil {
		t.Fatal("workspace add with 1 arg should fail, got nil")
	}
}

func TestWorkspaceCmd_Remove(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "remove", "homelab"})

	if err := root.Execute(); err != nil {
		t.Fatalf("workspace remove failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Removed") {
		t.Errorf("workspace remove output = %q, want removal confirmation", got)
	}
}

func TestWorkspaceCmd_Alias(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"ws", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("ws alias failed: %v", err)
	}
}
