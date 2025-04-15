package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHash(t *testing.T) {
    password := "Gladys#!456"
    hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("Couldn't hash password '%s': %s\n", password, err)
	}
	t.Logf("Hash of password '%s' is '%s'\n", password, hash)
}

func TestCheckHash(t * testing.T) {
    password := "Gladys#!456"
    hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("Couldn't hash password '%s': %s\n", password, err)
	}

	err = CheckPasswordHash(hash, password)
	if err != nil {
		t.Errorf("Check failed for hash '%s' of password '%s': %s\n", hash, password, err)
	}
}

func TestCheckHashFail(t * testing.T) {
    password := "Gladys#!456"

	err := CheckPasswordHash("not_!@##$@_a_$%R@#$%_real_@#$@#$!_hash", password)
	if err == nil {
		t.Errorf("CheckPasswordHash() should have failed for a bogus hash")
	}
}

func TestValidJWT(t * testing.T) {
	// Create JWT
	userID := uuid.New()
	secret := "eNc0d4_1f3"
	jwt, err := MakeJWT(userID, secret, 5 * time.Second)
	if err != nil {
		t.Errorf("MakeJWT() should have succeeded, err was: %s", err)
	}
	t.Logf("JWT: %s", jwt)

	// Validate JWT
	validatedUserID, err := ValidateJWT(jwt, secret)
	if err != nil {
		t.Errorf("ValidateJWT() should have succeeded, err was: %s", err)
	}
	t.Logf("UserID: %s", validatedUserID)

	if userID != validatedUserID {
		t.Errorf(`validated user ID should match original user ID:
	Original:  %s
	Validated: %s`,
			userID, validatedUserID)
	}
}

func TestExpiredJWT(t * testing.T) {
	// Create JWT
	userID := uuid.New()
	secret := "eNc0d4_1f3"
	jwt, err := MakeJWT(userID, secret, 0 * time.Second)
	if err != nil {
		t.Errorf("MakeJWT() should have succeeded, err was: %s", err)
	}
	t.Logf("JWT: %s", jwt)

	// Validate JWT
	_, err = ValidateJWT(jwt, secret)
	if err == nil {
		t.Errorf("ValidateJWT() should have failed")
	}
	
	if !strings.Contains(err.Error(), "xpire") {
		t.Errorf("ValidateJWT() should have failed with expiry message, but err was %s", err)
	}
}