package root

import (
	"testing"

	"github.com/spf13/viper"
)

func TestEnvironmentOnlyConfiguration(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("JIRA_SERVER", "https://example.atlassian.net")
	t.Setenv("JIRA_LOGIN", "user@example.com")
	t.Setenv("JIRA_API_TOKEN", "token")
	t.Setenv("JIRA_PROJECT_KEY", "TEST")
	t.Setenv("JIRA_PROJECT_TYPE", "software")

	cmd := NewCmdRoot()
	cmd.SetArgs([]string{"me"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute with environment-only configuration: %v", err)
	}

	if got := viper.GetString("server"); got != "https://example.atlassian.net" {
		t.Errorf("server = %q, want %q", got, "https://example.atlassian.net")
	}
	if got := viper.GetString("login"); got != "user@example.com" {
		t.Errorf("login = %q, want %q", got, "user@example.com")
	}
	if got := viper.GetString("project.key"); got != "TEST" {
		t.Errorf("project.key = %q, want %q", got, "TEST")
	}
	if got := viper.GetString("project.type"); got != "software" {
		t.Errorf("project.type = %q, want %q", got, "software")
	}
}
