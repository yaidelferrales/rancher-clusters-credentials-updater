package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	RancherURL   string
	RancherToken string

	UpdateGitLab        bool
	GitLabURL           string
	GitLabToken         string
	GitLabGroupID       string
	GitLabKubeconfigVar string

	UpdateLocal         bool
	LocalKubeconfigPath string

	// KubeconfigUserMappings maps user names in the updated kubeconfig to the
	// corresponding user names in the original. Set via KUBECONFIG_USER_MAPPINGS
	// as a comma-separated list of updated:original pairs, e.g. "alice:alice-admin,bob:bob-ops".
	KubeconfigUserMappings map[string]string
}

func Load() Config {
	cfg := Config{
		RancherURL:   strings.TrimRight(mustGetenv("RANCHER_URL"), "/"),
		RancherToken: mustGetenv("RANCHER_TOKEN"),

		UpdateGitLab: os.Getenv("UPDATE_GITLAB") == "true",
		UpdateLocal:  os.Getenv("UPDATE_LOCAL") == "true",

		KubeconfigUserMappings: parseUserMappings(os.Getenv("KUBECONFIG_USER_MAPPINGS")),
	}

	if cfg.UpdateGitLab {
		cfg.GitLabURL = mustGetenv("GITLAB_URL")
		cfg.GitLabToken = mustGetenv("GITLAB_TOKEN")
		cfg.GitLabGroupID = mustGetenv("GITLAB_GROUP_ID")
		cfg.GitLabKubeconfigVar = mustGetenv("GITLAB_KUBECONFIG_VAR")
	}

	if cfg.UpdateLocal {
		cfg.LocalKubeconfigPath = getenvOrDefault("LOCAL_KUBECONFIG_PATH", os.Getenv("HOME")+"/.kube/config")
	}

	return cfg
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func getenvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// parseUserMappings parses a comma-separated list of "updated:original" pairs.
func parseUserMappings(raw string) map[string]string {
	mappings := map[string]string{}
	if raw == "" {
		return mappings
	}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			mappings[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return mappings
}
