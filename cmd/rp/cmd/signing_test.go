package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newArmoredTestKey(t *testing.T) (string, *openpgp.Entity) {
	t.Helper()

	entity, err := openpgp.NewEntity("release bot", "", "release@example.com", &packet.Config{Algorithm: packet.PubKeyAlgoEdDSA})
	require.NoError(t, err)

	var buf bytes.Buffer
	w, err := armor.Encode(&buf, openpgp.PrivateKeyType, nil)
	require.NoError(t, err)
	require.NoError(t, entity.SerializePrivateWithoutSigning(w, nil))
	require.NoError(t, w.Close())

	return buf.String(), entity
}

func Test_signingKeyFromEnv(t *testing.T) {
	key, entity := newArmoredTestKey(t)
	keyFile := filepath.Join(t.TempDir(), "key.asc")
	require.NoError(t, os.WriteFile(keyFile, []byte(key), 0o600))

	tests := []struct {
		name    string
		env     map[string]string
		wantKey bool
		wantErr string
	}{
		{
			name: "nothing set",
			env:  map[string]string{},
		},
		{
			name:    "key in variable",
			env:     map[string]string{EnvSigningKey: key},
			wantKey: true,
		},
		{
			name:    "key in file",
			env:     map[string]string{EnvSigningKeyFile: keyFile},
			wantKey: true,
		},
		{
			name:    "both set",
			env:     map[string]string{EnvSigningKey: key, EnvSigningKeyFile: keyFile},
			wantErr: "set only one of RELEASER_PLEASER_SIGNING_KEY and RELEASER_PLEASER_SIGNING_KEY_FILE",
		},
		{
			name:    "missing file",
			env:     map[string]string{EnvSigningKeyFile: filepath.Join(t.TempDir(), "missing.asc")},
			wantErr: "failed to read RELEASER_PLEASER_SIGNING_KEY_FILE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, name := range []string{EnvSigningKey, EnvSigningKeyFile, EnvSigningKeyPassphrase} {
				t.Setenv(name, tt.env[name])
			}

			got, err := signingKeyFromEnv()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if !tt.wantKey {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, entity.PrimaryKey.KeyId, got.PrimaryKey.KeyId)
		})
	}
}
