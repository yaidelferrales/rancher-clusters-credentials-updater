package updategitlabkubeconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"rancher-clusters-credentials-updater/kubeconfigmerger"
	"rancher-clusters-credentials-updater/kubeconfigserializer"
)

type Config struct {
	GitLabURL        string
	Token            string
	GroupID          string
	VariableKey      string
	UserNameMappings map[string]string
}

type variableResponse struct {
	Value string `json:"value"`
}

// UpdateGitlabKubeconfigVar retrieves a GitLab CI/CD group variable, parses it as a kubeconfig,
// merges its tokens into newKubeconfig using the configured user name mappings, and returns the
// merged kubeconfig serialized as a YAML string.
func UpdateGitlabKubeconfigVar(newKubeconfig *clientcmdapi.Config, cfg Config) error {
	fmt.Printf("Fetching kubeconfig from GitLab variable %q (group %s)...", cfg.VariableKey, cfg.GroupID)
	variableKubeconfig, err := fetchVariable(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Fetched kubeconfig with %d cluster(s).", len(variableKubeconfig.Clusters))

	fmt.Println("Merging tokens...")
	merged := kubeconfigmerger.MergeTokens(variableKubeconfig, newKubeconfig, kubeconfigmerger.Config{
		UserNameMappings: cfg.UserNameMappings,
	})

	data, err := kubeconfigserializer.Serialize(merged)
	if err != nil {
		return err
	}

	fmt.Printf("Writing merged kubeconfig back to GitLab variable %q...", cfg.VariableKey)
	if err := writeVariable(cfg, data); err != nil {
		return fmt.Errorf("failed to update GitLab variable: %w", err)
	}

	fmt.Println("GitLab kubeconfig variable updated successfully.")
	return nil
}

// writeVariable updates the GitLab CI/CD group variable with the given value.
func writeVariable(cfg Config, value string) error {
	apiURL := fmt.Sprintf("%s/api/v4/groups/%s/variables/%s", cfg.GitLabURL, cfg.GroupID, cfg.VariableKey)

	body := url.Values{"value": {value}}
	req, err := http.NewRequest(http.MethodPut, apiURL, strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", cfg.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to GitLab failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab returned status %d when updating variable %q", resp.StatusCode, cfg.VariableKey)
	}

	return nil
}

// fetchVariable retrieves a GitLab CI/CD group variable, parses it as a kubeconfig.
func fetchVariable(cfg Config) (*clientcmdapi.Config, error) {
	url := fmt.Sprintf("%s/api/v4/groups/%s/variables/%s", cfg.GitLabURL, cfg.GroupID, cfg.VariableKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to GitLab failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab returned status %d for variable %q", resp.StatusCode, cfg.VariableKey)
	}

	var variable variableResponse
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("failed to decode GitLab response: %w", err)
	}

	kubeconfig, err := clientcmd.Load([]byte(variable.Value))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig from variable %q: %w", cfg.VariableKey, err)
	}

	return kubeconfig, nil
}
