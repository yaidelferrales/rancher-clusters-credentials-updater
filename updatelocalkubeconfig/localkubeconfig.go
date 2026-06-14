package updatelocalkubeconfig

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"rancher-clusters-credentials-updater/kubeconfigmerger"
	"rancher-clusters-credentials-updater/kubeconfigserializer"
)

type Config struct {
	Path             string
	UserNameMappings map[string]string
}

// UpdateLocalKubeconfig reads the kubeconfig at cfg.Path, merges tokens from
// newKubeconfig into it, and writes the result back to the same file.
func UpdateLocalKubeconfig(newKubeconfig *clientcmdapi.Config, cfg Config) error {
	fmt.Printf("Reading local kubeconfig from %q...", cfg.Path)
	localKubeconfig, err := readKubeconfig(cfg.Path)
	if err != nil {
		return err
	}
	fmt.Printf("Read kubeconfig with %d cluster(s).", len(localKubeconfig.Clusters))

	fmt.Println("Merging tokens...")
	merged := kubeconfigmerger.MergeTokens(localKubeconfig, newKubeconfig, kubeconfigmerger.Config{
		UserNameMappings: cfg.UserNameMappings,
	})

	data, err := kubeconfigserializer.Serialize(merged)
	if err != nil {
		return err
	}

	fmt.Printf("Writing merged kubeconfig to %q...", cfg.Path)
	if err := writeKubeconfig(cfg.Path, data); err != nil {
		return err
	}

	fmt.Println("Local kubeconfig updated successfully.")
	return nil
}

func readKubeconfig(path string) (*clientcmdapi.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig at %q: %w", path, err)
	}

	kubeconfig, err := clientcmd.Load(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig at %q: %w", path, err)
	}

	return kubeconfig, nil
}

func writeKubeconfig(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig to %q: %w", path, err)
	}
	return nil
}
