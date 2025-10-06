package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func Test_mainRun(t *testing.T) {
	origStdin := os.Stdin
	os.Stdin = nil
	defer func() {
		os.Stdin = origStdin
	}()

	inputFile, err := os.CreateTemp("", "go-readability-test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inputFile.WriteString("<h1>Hello World</h1><p>This is an article.</p>"); err != nil {
		t.Fatal(err)
	}
	inputFile.Close()

	gotStdout, gotStderr := captureOutput(func() {
		err := mainRun([]string{"go-readability-test", "-f", inputFile.Name()})
		if err != nil {
			t.Error(err)
		}
	})

	expectedStdout := "<div id=\"readability-page-1\" class=\"page\"><h2>Hello World</h2><p>This is an article.</p></div>\n"
	expectedStderr := "warning: the page might not have article contents\n"

	if gotStdout != expectedStdout {
		t.Errorf("expected stdout %q, got %q", expectedStdout, gotStdout)
	}
	if gotStderr != expectedStderr {
		t.Errorf("expected stderr %q, got %q", expectedStderr, gotStderr)
	}
}

func captureOutput(f func()) (stdout string, stderr string) {
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)

	// Read stdout
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	// Read stderr
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	f() // Run the code

	wOut.Close()
	wErr.Close()
	stdout = <-outC
	stderr = <-errC

	return
}
