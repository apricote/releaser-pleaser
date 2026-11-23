package git

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// LoadSigningKey parses an ASCII-armored OpenPGP private key, decrypting it with passphrase when it is encrypted.
func LoadSigningKey(armored, passphrase string) (*openpgp.Entity, error) {
	entities, err := openpgp.ReadArmoredKeyRing(strings.NewReader(armored))
	if err != nil {
		return nil, fmt.Errorf("failed to read signing key: %w", err)
	}
	if len(entities) != 1 {
		return nil, fmt.Errorf("signing key must contain exactly one key, found %d", len(entities))
	}

	entity := entities[0]
	if entity.PrivateKey == nil {
		return nil, errors.New("signing key must be a private key")
	}

	if isEncrypted(entity) {
		if passphrase == "" {
			return nil, errors.New("signing key is encrypted, but no passphrase was given")
		}
		if err = entity.DecryptPrivateKeys([]byte(passphrase)); err != nil {
			return nil, fmt.Errorf("failed to decrypt signing key: %w", err)
		}
	}

	if _, ok := entity.SigningKey(time.Now()); !ok {
		return nil, errors.New("signing key has no valid key that can sign")
	}

	return entity, nil
}

func isEncrypted(entity *openpgp.Entity) bool {
	if entity.PrivateKey.Encrypted {
		return true
	}
	for _, sub := range entity.Subkeys {
		if sub.PrivateKey != nil && sub.PrivateKey.Encrypted {
			return true
		}
	}
	return false
}
