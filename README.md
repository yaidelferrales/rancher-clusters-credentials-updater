# Rancher Clusters Credentials Manager

A CLI tool that fetches kubeconfig credentials for all clusters from a Rancher instance via its API, and distributes the updated tokens to one or both of:

- A **GitLab CI/CD group variable** (so pipelines always have fresh credentials)
- The **local kubeconfig file** (`~/.kube/config` by default)

## How it works

1. Uses a Rancher API bearer token to list all clusters via the Rancher v3 API.
2. Generates a kubeconfig per cluster (`POST /v3/clusters/{id}?action=generateKubeconfig`) and merges them in memory.
3. Merges the new tokens into the existing kubeconfig (GitLab variable or local file), preserving the rest of the config unchanged.
4. Writes the result back as a well-ordered YAML file (`apiVersion → kind → clusters → users → contexts → current-context`).

Multiple Rancher/GitLab instances are supported by using a separate `.env` file per configuration.

## Installation

```bash
go build -o bin/rancher-clusters-credentials-updater .
```

## Usage

```bash
# Use the default .env file
./bin/rancher-clusters-credentials-updater

# Use a specific .env file
./bin/rancher-clusters-credentials-updater -env staging.env
```

## Configuration

Copy `.env.example` to `.env` (or any name you like) and fill in the values:

```env
# Rancher — required
RANCHER_URL=https://rancher.example.com
RANCHER_TOKEN=token-xxxxx:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# GitLab — required only when UPDATE_GITLAB=true
UPDATE_GITLAB=true
GITLAB_URL=https://gitlab.com
GITLAB_TOKEN=glpat-xxxxxxxxxxxx
GITLAB_GROUP_ID=123
GITLAB_KUBECONFIG_VAR=KUBECONFIG

# Local kubeconfig — required only when UPDATE_LOCAL=true
UPDATE_LOCAL=true
LOCAL_KUBECONFIG_PATH=~/.kube/config

# Optional: map user names that differ between the Rancher kubeconfig and the
# stored one. Comma-separated updated:original pairs.
KUBECONFIG_USER_MAPPINGS=alice:alice-admin,bob:bob-ops
```

At least one of `UPDATE_GITLAB` or `UPDATE_LOCAL` must be set to `true`.

### Obtaining a Rancher API token

In the Rancher UI go to **User Settings → API & Keys → Add Key**. Copy the generated token (format: `token-xxxxx:xxxxxxxx...`) — it is only shown once.

## Project structure

```
.
├── main.go                         # Entry point, flag parsing, orchestration
├── config/                         # Environment variable loading
├── credentialsdownloader/          # Rancher API client — lists clusters, fetches kubeconfigs
├── kubeconfigmerger/               # Merges tokens from new config into existing config
├── kubeconfigserializer/           # Serializes kubeconfig to YAML with deterministic field order
├── updategitlabkubeconfig/         # Fetches GitLab variable, merges, writes back
└── updatelocalkubeconfig/          # Reads local file, merges, writes back
```

## Scheduling

To run automatically, add a crontab entry and redirect output to a log file:

```cron
0 6 * * * /path/to/bin/rancher-clusters-credentials-updater -env /path/to/prod.env >> /var/log/credentials-manager.log 2>&1
```
