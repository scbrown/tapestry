package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestWorkspaceCmd_ListEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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

func TestWorkspaceCmd_AddAndList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Add
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "add", "homelab", "/home/braino/gt"})

	if err := root.Execute(); err != nil {
		t.Fatalf("workspace add failed: %v", err)
	}
	if !strings.Contains(buf.String(), "homelab") {
		t.Errorf("workspace add output = %q, want workspace name", buf.String())
	}

	// List
	root2 := newRootCmd("test")
	root2.AddCommand(newWorkspaceCmd())
	buf2 := new(bytes.Buffer)
	root2.SetOut(buf2)
	root2.SetErr(buf2)
	root2.SetArgs([]string{"workspace", "list"})

	if err := root2.Execute(); err != nil {
		t.Fatalf("workspace list failed: %v", err)
	}
	got := buf2.String()
	if !strings.Contains(got, "homelab") {
		t.Errorf("workspace list output = %q, want homelab", got)
	}
	if !strings.Contains(got, "/home/braino/gt") {
		t.Errorf("workspace list output = %q, want path", got)
	}
}

func TestWorkspaceCmd_AddDuplicate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Add first
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "add", "homelab", "/home/braino/gt"})
	if err := root.Execute(); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	// Add duplicate
	root2 := newRootCmd("test")
	root2.AddCommand(newWorkspaceCmd())
	buf2 := new(bytes.Buffer)
	root2.SetOut(buf2)
	root2.SetErr(buf2)
	root2.SetArgs([]string{"workspace", "add", "homelab", "/other/path"})
	if err := root2.Execute(); err == nil {
		t.Fatal("duplicate workspace add should fail")
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

func TestWorkspaceCmd_RemoveExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Add
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "add", "homelab", "/home/braino/gt"})
	if err := root.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Remove
	root2 := newRootCmd("test")
	root2.AddCommand(newWorkspaceCmd())
	buf2 := new(bytes.Buffer)
	root2.SetOut(buf2)
	root2.SetErr(buf2)
	root2.SetArgs([]string{"workspace", "remove", "homelab"})

	if err := root2.Execute(); err != nil {
		t.Fatalf("workspace remove failed: %v", err)
	}
	if !strings.Contains(buf2.String(), "Removed") {
		t.Errorf("workspace remove output = %q, want removal confirmation", buf2.String())
	}

	// Verify removed — list should be empty
	root3 := newRootCmd("test")
	root3.AddCommand(newWorkspaceCmd())
	buf3 := new(bytes.Buffer)
	root3.SetOut(buf3)
	root3.SetErr(buf3)
	root3.SetArgs([]string{"workspace", "list"})

	if err := root3.Execute(); err != nil {
		t.Fatalf("list after remove failed: %v", err)
	}
	if !strings.Contains(buf3.String(), "No workspaces configured") {
		t.Errorf("workspace list after remove = %q, want empty", buf3.String())
	}
}

func TestWorkspaceCmd_RemoveNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := newRootCmd("test")
	root.AddCommand(newWorkspaceCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"workspace", "remove", "nonexistent"})

	if err := root.Execute(); err == nil {
		t.Fatal("removing nonexistent workspace should fail")
	}
}

func TestWorkspaceCmd_Alias(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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
