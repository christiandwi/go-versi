package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func HashRepository(name string) RepositoryID {
	sum := sha256.Sum256([]byte(fmt.Sprintf("repository:%d\x00%s", len(name), name)))
	return RepositoryID(hex.EncodeToString(sum[:]))
}
