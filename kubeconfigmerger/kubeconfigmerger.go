package kubeconfigmerger

import (
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Config holds options for MergeTokens.
type Config struct {
	// UserNameMappings maps a user name as it appears in updated to the
	// corresponding user name in original, for users whose names differ.
	UserNameMappings map[string]string
}

// MergeTokens returns a deep copy of original where each user's token is
// replaced with the token from the matching user in updated, when one exists.
// Names are matched directly, or via cfg.UserNameMappings when they differ.
func MergeTokens(original, updated *clientcmdapi.Config, cfg Config) *clientcmdapi.Config {
	merged := original.DeepCopy()

	for updatedName, updatedUser := range updated.AuthInfos {
		originalName := updatedName
		if mapped, ok := cfg.UserNameMappings[updatedName]; ok {
			originalName = mapped
		}

		if originalUser, exists := merged.AuthInfos[originalName]; exists {
			originalUser.Token = updatedUser.Token
		}
	}

	return merged
}
