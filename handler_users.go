package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/venzy/chirpy/internal/auth"
	"github.com/venzy/chirpy/internal/database"
)

const accessTokenExpiry = time.Hour
const refreshTokenExpiry = 60 * 24 * time.Hour // 60 days

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

func (cfg *apiConfig) handleCreateUser(response http.ResponseWriter, request *http.Request) {
	type requestParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("users: Error decoding createUser params: %s\n", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Basic validation
	if _, err := mail.ParseAddress(params.Email); err != nil {
		msg := fmt.Sprintf("users: Bad email address: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	if len(params.Password) < 1 {
		msg := fmt.Sprintf("users: Password len > 1 required")
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Create in DB, then return representation of new row as response
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		msg := fmt.Sprintf("users: Problem hashing supplied password")
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	newRow, err := cfg.db.CreateUser(request.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		msg := fmt.Sprintf("users: Problem creating user: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	newUser := User{
		ID: newRow.ID,
		CreatedAt: newRow.CreatedAt,
		UpdatedAt: newRow.UpdatedAt,
		Email: newRow.Email,
	}
	respondWithJSON(response, http.StatusCreated, newUser)
}

func (cfg *apiConfig) handleLogin(response http.ResponseWriter, request *http.Request) {
	type requestParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("users: Error decoding login params: %s\n", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Validate email is well-formed
	if _, err := mail.ParseAddress(params.Email); err != nil {
		msg := fmt.Sprintf("users: Bad email address: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Get user
	user, err := cfg.db.GetUserByEmail(request.Context(), params.Email)
	if err != nil {
		msg := fmt.Sprintf("users: No such user '%s' or error: %s", params.Email, err)
		log.Println(msg)
		// Don't leak which was wrong (email or password)
		respondWithError(response, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// Check password
	if err = auth.CheckPasswordHash(user.HashedPassword, params.Password); err != nil {
		msg := fmt.Sprintf("users: Password hash mismatch or error: %s", err)
		log.Println(msg)
		// Don't leak which was wrong (email or password)
		respondWithError(response, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// Create access token
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, accessTokenExpiry)
	if err != nil {
		msg := fmt.Sprintf("users: login couldn't create JWT: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	// Create refresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		msg := fmt.Sprintf("users: login couldn't create refresh token: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	// Store refresh token in DB
	_, err = cfg.db.CreateRefreshToken(request.Context(), database.CreateRefreshTokenParams{
		UserID: user.ID,
		Token: refreshToken,
		ExpiresAt: time.Now().Add(refreshTokenExpiry),
	})
	if err != nil {
		msg := fmt.Sprintf("users: login couldn't store refresh token: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	loggedInUser := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
		RefreshToken: refreshToken,
	}
	respondWithJSON(response, http.StatusOK, loggedInUser)
}

// DV: Here is a future-looking commentary, based on a long discussion with the
// boot.dev AI, based on me identifying a cleanup issue, and deciding not to
// address it for now:
//
// This function generates and issues a new refresh token for the user. 
// While cleanup of old tokens will not be performed currently, 
// we recognize the importance of preventing refresh token bloat over time.
//
// Potential future improvements:
// 1. Retain only the last 2-3 refresh tokens per session/user to allow
//    recovery from communication failures while minimizing risks of misuse.
// 2. Cleanup stale or expired tokens during future refresh requests, 
//    ensuring efficiency without leaving behind unnecessary data.
// 3. Consider session-specific identifiers (e.g., UUIDs) to manage tokens 
//    across different sessions or browser modes on the same device.
// 4. Device fingerprints or session annotations could enhance management 
//    and allow multi-session support (e.g., regular vs incognito mode).
// 5. Use database indices (e.g., on `expires_at`) to ensure efficient 
//    queries and pruning during cleanup.
//
// Current tokens remain valid until explicitly revoked or naturally expired.
func (cfg *apiConfig) handleRefresh(response http.ResponseWriter, request *http.Request) {
    // Get refresh token from Authorization header
    refreshTokenHeader, err := auth.GetBearerToken(request.Header)
    if err != nil {
        respondWithError(response, http.StatusUnauthorized, "Invalid refresh token")
        return
    }

    // Look up the refresh token in the database, validating the user ID at the same time
    refreshTokenDB, err := cfg.db.GetUserIDWithRefreshToken(request.Context(), refreshTokenHeader)
    if err != nil {
        respondWithError(response, http.StatusUnauthorized, "Invalid refresh token")
        return
    }

    // Check if token is expired or revoked
    if refreshTokenDB.ExpiresAt.Before(time.Now()) || refreshTokenDB.RevokedAt.Valid {
        respondWithError(response, http.StatusUnauthorized, "Invalid refresh token")
        return
    }

    // Generate a new access token
    newAccessJWT, err := auth.MakeJWT(refreshTokenDB.UserID, cfg.jwtSecret, accessTokenExpiry)
    if err != nil {
        respondWithError(response, http.StatusInternalServerError, "Could not generate new access token")
        return
    }

    // Respond with the new access token
    respondWithJSON(response, http.StatusOK, map[string]string{
        "token": newAccessJWT,
    })
}

func (cfg *apiConfig) handleRevoke(response http.ResponseWriter, request *http.Request) {
	// Get refresh token from Authorization header
	refreshTokenHeader, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(response, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Revoke the refresh token in the database
	err = cfg.db.RevokeRefreshToken(request.Context(), refreshTokenHeader)
	if err != nil {
		respondWithError(response, http.StatusInternalServerError, "Could not revoke refresh token")
		return
	}
	// Respond with No Content
	response.WriteHeader(http.StatusNoContent)
}