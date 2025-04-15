package auth

import (
	"net/http"
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
		return
	}
	t.Logf("JWT: %s", jwt)

	// Validate JWT
	validatedUserID, err := ValidateJWT(jwt, secret)
	if err != nil {
		t.Errorf("ValidateJWT() should have succeeded, err was: %s", err)
		return
	}
	t.Logf("UserID: %s", validatedUserID)

	if userID != validatedUserID {
		t.Errorf(`validated user ID should match original user ID:
	Original:  %s
	Validated: %s`,
			userID, validatedUserID)
		return
	}
}

func TestExpiredJWT(t * testing.T) {
	// Create JWT
	userID := uuid.New()
	secret := "eNc0d4_1f3"
	jwt, err := MakeJWT(userID, secret, 0 * time.Second)
	if err != nil {
		t.Errorf("MakeJWT() should have succeeded, err was: %s", err)
		return
	}
	t.Logf("JWT: %s", jwt)

	// Validate JWT
	_, err = ValidateJWT(jwt, secret)
	if err == nil {
		t.Errorf("ValidateJWT() should have failed")
		return
	}
	
	if !strings.Contains(err.Error(), "xpire") {
		t.Errorf("ValidateJWT() should have failed with expiry message, but err was %s", err)
		return
	}
}


func TestGetBearerToken(t * testing.T) {
	// This way of doing data-driven sub-tests was cribbed from Ch6 L6 solution files. Trying it out.

	tests := []struct {
		name string
		header string
		wantToken string
		wantErr bool
	}{
		{
			name: "Valid header",
			header: "Bearer some.kind.of.token",
			wantToken: "some.kind.of.token",
			wantErr: false,
		},
		{
			name: "Valid header more whitespace",
			header: "Bearer 	some.kind.of.token",
			wantToken: "some.kind.of.token",
			wantErr: false,
		},
		{
			name: "Valid header actual example",
			header: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHkiLCJzdWIiOiJhZGE0Y2Q2Ni1iNTkyLTRhNGItYjU2NC00ZTE1Zjk1ZjRkMmMiLCJleHAiOjE3NDQ3MjM1NDIsImlhdCI6MTc0NDcyMzUzN30.68muRMBn_Wr4PpirNJk60ZBAgJDhPZz5l3SgXfpXDFg",
			wantToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHkiLCJzdWIiOiJhZGE0Y2Q2Ni1iNTkyLTRhNGItYjU2NC00ZTE1Zjk1ZjRkMmMiLCJleHAiOjE3NDQ3MjM1NDIsImlhdCI6MTc0NDcyMzUzN30.68muRMBn_Wr4PpirNJk60ZBAgJDhPZz5l3SgXfpXDFg",
			wantErr: false,
		},
		{
			name: "Empty header",
			header: "",
			wantToken: "",
			wantErr: true,
		},
		{
			name: "Header missing token",
			header: "Bearer",
			wantToken: "",
			wantErr: true,
		},
		{
			name: "Header missing token extra whitespace",
			header: "Bearer     ",
			wantToken: "",
			wantErr: true,
		},
		{
			name: "Header malformed",
			header: "Bearer some.token with.whitespace",
			wantToken: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			headers.Add("Authorization", tt.header)
			token, err := GetBearerToken(headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if token != tt.wantToken {
				t.Errorf("GetBearerToken() token = %v, wantToken = %v", token, tt.wantToken)
				return
			}
		})
	}
}