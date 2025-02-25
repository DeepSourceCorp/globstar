package golang

// Mock implementations to test analyzer detection
type bcryptMock struct{}

func (b *bcryptMock) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	return nil, nil
}

func testBcryptCost() {
	var (
		bcrypt bcryptMock
	)

	// <expect-error>
	bcrypt.GenerateFromPassword([]byte("password"), 8)

	// <no-error>
	bcrypt.GenerateFromPassword([]byte("password"), 12)
}
