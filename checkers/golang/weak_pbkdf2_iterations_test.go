package golang

import (
	"crypto/rand"
	"crypto/sha256"
	"hash"
)

const fixedSalt = "INSECURE_CONSTANT"

// Mock implementations to test analyzer detection
type pbkdf2Mock struct{}

func (p *pbkdf2Mock) Key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	return nil
}

func testPbkdf2Iterations() {
	var (
		pbkdf2 pbkdf2Mock
	)

	// <expect-error>
	pbkdf2.Key([]byte("password"), []byte(fixedSalt), 100000, 32, sha256.New)

	// <expect-error>
	pbkdf2.Key([]byte("password"), makeRandomSalt(), 250000, 32, sha256.New)

	// Valid examples (no errors)
	// <no-error>
	pbkdf2.Key([]byte("password"), makeRandomSalt(), 310000, 32, sha256.New)
}

func makeRandomSalt() []byte {
	salt := make([]byte, 16)
	rand.Read(salt)
	return salt
}
