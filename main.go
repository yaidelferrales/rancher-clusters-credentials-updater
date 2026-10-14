package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"

	"rancher-clusters-credentials-updater/config"
	"rancher-clusters-credentials-updater/credentialsdownloader"
	"rancher-clusters-credentials-updater/updatelocalkubeconfig"
	"rancher-clusters-credentials-updater/updategitlabkubeconfig"
)

func main() {
	envFile := flag.String("env", ".env", "path to the .env file to load")
	flag.Parse()

	// Overload so the selected .env file takes precedence over variables already set in the shell.
	_ = godotenv.Overload(*envFile)
	cfg := config.Load()

	if !cfg.UpdateGitLab && !cfg.UpdateLocal {
		log.Fatal("neither UPDATE_GITLAB nor UPDATE_LOCAL is set to true; nothing to do")
	}

	rancherKubeconfig := credentialsdownloader.Run(cfg.RancherURL, cfg.RancherToken)

	if cfg.UpdateGitLab {
		err := updategitlabkubeconfig.UpdateGitlabKubeconfigVar(rancherKubeconfig, updategitlabkubeconfig.Config{
			GitLabURL:        cfg.GitLabURL,
			Token:            cfg.GitLabToken,
			GroupID:          cfg.GitLabGroupID,
			VariableKey:      cfg.GitLabKubeconfigVar,
			UserNameMappings: cfg.KubeconfigUserMappings,
		})

		if err != nil {
			log.Fatalf("failed to update GitLab kubeconfig variable: %v", err)
		}
	}

	if cfg.UpdateLocal {
		err := updatelocalkubeconfig.UpdateLocalKubeconfig(rancherKubeconfig, updatelocalkubeconfig.Config{
			Path:             cfg.LocalKubeconfigPath,
			UserNameMappings: cfg.KubeconfigUserMappings,
		})

		if err != nil {
			log.Fatalf("failed to update local kubeconfig: %v", err)
		}
	}
}
