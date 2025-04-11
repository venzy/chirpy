package auth

import "golang.org/x/crypto/bcrypt"

// Note - if you're concerned about how strings and temporary byte slices hang
// around in memory until garbage collected - IMO there seems to be little
// practical benefit to making code more error prone and obscure by trying to
// deal with everything as byte slices that you then zero-out after use, when
// golang and the http library deal with everything as strings anyway...

func HashPassword(password string) (string, error) {
	// Use DefaultCost (passing '0' to get this)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	return string(hash), err
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}