package main

import (
	"os"
	"testing"
)

func TestMainApp(t *testing.T) {
	// Create a temporary htempl file
	tmpDir, err := os.MkdirTemp("", "htempl-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	srcFile := "test.htempl"
	err = os.WriteFile(srcFile, []byte("---\nvar: value\n...\n{{.var}}"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Save original args and restore later
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test direct file processing
	os.Args = []string{"htempl", "-dest", ".", srcFile}
	if err := run(); err != nil {
		t.Errorf("run() error = %v", err)
	}

	destFile := "test.html"
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Errorf("expected %s to be created", destFile)
	}

	// Test directory walking
	*src = "."
	*dest = "."
	*verbose = true
	os.Args = []string{"htempl"}
	if err := run(); err != nil {
		t.Errorf("run() error = %v", err)
	}
}

func TestRunErrors(t *testing.T) {
	t.Run("WalkError", func(t *testing.T) {
		*src = "non_existent_dir"
		if err := run(); err == nil {
			t.Error("expected error for non-existent src dir")
		}
	})
	t.Run("ProcessFileError", func(t *testing.T) {
		*src = ""
		os.Args = []string{"htempl", "non_existent_file"}
		if err := run(); err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestMainSuccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "htempl-test-main")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("test.htempl", []byte("Body"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"htempl", "test.htempl"}
	main()

	if _, err := os.Stat("test.html"); os.IsNotExist(err) {
		t.Error("expected test.html to exist")
	}
}

func TestProcessFileErrors(t *testing.T) {
	t.Run("CreateError", func(t *testing.T) {
		// This should fail because the directory doesn't exist
		err := processFile("non_existent_dir/test.htempl")
		if err == nil {
			t.Error("expected error")
		}
	})
	t.Run("ExecuteError", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "htempl-exec-err")
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()
		fname := "test.htempl"
		cwd, _ := os.Getwd()
		defer func() {
			_ = os.Chdir(cwd)
		}()
		err := os.Chdir(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fname, []byte(`{{map "only_one"}}`), 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = processFile(fname)
		if err == nil {
			t.Error("expected execution error")
		}
	})
}
