package engine

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func HashRepository(name string) RepositoryID {
	sum := sha256.Sum256([]byte(fmt.Sprintf("repository:%d\x00%s", len(name), name)))
	return RepositoryID(hex.EncodeToString(sum[:]))
}

func NewObjectID() (ObjectID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("new object id: %w", err)
	}
	return ObjectID(hex.EncodeToString(b[:])), nil
}
