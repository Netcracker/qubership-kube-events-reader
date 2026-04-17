package test

import (
	"os"
	"testing"
)

func TestChangeStdoutToFileAndBack(t *testing.T) {
	initialStdout := os.Stdout

	fname, err := ChangeStdoutToFile("codex-test-stdout")
	if err != nil {
		t.Fatalf("unexpected error switching stdout: %v", err)
	}

	if _, err = os.Stdout.WriteString("hello"); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if err = ChangeFileToStdout(initialStdout); err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}
	if os.Stdout != initialStdout {
		t.Fatal("expected stdout to be restored")
	}
	if _, err = os.Stat(fname); !os.IsNotExist(err) {
		t.Fatalf("expected temporary stdout file to be removed, stat err=%v", err)
	}
}
