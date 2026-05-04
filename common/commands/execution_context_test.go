package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectAgent_EnvVar(t *testing.T) {
	cases := []struct {
		envVar string
		want   string
	}{
		{"CLAUDECODE", "claude"},
		{"CLAUDE_CODE_ENTRYPOINT", "claude"},
		{"GEMINI_CLI", "gemini"},
		{"GOOSE_TERMINAL", "goose"},
		{"CURSOR_AGENT", "cursor"},
		{"CURSOR_TRACEID", "cursor"},
		{"COPILOT_CLI", "copilot"},
		{"KILO_IPC_SOCKET_PATH", "kilocode"},
		{"ROO_CODE_IPC_SOCKET_PATH", "roo_code"},
		{"REPLIT_AGENT", "replit"},
		{"WINDSURF_SESSION_ID", "windsurf"},
		{"AIDER_MODEL", "aider"},
		{"CODEX_HOME", "codex"},
	}
	for _, c := range cases {
		t.Run(c.envVar, func(t *testing.T) {
			clearAgentEnvVars(t)
			t.Setenv(c.envVar, "1")
			assert.Equal(t, c.want, detectAgent())
		})
	}
}

func TestDetectAgent_GenericFallback(t *testing.T) {
	clearAgentEnvVars(t)
	t.Setenv("AGENT", "custom_bot")
	assert.Equal(t, "custom_bot", detectAgent())
}

func TestDetectAgent_None(t *testing.T) {
	clearAgentEnvVars(t)
	assert.Equal(t, "", detectAgent())
}

func TestDetectAgentTraceID(t *testing.T) {
	t.Setenv("CURSOR_TRACEID", "trace-abc")
	assert.Equal(t, "trace-abc", detectAgentTraceID("cursor"))
	// Trace ID gated on agent identity: a leaked CURSOR_TRACEID from an outer
	// shell must not be reused when the real invoker is a different agent.
	assert.Equal(t, "", detectAgentTraceID("claude"))
	assert.Equal(t, "", detectAgentTraceID(""))
}

func TestDetectExecutionContext_AgentAndCI(t *testing.T) {
	clearAgentEnvVars(t)
	clearCIEnvVars(t)
	t.Setenv("CLAUDECODE", "1")
	t.Setenv("GITHUB_ACTIONS", "true")

	inv := DetectExecutionContext()
	assert.True(t, inv.IsAgent)
	assert.True(t, inv.IsCI)
	assert.Equal(t, "claude", inv.Agent)
	assert.Equal(t, "github_actions", inv.CISystem)
}

func TestDetectExecutionContext_CIOnly(t *testing.T) {
	clearAgentEnvVars(t)
	clearCIEnvVars(t)
	t.Setenv("GITHUB_ACTIONS", "true")

	inv := DetectExecutionContext()
	assert.False(t, inv.IsAgent)
	assert.True(t, inv.IsCI)
	assert.Equal(t, "github_actions", inv.CISystem)
}

func TestDetectExecutionContext_NoEnv(t *testing.T) {
	clearAgentEnvVars(t)
	clearCIEnvVars(t)

	inv := DetectExecutionContext()
	assert.False(t, inv.IsAgent)
	assert.False(t, inv.IsCI)
	assert.Equal(t, "", inv.Agent)
	assert.Equal(t, "", inv.CISystem)
}

func clearAgentEnvVars(t *testing.T) {
	t.Helper()
	for _, d := range agentEnvDetectors {
		for _, e := range d.envs {
			t.Setenv(e, "")
		}
	}
	t.Setenv("AGENT", "")
}

func clearCIEnvVars(t *testing.T) {
	t.Helper()
	for _, e := range []string{
		"JENKINS_URL", "TRAVIS", "CIRCLECI", "GITHUB_ACTIONS", "GITLAB_CI",
		"BUILDKITE", "BAMBOO_BUILD_KEY", "TF_BUILD", "TEAMCITY_VERSION",
		"DRONE", "BITBUCKET_BUILD_NUMBER", "CODEBUILD_BUILD_ID",
		"CI", "CONTINUOUS_INTEGRATION", "BUILD_ID", "BUILD_NUMBER",
	} {
		t.Setenv(e, "")
	}
}
