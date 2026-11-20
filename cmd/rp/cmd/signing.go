package cmd

import (
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"

	"github.com/apricote/releaser-pleaser/internal/git"
)

const (
	EnvSigningKey           = "RELEASER_PLEASER_SIGNING_KEY"
	EnvSigningKeyFile       = "RELEASER_PLEASER_SIGNING_KEY_FILE"
	EnvSigningKeyPassphrase = "RELEASER_PLEASER_SIGNING_KEY_PASSPHRASE"
)

// signingKeyFromEnv loads the release commit signing key, returning nil when none is configured.
func signingKeyFromEnv() (*openpgp.Entity, error) {
	armored := os.Getenv(EnvSigningKey)

	if path := os.Getenv(EnvSigningKeyFile); path != "" {
		if armored != "" {
			return nil, fmt.Errorf("set only one of %s and %s", EnvSigningKey, EnvSigningKeyFile)
		}

		content, err := os.ReadFile(path) // nolint:gosec // The user chose this path to point at their key.
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", EnvSigningKeyFile, err)
		}
		armored = string(content)
	}

	if armored == "" {
		return nil, nil
	}

	return git.LoadSigningKey(armored, os.Getenv(EnvSigningKeyPassphrase))
}
