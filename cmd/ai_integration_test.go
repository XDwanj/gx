package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestAIIntegrationDefineInCommands(t *testing.T) {
	requireOpenAIIntegrationEnv(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	ensureCommandLanguages(t, "go")

	root := commandProject(t, map[string]string{
		"user.go": `package main

type UserService struct{}

func (UserService) login() {
	auditUser()
}

func auditUser() {}
`,
		"employee.go": `package main

type EmployeeService struct{}

func (EmployeeService) login() {
	auditEmployee()
}

func auditEmployee() {}
`,
		"main.go": `package main

func run(user UserService, employee EmployeeService) {
	user.login()
	employee.login()
}
`,
	})

	tests := []struct {
		name       string
		command    string
		assertFunc func(t *testing.T, stdout string)
	}{
		{
			name:    "symbols",
			command: "symbols",
			assertFunc: func(t *testing.T, stdout string) {
				t.Helper()
				var rows []struct {
					File      string `json:"file"`
					Name      string `json:"name"`
					Signature string `json:"signature"`
				}
				decodeAIIntegrationJSON(t, stdout, &rows)
				if len(rows) != 1 {
					t.Fatalf("expected one selected symbol, got %d: %s", len(rows), stdout)
				}
				if rows[0].File != "user.go" || rows[0].Name != "login" {
					t.Fatalf("expected user.go login symbol, got %+v", rows[0])
				}
				if strings.Contains(stdout, "EmployeeService") || strings.Contains(stdout, "employee.go") {
					t.Fatalf("expected employee symbol to be filtered, got %s", stdout)
				}
			},
		},
		{
			name:    "definition",
			command: "definition",
			assertFunc: func(t *testing.T, stdout string) {
				t.Helper()
				var rows []struct {
					File string `json:"file"`
					Body string `json:"body"`
				}
				decodeAIIntegrationJSON(t, stdout, &rows)
				if len(rows) != 1 {
					t.Fatalf("expected one selected definition, got %d: %s", len(rows), stdout)
				}
				if rows[0].File != "user.go" || !strings.Contains(rows[0].Body, "auditUser()") {
					t.Fatalf("expected UserService.login definition, got %+v", rows[0])
				}
				if strings.Contains(stdout, "auditEmployee") || strings.Contains(stdout, "EmployeeService") {
					t.Fatalf("expected employee definition to be filtered, got %s", stdout)
				}
			},
		},
		{
			name:    "callees",
			command: "callees",
			assertFunc: func(t *testing.T, stdout string) {
				t.Helper()
				var rows []struct {
					File   string `json:"file"`
					Caller string `json:"caller"`
					Callee string `json:"callee"`
				}
				decodeAIIntegrationJSON(t, stdout, &rows)
				if len(rows) != 1 {
					t.Fatalf("expected one selected callee row, got %d: %s", len(rows), stdout)
				}
				if rows[0].File != "user.go" || rows[0].Caller != "login" || rows[0].Callee != "auditUser" {
					t.Fatalf("expected UserService.login callee, got %+v", rows[0])
				}
				if strings.Contains(stdout, "auditEmployee") || strings.Contains(stdout, "employee.go") {
					t.Fatalf("expected employee callee to be filtered, got %s", stdout)
				}
			},
		},
		{
			name:    "references",
			command: "references",
			assertFunc: func(t *testing.T, stdout string) {
				t.Helper()
				var rows []struct {
					File    string `json:"file"`
					Context string `json:"context"`
				}
				decodeAIIntegrationJSON(t, stdout, &rows)
				if len(rows) == 0 {
					t.Fatalf("expected selected references, got %s", stdout)
				}
				if !strings.Contains(stdout, "user.login()") {
					t.Fatalf("expected user.login reference, got %s", stdout)
				}
				if strings.Contains(stdout, "employee.login") || strings.Contains(stdout, "EmployeeService") {
					t.Fatalf("expected employee references to be filtered, got %s", stdout)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			args := []string{
				testCase.command,
				"--name", "login",
				"--define-in", "user.go",
				".",
			}
			stdout, stderr, exitCode := executeRootFixtureCommand(t, root, args...)
			if exitCode != 0 {
				t.Fatalf("%s exited %d, stderr=%q stdout=%q", testCase.command, exitCode, stderr, stdout)
			}
			if stderr != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}
			testCase.assertFunc(t, stdout)
		})
	}
}

func requireOpenAIIntegrationEnv(t *testing.T) {
	t.Helper()
	missing := make([]string, 0)
	for _, name := range []string{"GX_OPENAI_API_KEY", "GX_OPENAI_BASE_URL", "GX_OPENAI_MODEL"} {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Skipf("skipping AI integration test; missing %s", strings.Join(missing, ", "))
	}
}

func decodeAIIntegrationJSON(t *testing.T, stdout string, target any) {
	t.Helper()
	if err := json.Unmarshal([]byte(stdout), target); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout=%s", err, stdout)
	}
}

func TestAIIntegrationCacheSurvivesBadAPIKey(t *testing.T) {
	requireOpenAIIntegrationEnv(t)
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)
	ensureCommandLanguages(t, "go")

	root := commandProject(t, map[string]string{
		"user.go": `package main

type UserService struct{}

func (UserService) login() {}
`,
		"employee.go": `package main

type EmployeeService struct{}

func (EmployeeService) login() {}
`,
		"main.go": `package main

func run(user UserService, employee EmployeeService) {
	user.login()
	employee.login()
}
`,
	})

	args := []string{
		"references",
		"--name", "login",
		"--define-in", "user.go",
		".",
	}
	stdout, stderr, exitCode := executeRootFixtureCommand(t, root, args...)
	if exitCode != 0 {
		t.Fatalf("initial references exited %d, stderr=%q stdout=%q", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "user.login()") || strings.Contains(stdout, "employee.login") {
		t.Fatalf("unexpected initial references output: %s", stdout)
	}

	t.Setenv("GX_OPENAI_API_KEY", "bad-key")
	stdout, stderr, exitCode = executeRootFixtureCommand(t, root, args...)
	if exitCode != 0 {
		t.Fatalf("cached references exited %d, stderr=%q stdout=%q", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "user.login()") || strings.Contains(stdout, "employee.login") {
		t.Fatalf("unexpected cached references output: %s", stdout)
	}
}
