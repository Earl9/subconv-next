package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
)

func TestExportInspectAndRestore(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "etc", "config", "subconv-next")
	dataDir := filepath.Join(root, "mnt", "data")
	mustWrite(t, configPath, "config subconv-next 'main'\n\toption port '9876'\n\toption data_dir '/mnt/data'\n")
	mustWrite(t, filepath.Join(dataDir, "state.json"), `{"ok":true}`)
	mustWrite(t, filepath.Join(dataDir, "logs", "app.log"), "excluded")
	mustWrite(t, filepath.Join(dataDir, "cache", "item"), "excluded")
	archivePath := filepath.Join(root, "backup.tar.gz")

	// Exercise the production path rules through a temporary bind-like symlink-free fixture.
	original := safeDataDir
	safeDataDir = regexpForTest(root)
	t.Cleanup(func() { safeDataDir = original })
	result, err := Export(ExportOptions{ConfigPath: configPath, DataDir: dataDir, OutputPath: archivePath, AppVersion: "test"})
	if err != nil {
		t.Fatalf("Export() error: %v", err)
	}
	if result.Size == 0 || result.SHA256 == "" {
		t.Fatalf("Export() result incomplete: %+v", result)
	}
	manifest, err := Inspect(archivePath)
	if err != nil {
		t.Fatalf("Inspect() error: %v", err)
	}
	if len(manifest.Files) != 2 {
		t.Fatalf("manifest files = %d, want config and state only", len(manifest.Files))
	}

	mustWrite(t, configPath, "config subconv-next 'main'\n\toption port '1234'\n\toption data_dir '"+dataDir+"'\n")
	mustWrite(t, filepath.Join(dataDir, "state.json"), "changed")
	initScript := filepath.Join(root, "init")
	mustWriteMode(t, initScript, "#!/bin/sh\ncase \"$1\" in running) exit 1;; *) exit 0;; esac\n", 0755)
	restored, err := Restore(RestoreOptions{ArchivePath: archivePath, ConfigPath: configPath, DataDir: dataDir, InitScript: initScript, LockPath: filepath.Join(root, "restore.lock")})
	if err != nil {
		t.Fatalf("Restore() error: %v", err)
	}
	if restored.RestoredFiles != 2 {
		t.Fatalf("restored files = %d, want 2", restored.RestoredFiles)
	}
	state, _ := os.ReadFile(filepath.Join(dataDir, "state.json"))
	if string(state) != `{"ok":true}` {
		t.Fatalf("restored state = %q", state)
	}
	config, _ := os.ReadFile(configPath)
	if !strings.Contains(string(config), "option port '9876'") || !strings.Contains(string(config), "option data_dir '"+dataDir+"'") {
		t.Fatalf("restored config = %q", config)
	}
}

func TestInspectRejectsUnsafeArchives(t *testing.T) {
	tests := []struct {
		name     string
		entries  []testEntry
		manifest Manifest
	}{
		{name: "traversal", entries: []testEntry{{name: "../etc/passwd", body: "x"}}},
		{name: "absolute", entries: []testEntry{{name: "/etc/passwd", body: "x"}}},
		{name: "symlink", entries: []testEntry{{name: "data/link", body: "", typeflag: tar.TypeSymlink, linkname: "/etc"}}},
		{name: "hardlink", entries: []testEntry{{name: "data/link", body: "", typeflag: tar.TypeLink, linkname: "config/subconv-next"}}},
		{name: "unsupported version", entries: []testEntry{{name: "config/subconv-next", body: "config subconv-next 'main'\n"}}, manifest: Manifest{FormatVersion: 99, Application: "subconv-next", AppVersion: "test", CreatedAt: "now"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := filepath.Join(t.TempDir(), "bad.tar.gz")
			writeTestArchive(t, archivePath, tt.entries, tt.manifest)
			if _, err := Inspect(archivePath); err == nil {
				t.Fatal("Inspect() accepted unsafe archive")
			}
		})
	}
}

func TestInspectRejectsChecksumFailureAndDamage(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "bad-checksum.tar.gz")
	manifest := Manifest{FormatVersion: FormatVersion, Application: "subconv-next", AppVersion: "test", CreatedAt: "now", Files: []FileRecord{{Path: "config/subconv-next", Size: 1, SHA256: strings.Repeat("0", 64)}}}
	writeTestArchive(t, archivePath, []testEntry{{name: "config/subconv-next", body: "x"}}, manifest)
	if _, err := Inspect(archivePath); err == nil {
		t.Fatal("Inspect() accepted checksum mismatch")
	}
	mustWrite(t, filepath.Join(root, "damaged.tar.gz"), "not gzip")
	if _, err := Inspect(filepath.Join(root, "damaged.tar.gz")); err == nil {
		t.Fatal("Inspect() accepted damaged archive")
	}
}

func TestRestoreLockAndRollback(t *testing.T) {
	root := t.TempDir()
	original := safeDataDir
	safeDataDir = regexpForTest(root)
	t.Cleanup(func() { safeDataDir = original })
	configPath := filepath.Join(root, "etc", "config", "subconv-next")
	dataDir := filepath.Join(root, "mnt", "data")
	mustWrite(t, configPath, "config subconv-next 'main'\n\toption port '9876'\n\toption data_dir '"+dataDir+"'\n")
	mustWrite(t, filepath.Join(dataDir, "state.json"), "before")
	archivePath := filepath.Join(root, "backup.tar.gz")
	if _, err := Export(ExportOptions{ConfigPath: configPath, DataDir: dataDir, OutputPath: archivePath, AppVersion: "test"}); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dataDir, "state.json"), "current")
	lockPath := filepath.Join(root, "restore.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(RestoreOptions{ArchivePath: archivePath, ConfigPath: configPath, DataDir: dataDir, LockPath: lockPath}); err == nil {
		t.Fatal("Restore() ignored active lock")
	}
	_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	_ = lock.Close()

	initScript := filepath.Join(root, "init")
	statePath := filepath.Join(root, "service-state")
	mustWrite(t, statePath, "running")
	script := "#!/bin/sh\nstate='" + statePath + "'\ncase \"$1\" in running) [ \"$(cat \"$state\")\" = running ];; stop) echo stopped >\"$state\";; start) exit 1;; esac\n"
	mustWriteMode(t, initScript, script, 0755)
	if _, err := Restore(RestoreOptions{ArchivePath: archivePath, ConfigPath: configPath, DataDir: dataDir, InitScript: initScript, LockPath: lockPath}); err == nil {
		t.Fatal("Restore() unexpectedly succeeded when service restart failed")
	}
	state, _ := os.ReadFile(filepath.Join(dataDir, "state.json"))
	if string(state) != "current" {
		t.Fatalf("rollback state = %q, want current", state)
	}
}

func TestPreserveDataDirInsertsIntoMainSection(t *testing.T) {
	input := []byte("config 'subconv-next' 'main'\n\toption port '9876'\n\nconfig other 'tail'\n\toption value '1'\n")
	got, err := preserveDataDir(input, "/mnt/subconv-data")
	if err != nil {
		t.Fatalf("preserveDataDir() error = %v", err)
	}
	want := "config 'subconv-next' 'main'\n\toption data_dir '/mnt/subconv-data'\n\toption port '9876'\n\nconfig other 'tail'\n\toption value '1'\n"
	if string(got) != want {
		t.Fatalf("preserveDataDir() = %q, want %q", got, want)
	}
}

func TestPreserveDataDirReplacesDuplicatesOnlyInMainSection(t *testing.T) {
	input := []byte("config subconv-next main\n\toption data_dir '/mnt/old'\n\toption 'data_dir' '/mnt/duplicate'\nconfig other tail\n\toption data_dir '/keep'\n")
	got, err := preserveDataDir(input, "/overlay/subconv-data")
	if err != nil {
		t.Fatalf("preserveDataDir() error = %v", err)
	}
	text := string(got)
	if strings.Count(text, "option data_dir '/overlay/subconv-data'") != 1 {
		t.Fatalf("main data_dir was not normalized: %q", text)
	}
	if !strings.Contains(text, "option data_dir '/keep'") {
		t.Fatalf("unrelated section was modified: %q", text)
	}
}

type testEntry struct {
	name     string
	body     string
	typeflag byte
	linkname string
}

func writeTestArchive(t *testing.T, archivePath string, entries []testEntry, manifest Manifest) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	if manifest.FormatVersion != 0 {
		if len(manifest.Files) == 0 {
			for _, entry := range entries {
				if entry.typeflag != 0 && entry.typeflag != tar.TypeReg {
					continue
				}
				sum := sha256.Sum256([]byte(entry.body))
				manifest.Files = append(manifest.Files, FileRecord{Path: entry.name, Size: int64(len(entry.body)), SHA256: hex.EncodeToString(sum[:])})
			}
		}
	}
	for _, entry := range entries {
		typeflag := entry.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		header := &tar.Header{Name: entry.name, Mode: 0600, Size: int64(len(entry.body)), Typeflag: typeflag, Linkname: entry.linkname}
		if typeflag != tar.TypeReg {
			header.Size = 0
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Size > 0 {
			_, _ = tw.Write([]byte(entry.body))
		}
	}
	if manifest.FormatVersion != 0 {
		data, _ := json.Marshal(manifest)
		_ = tw.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(data)), Typeflag: tar.TypeReg})
		_, _ = tw.Write(data)
	}
	_ = tw.Close()
	_ = gz.Close()
	_ = file.Close()
}

func mustWrite(t *testing.T, path, value string) { mustWriteMode(t, path, value, 0600) }

func mustWriteMode(t *testing.T, path, value string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), mode); err != nil {
		t.Fatal(err)
	}
}

func regexpForTest(root string) *regexp.Regexp {
	return regexp.MustCompile("^" + regexp.QuoteMeta(root) + `/[A-Za-z0-9._/-]+$`)
}
