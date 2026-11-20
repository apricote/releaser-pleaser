package git

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testKey struct {
	entity  *openpgp.Entity
	private string
	public  string
}

func newTestKey(t *testing.T, passphrase string) testKey {
	t.Helper()

	config := &packet.Config{Algorithm: packet.PubKeyAlgoEdDSA}
	entity, err := openpgp.NewEntity("release bot", "", "release@example.com", config)
	require.NoError(t, err)

	public := armored(t, openpgp.PublicKeyType, entity.Serialize)

	if passphrase != "" {
		require.NoError(t, entity.EncryptPrivateKeys([]byte(passphrase), config))
	}
	private := armored(t, openpgp.PrivateKeyType, func(w io.Writer) error {
		return entity.SerializePrivateWithoutSigning(w, nil)
	})

	return testKey{entity: entity, private: private, public: public}
}

func armored(t *testing.T, blockType string, serialize func(w io.Writer) error) string {
	t.Helper()

	var buf bytes.Buffer
	w, err := armor.Encode(&buf, blockType, nil)
	require.NoError(t, err)
	require.NoError(t, serialize(w))
	require.NoError(t, w.Close())

	return buf.String()
}

func TestLoadSigningKey(t *testing.T) {
	plain := newTestKey(t, "")
	encrypted := newTestKey(t, "hunter2")
	keyring := armored(t, openpgp.PrivateKeyType, func(w io.Writer) error {
		if err := plain.entity.SerializePrivateWithoutSigning(w, nil); err != nil {
			return err
		}
		return encrypted.entity.SerializePrivateWithoutSigning(w, nil)
	})

	tests := []struct {
		name       string
		armored    string
		passphrase string
		wantErr    string
	}{
		{
			name:    "unencrypted key",
			armored: plain.private,
		},
		{
			name:       "unencrypted key ignores passphrase",
			armored:    plain.private,
			passphrase: "unused",
		},
		{
			name:       "encrypted key with passphrase",
			armored:    encrypted.private,
			passphrase: "hunter2",
		},
		{
			name:    "encrypted key without passphrase",
			armored: encrypted.private,
			wantErr: "signing key is encrypted, but no passphrase was given",
		},
		{
			name:       "encrypted key with wrong passphrase",
			armored:    encrypted.private,
			passphrase: "wrong",
			wantErr:    "failed to decrypt signing key",
		},
		{
			name:    "public key",
			armored: plain.public,
			wantErr: "signing key must be a private key",
		},
		{
			name:    "not a key",
			armored: "hello",
			wantErr: "failed to read signing key",
		},
		{
			name:    "keyring with two keys",
			armored: keyring,
			wantErr: "signing key must contain exactly one key, found 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := LoadSigningKey(tt.armored, tt.passphrase)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.False(t, entity.PrivateKey.Encrypted)
		})
	}
}

func TestRepository_CommitSignsWithKey(t *testing.T) {
	key := newTestKey(t, "hunter2")
	entity, err := LoadSigningKey(key.private, "hunter2")
	require.NoError(t, err)

	repo := WithTestRepo()(t)
	err = repo.UpdateFile(context.Background(), "README.md", false, func(content string) (string, error) {
		return content + "\nrelease", nil
	})
	require.NoError(t, err)

	commit, err := repo.Commit(context.Background(), "chore: release v1.2.3", Author{Name: "release bot", Email: "release@example.com"}, entity)
	require.NoError(t, err)

	obj, err := repo.r.CommitObject(plumbing.NewHash(commit.Hash))
	require.NoError(t, err)

	require.NotEmpty(t, obj.PGPSignature)
	signer, err := obj.Verify(key.public)
	require.NoError(t, err)
	assert.Equal(t, entity.PrimaryKey.KeyId, signer.PrimaryKey.KeyId)
}
