package rclone

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestListFolders(t *testing.T) {
	root := t.TempDir()

	err := os.Mkdir(filepath.Join(root, "beta"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(root, "alpha"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(
		filepath.Join(root, "alpha", "nested"),
		0755,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(root, "file.txt"),
		[]byte("not a directory"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ListFolders(root)
	if err != nil {
		t.Fatalf("ListFolders() error = %v", err)
	}

	want := []string{
		"alpha",
		"beta",
	}

	if !slices.Equal(got, want) {
		t.Errorf("ListFolders() = %v, want %v", got, want)
	}
}

func TestListFoldersEmpty(t *testing.T) {
	root := t.TempDir()

	got, err := ListFolders(root)
	if err != nil {
		t.Fatalf("ListFolders() error = %v", err)
	}

	if got != nil {
		t.Errorf("ListFolders() = %v, want nil", got)
	}
}

func TestListFoldersRootDoesNotExist(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")

	got, err := ListFolders(root)

	if err == nil {
		t.Fatal("ListFolders() error = nil, want error")
	}

	if got != nil {
		t.Errorf("ListFolders() = %v, want nil", got)
	}
}
