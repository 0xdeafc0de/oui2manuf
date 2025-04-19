package ouidb

import (
	"os"
	"path/filepath"
	_ "strings"
	"testing"
)

const sampleData = `
# Comment
00:00:00            00:00:00        Xerox Corporation
FC:D2:B6:00/28      CgPowerAndIn    Cg Power And Industrial Solutions Ltd
FC:D2:B6:10/28      Link            Link (Far-East) Corporation
FC:D2:B6:20/28      Soma            Soma GmbH
8C:1F:64:DC:60/36   R&K             R&K
`

func writeTempManuf(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "manuf.db")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test manuf file: %v", err)
	}
	return path
}

func TestUpdateDatabase(t *testing.T) {
	path := writeTempManuf(t, sampleData)
	err := UpdateDatabase(path)
	if err != nil {
		t.Fatalf("UpdateDatabase failed: %v", err)
	}

	if len(ouiData) != 5 {
		t.Errorf("expected 5 entries, got %d", len(ouiData))
	}

	expectedKeys := []string{
		"00:00:00",
		"FC:D2:B6:0",
		"FC:D2:B6:1",
		"FC:D2:B6:2",
		"8C:1F:64:DC:6",
	}

	for _, k := range expectedKeys {
		if _, ok := ouiData[k]; !ok {
			t.Errorf("missing expected key: %s", k)
		}
	}
}

func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestLookup(t *testing.T) {
	path := writeTempManuf(t, sampleData)
	err := UpdateDatabase(path)
	if err != nil {
		t.Fatalf("UpdateDatabase failed: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"00:00:00:AA:BB:CC", "00:00:00"},
		{"fc:d2:b6:20:11:22", "Soma"},
		{"8C:1F:64:DC:60:FF", "R&K"},
		{"FC:D2:B6:00:10:20", "CgPowerAndIn"},
		{"FC:D2:B6:10", "Link"},
		{"fc:d2:b2:30", ""}, // Not in test data
		{"invalid-mac", ""},
	}

	for _, tt := range tests {
		//t.Logf("Keys in ouiData: %v", strings.Join(getMapKeys(ouiData), ", "))
		got, err := Lookup(tt.input)
		if err != nil && tt.want != "" {
			t.Errorf("Lookup(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("Lookup(%q) = %q; want %q", tt.input, got, tt.want)
			continue
		}
		if err != nil && tt.want == "" {
			t.Logf("Expected failure for %q: %v", tt.input, err)
		} else {
			t.Logf("Lookup(%q) returned: %q", tt.input, got)
		}
	}
}

func TestUpdateDatabase_FileError(t *testing.T) {
	err := UpdateDatabase("nonexistentfile.db")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLookup_FileMissingTriggersDownloadFail(t *testing.T) {
	// Simulate missing data directory
	oldDir := dataDir
	defer func() { os.MkdirAll(oldDir, 0755) }() // Restore directory

	os.RemoveAll(dataDir)
	_, err := Lookup("00:00:00:00:00:00")
	if err == nil {
		t.Error("expected error when manuf file is missing")
	}
}
