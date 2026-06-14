package kubeconfigserializer

import (
	"encoding/base64"
	"fmt"

	"gopkg.in/yaml.v2"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type kubeconfigYAML struct {
	APIVersion     string             `yaml:"apiVersion"`
	Kind           string             `yaml:"kind"`
	Clusters       []namedClusterYAML `yaml:"clusters"`
	Users          []namedUserYAML    `yaml:"users"`
	Contexts       []namedContextYAML `yaml:"contexts"`
	CurrentContext string             `yaml:"current-context"`
}

type namedClusterYAML struct {
	Name    string      `yaml:"name"`
	Cluster clusterYAML `yaml:"cluster"`
}

type clusterYAML struct {
	Server                   string `yaml:"server,omitempty"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
	InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify,omitempty"`
}

type namedUserYAML struct {
	Name string   `yaml:"name"`
	User userYAML `yaml:"user"`
}

type userYAML struct {
	Token                 string `yaml:"token,omitempty"`
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
}

type namedContextYAML struct {
	Name    string      `yaml:"name"`
	Context contextYAML `yaml:"context"`
}

type contextYAML struct {
	Cluster   string `yaml:"cluster,omitempty"`
	User      string `yaml:"user,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

// Serialize converts a kubeconfig into YAML with a deterministic field order:
// apiVersion → kind → clusters → users → contexts → current-context.
func Serialize(cfg *clientcmdapi.Config) (string, error) {
	out := kubeconfigYAML{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: cfg.CurrentContext,
	}

	for name, cluster := range cfg.Clusters {
		out.Clusters = append(out.Clusters, namedClusterYAML{
			Name: name,
			Cluster: clusterYAML{
				Server:                   cluster.Server,
				CertificateAuthorityData: base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData),
				InsecureSkipTLSVerify:    cluster.InsecureSkipTLSVerify,
			},
		})
	}

	for name, user := range cfg.AuthInfos {
		out.Users = append(out.Users, namedUserYAML{
			Name: name,
			User: userYAML{
				Token:                 user.Token,
				ClientCertificateData: base64.StdEncoding.EncodeToString(user.ClientCertificateData),
				ClientKeyData:         base64.StdEncoding.EncodeToString(user.ClientKeyData),
			},
		})
	}

	for name, ctx := range cfg.Contexts {
		out.Contexts = append(out.Contexts, namedContextYAML{
			Name: name,
			Context: contextYAML{
				Cluster:   ctx.Cluster,
				User:      ctx.AuthInfo,
				Namespace: ctx.Namespace,
			},
		})
	}

	data, err := yaml.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("failed to marshal kubeconfig: %w", err)
	}

	return string(data), nil
}
