package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name       string
		level      string
		callInfo   bool
		wantOutput bool
	}{
		{"info-level info prints", "info", true, true},
		{"info-level error prints", "info", false, true},
		{"error-level info suppressed", "error", true, false},
		{"error-level error prints", "error", false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			l := &Logger{
				l:     log.New(buf, "", 0),
				level: parseLevel(tc.level),
			}

			if tc.callInfo {
				l.Info("hello")
			} else {
				l.Error("boom")
			}

			output := buf.String()
			if tc.wantOutput && output == "" {
				t.Fatal("expected output, got none")
			}
			if !tc.wantOutput && output != "" {
				t.Fatalf("expected no output, got %q", output)
			}
			if tc.wantOutput {
				wantPrefix := "INFO:"
				if !tc.callInfo {
					wantPrefix = "ERROR:"
				}
				if !strings.HasPrefix(strings.TrimSpace(output), wantPrefix) {
					t.Fatalf("expected prefix %q, got %q", wantPrefix, output)
				}
			}
		})
	}
}
