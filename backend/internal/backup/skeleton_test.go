package backup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupScriptsPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	required := []string{
		"scripts/backup-postgres.sh",
		"scripts/restore-postgres.sh",
		"scripts/install-backup-cron.sh",
		"scripts/Backup.ps1",
	}
	for _, rel := range required {
		path := filepath.Join(root, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
		body := string(data)
		if len(strings.TrimSpace(body)) == 0 {
			t.Fatalf("%s is empty", rel)
		}
		switch {
		case strings.HasSuffix(rel, ".sh"):
			if !strings.HasPrefix(body, "#!/") {
				t.Fatalf("%s missing shebang", rel)
			}
			if !strings.Contains(body, "pg_dump") &&
				!strings.Contains(body, "backup-postgres.sh") &&
				!strings.Contains(body, "psql") {
				t.Fatalf("%s should reference pg_dump, psql, or backup script", rel)
			}
		case strings.HasSuffix(rel, ".ps1"):
			if !strings.Contains(body, "pg_dump") {
				t.Fatalf("%s missing pg_dump", rel)
			}
		}
	}
}

func TestBackupScriptRetentionAndGzip(t *testing.T) {
	path := filepath.Join("..", "..", "..", "scripts", "backup-postgres.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, needle := range []string{
		"RETENTION_DAYS",
		"gzip",
		"streampass_latest.sql.gz",
		"pg_dump",
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("backup-postgres.sh missing %q", needle)
		}
	}
}

func TestRestoreRequiresConfirm(t *testing.T) {
	path := filepath.Join("..", "..", "..", "scripts", "restore-postgres.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "CONFIRM=yes") {
		t.Fatal("restore must require CONFIRM=yes")
	}
}
