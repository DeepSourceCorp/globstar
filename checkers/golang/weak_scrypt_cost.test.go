package golang

import (
	"crypto/rand"
)

// Mock implementations to test analyzer detection
type scryptMock struct{}

func (s *scryptMock) Key(password, salt []byte, N, r, p, keyLen int) ([]byte, error) {
	return nil, nil
}

func testScryptCost() {
	var (
		scrypt scryptMock
	)

	// <expect-error>
	scrypt.Key([]byte("password"), makeRandomSalt(), 16384, 8, 1, 32)

	// Valid examples (no errors)
	// <no-error>
	scrypt.Key([]byte("password"), makeRandomSalt(), 32768, 8, 1, 32)
}

func makeRandomSalt() []byte {
	salt := make([]byte, 16)
	rand.Read(salt)
	return salt
}
