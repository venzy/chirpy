package auth

import (
	"testing"
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