package secrets

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureDashboardCredentialsCreatesBoth(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	creds, generated, err := EnsureDashboardCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if !generated || len(creds.Username) < 12 || len(creds.Password) < 22 {
		t.Fatalf("weak or missing generated credentials: user=%d password=%d", len(creds.Username), len(creds.Password))
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if strings.Count(text, "DASHBOARD_USERNAME=") != 1 || strings.Count(text, "DASHBOARD_PASSWORD=") != 1 {
		t.Fatalf("unexpected env content %q", text)
	}
}

func TestEnsureDashboardCredentialsGeneratesOnlyMissingValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("DASHBOARD_USERNAME=operator\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	creds, generated, err := EnsureDashboardCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if !generated || creds.Username != "operator" || creds.Password == "" {
		t.Fatalf("unexpected credentials %#v", creds)
	}
}

func TestEnsureDashboardCredentialsPreservesCommentsAndAccountSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	original := "# upstream\nuser1=account@example.com\npassword1=secret\nDASHBOARD_USERNAME=admin\nDASHBOARD_PASSWORD=existing-password\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	creds, generated, err := EnsureDashboardCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if generated || creds.Username != "admin" || creds.Password != "existing-password" {
		t.Fatalf("unexpected result %#v generated=%v", creds, generated)
	}
	b, _ := os.ReadFile(path)
	if string(b) != original {
		t.Fatal("existing dotenv file was rewritten")
	}
}

func TestEnsureDashboardCredentialsDoesNotUseProcessEnvironment(t *testing.T) {
	t.Setenv("DASHBOARD_USERNAME", "process-user")
	t.Setenv("DASHBOARD_PASSWORD", "process-password")
	creds, _, err := EnsureDashboardCredentials(filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if creds.Username == "process-user" || creds.Password == "process-password" {
		t.Fatal("process environment unexpectedly overrode .env")
	}
}

func TestEnsureDashboardCredentialsRestrictsExistingFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows ACLs do not expose Unix permission bits")
	}
	path := filepath.Join(t.TempDir(), ".env")
	content := "DASHBOARD_USERNAME=admin\nDASHBOARD_PASSWORD=existing-password\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := EnsureDashboardCredentials(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions = %o, want 600", got)
	}
}
