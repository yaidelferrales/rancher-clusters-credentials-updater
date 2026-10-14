package credentialsdownloader

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func Run(rancherURL, rancherToken string) *clientcmdapi.Config {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fmt.Println("Listing clusters...")
	clusterIDs, err := listClusters(client, rancherURL, rancherToken)
	if err != nil {
		log.Fatalf("failed to list clusters: %v", err)
	}
	fmt.Printf("Found %d cluster(s).\n", len(clusterIDs))

	merged := clientcmdapi.NewConfig()
	for _, id := range clusterIDs {
		fmt.Printf("Fetching kubeconfig for cluster %s...\n", id)
		cfg, err := generateKubeconfig(client, rancherURL, rancherToken, id)
		if err != nil {
			log.Fatalf("failed to generate kubeconfig for cluster %s: %v", id, err)
		}
		mergeInto(merged, cfg)
	}

	return merged
}

func setAuth(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

type clusterList struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func listClusters(client *http.Client, rancherURL, token string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, rancherURL+"/v3/clusters", nil)
	if err != nil {
		return nil, err
	}
	setAuth(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("listing clusters returned status %d: %s", resp.StatusCode, body)
	}

	var list clusterList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(list.Data))
	for _, c := range list.Data {
		ids = append(ids, c.ID)
	}
	return ids, nil
}

type kubeconfigResponse struct {
	Config string `json:"config"`
}

func generateKubeconfig(client *http.Client, rancherURL, token, clusterID string) (*clientcmdapi.Config, error) {
	url := fmt.Sprintf("%s/v3/clusters/%s?action=generateKubeconfig", rancherURL, clusterID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	setAuth(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generateKubeconfig returned status %d: %s", resp.StatusCode, body)
	}

	var result kubeconfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return clientcmd.Load([]byte(result.Config))
}

func mergeInto(dst, src *clientcmdapi.Config) {
	for name, cluster := range src.Clusters {
		dst.Clusters[name] = cluster
	}
	for name, user := range src.AuthInfos {
		dst.AuthInfos[name] = user
	}
	for name, ctx := range src.Contexts {
		dst.Contexts[name] = ctx
	}
	if dst.CurrentContext == "" {
		dst.CurrentContext = src.CurrentContext
	}
}
