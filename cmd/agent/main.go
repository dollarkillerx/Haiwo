package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/haiwo-ci/haiwo/internal/agent"
)

func main() {
	cfg := agent.Config{
		ServerURL:     env("HAIWO_SERVER_URL", "ws://localhost:8080/rpc/agent/ws"),
		Token:         env("HAIWO_AGENT_TOKEN", "dev-agent-token"),
		AgentID:       env("HAIWO_AGENT_ID", "local-agent"),
		Name:          env("HAIWO_AGENT_NAME", "local-agent"),
		Version:       "0.1.0",
		Labels:        splitCSV(env("HAIWO_AGENT_LABELS", "build,deploy,staging")),
		WorkDir:       env("HAIWO_AGENT_WORKDIR", ".haiwo-agent"),
		MaxRunning:    envInt("HAIWO_AGENT_MAX_RUNNING", 1),
		SSHEnabled:    envBool("HAIWO_AGENT_SSH_ENABLED", false),
		ReverseSSHURL: env("HAIWO_AGENT_REVERSE_SSH_URL", ""),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := agent.New(cfg).Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func envInt(key string, fallback int) int {
	v := env(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(env(key, "")))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
