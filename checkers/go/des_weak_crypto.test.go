import (
	"crypto/des"
)

func test_des() {
	
    ede2Key := []byte("example key 1234")

	var tripleDESKey []byte
	tripleDESKey = append(tripleDESKey, ede2Key[:16]...)
	tripleDESKey = append(tripleDESKey, ede2Key[:8]...)
	// <expect-error> Weak encryption cipher
	_, err := des.NewTripleDESCipher(tripleDESKey)
	if err != nil {
		panic(err)
	}

}


import (
	"crypto/des"
)

func test_des() {
	
    ede2Key := []byte("example key 1234")
	// <expect-error> Weak encryption cipher
	_, err := des.NewCipher(ede2Key)
	if err != nil {
		panic(err)
	}

}

import (
	"crypto/aes"
)

func test_aes() {
	
	key := []byte("example key 1234")
	// Safe
	_, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

}