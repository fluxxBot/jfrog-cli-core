package commands

import (
	"os"

	"golang.org/x/term"
)

// ExecutionContext describes how a CLI invocation was launched.
// Dimensions are independent: an agent may run inside CI; both will be reported.
type ExecutionContext struct {
	Agent         string // e.g. "claude", "cursor", "gemini", "" if none
	CISystem      string // e.g. "github_actions", "" if not CI
	IsCI          bool
	IsAgent       bool
	IsInteractive bool   // stdin is a TTY
	TraceID       string // propagated trace ID (e.g. CURSOR_TRACEID), empty if none
}

// agent env var detection table. First match wins.
var agentEnvDetectors = []struct {
	name string
	envs []string
}{
	{"claude", []string{"CLAUDECODE", "CLAUDE_CODE_ENTRYPOINT"}},
	{"gemini", []string{"GEMINI_CLI"}},
	{"goose", []string{"GOOSE_TERMINAL"}},
	{"cursor", []string{"CURSOR_AGENT", "CURSOR_TRACEID"}},
	{"copilot", []string{"COPILOT_CLI"}},
	{"kilocode", []string{"KILO_IPC_SOCKET_PATH", "KILO_SERVER_PASSWORD"}},
	{"roo_code", []string{"ROO_CODE_IPC_SOCKET_PATH"}},
	{"replit", []string{"REPLIT_AGENT"}},
	{"windsurf", []string{"WINDSURF_SESSION_ID"}},
	{"aider", []string{"AIDER_MODEL"}},
	{"codex", []string{"CODEX_HOME"}},
}

// DetectExecutionContext captures all signals about who executed the CLI.
func DetectExecutionContext() ExecutionContext {
	ec := ExecutionContext{
		IsInteractive: term.IsTerminal(int(os.Stdin.Fd())),
	}

	ec.Agent = detectAgent()
	ec.IsAgent = ec.Agent != ""
	ec.TraceID = detectAgentTraceID(ec.Agent)

	ec.CISystem = detectCISystem()
	ec.IsCI = ec.CISystem != ""

	return ec
}

func detectAgent() string {
	for _, d := range agentEnvDetectors {
		for _, e := range d.envs {
			if os.Getenv(e) != "" {
				return d.name
			}
		}
	}
	// Fallback: generic AGENT env var (goose convention, codex pending).
	if v := os.Getenv("AGENT"); v != "" {
		return v
	}
	return ""
}

// detectAgentTraceID returns a trace ID propagated by the parent agent, if any.
// Only used when the detected agent itself owns the trace ID env var, so we don't
// reuse a stale ID leaked from an outer shell (e.g. CURSOR_TRACEID present while
// the actual invoker is Claude Code). Empty result means the CLI should generate
// its own trace ID.
func detectAgentTraceID(agent string) string {
	switch agent {
	case "cursor":
		return os.Getenv("CURSOR_TRACEID")
	}
	return ""
}
