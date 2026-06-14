# Agent Guidelines

## Project overview

Go CLI that authenticates with Rancher using a bearer token, fetches kubeconfigs for all clusters through the Rancher v3 API, and updates a GitLab CI/CD group variable and/or a local kubeconfig file with the refreshed tokens.

## Module name

`rancher-clusters-credentials-updater` — use this as the import path prefix for all internal packages.

## Package responsibilities

| Package                  | Responsibility                                                                                                                                                                         |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `config`                 | Load and validate environment variables. GitLab vars are only required when `UPDATE_GITLAB=true`; local path only when `UPDATE_LOCAL=true`.                                            |
| `credentialsdownloader`  | Rancher v3 API client — lists clusters and generates kubeconfigs using a bearer token.                                                                                                 |
| `kubeconfigmerger`       | `MergeTokens` — copies tokens from the new kubeconfig into the original, matching users by name or via `UserNameMappings`. Does not touch clusters, contexts, or any other field.      |
| `kubeconfigserializer`   | `Serialize` — marshals a `*clientcmdapi.Config` to YAML using explicit ordered structs (`apiVersion → kind → clusters → users → contexts → current-context`). Uses `gopkg.in/yaml.v2`. |
| `updategitlabkubeconfig` | Fetches the kubeconfig stored in a GitLab group variable, merges tokens, writes it back via `PUT /api/v4/groups/{id}/variables/{key}`.                                                 |
| `updatelocalkubeconfig`  | Reads a kubeconfig from a local file path, merges tokens, writes the result back with `0600` permissions.                                                                              |

## Key design decisions

- **Bearer token auth**: all Rancher API calls use `Authorization: Bearer <token>`.
- **No file downloads**: kubeconfig data is fetched via `POST /v3/clusters/{id}?action=generateKubeconfig` and kept in memory.
- **Token-only merge**: `kubeconfigmerger` intentionally only updates tokens. Cluster endpoints, certificate data, and contexts come from the stored config, not from Rancher.
- **Ordered YAML serialization**: `clientcmd.Write` is not used because it does not guarantee field order. Custom structs with `yaml` tags are used instead.
- **Optional targets**: both `UPDATE_GITLAB` and `UPDATE_LOCAL` are opt-in via env vars. The app exits early if neither is enabled.
- **Multiple instances**: the `-env` CLI flag selects which `.env` file to load, enabling one binary to serve multiple Rancher/GitLab configurations.
- **TLS**: the HTTP client in `credentialsdownloader` skips TLS verification to support Rancher instances with self-signed certificates.

## Environment variables

| Variable                   | Required         | Default              | Notes                                     |
| -------------------------- | ---------------- | -------------------- | ----------------------------------------- |
| `RANCHER_URL`              | Always           | —                    | Trailing slash is stripped                |
| `RANCHER_TOKEN`            | Always           | —                    | Format: `token-xxxxx:xxxxxxxx...`         |
| `UPDATE_GITLAB`            | No               | `false`              | Must be `true` to enable GitLab update    |
| `GITLAB_URL`               | If UPDATE_GITLAB | `https://gitlab.com` |                                           |
| `GITLAB_TOKEN`             | If UPDATE_GITLAB | —                    | Personal or group access token            |
| `GITLAB_GROUP_ID`          | If UPDATE_GITLAB | —                    | Numeric group ID                          |
| `GITLAB_KUBECONFIG_VAR`    | If UPDATE_GITLAB | —                    | Variable key name                         |
| `UPDATE_LOCAL`             | No               | `false`              | Must be `true` to enable local update     |
| `LOCAL_KUBECONFIG_PATH`    | If UPDATE_LOCAL  | `$HOME/.kube/config` |                                           |
| `KUBECONFIG_USER_MAPPINGS` | No               | —                    | `updated:original` pairs, comma-separated |

## Build

```bash
go build -o bin/rancher-clusters-credentials-updater .
```

## Dependencies to be aware of

- `gopkg.in/yaml.v2` — YAML serialization (do not switch to `sigs.k8s.io/yaml` as it marshals via JSON and loses field order)
- `k8s.io/client-go` — kubeconfig types and `clientcmd.Load`
