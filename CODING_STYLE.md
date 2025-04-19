# Coding Style Guide

This document describes conventions and best practices for the Chirpy project.
It is mostly AI generated.

---

## Imports

- **Standard library imports** should be listed first, in alphabetical order.
- Leave a single blank line.
- **Third-party imports** (those with website-looking names) should follow, also in alphabetical order.

**Example:**
```go
import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/google/uuid"
    "github.com/venzy/chirpy/internal/auth"
)
```

---

## Variable Naming

- When handling data from different sources, use suffixes to clarify origin:
  - Use `Header` or `Req` for variables holding data from HTTP requests or headers (e.g., `refreshTokenHeader`).
  - Use `DB` for variables holding data fetched from the database (e.g., `refreshTokenDB`).
  - Use `New` for newly generated values (e.g., `newRefreshToken`).

**Example:**
```go
refreshTokenHeader, err := auth.GetBearerToken(request.Header)
refreshTokenDB, err := cfg.db.GetRefreshTokenByToken(request.Context(), refreshTokenHeader)
newRefreshToken, err := auth.MakeRefreshToken()
```

---

## Struct Field Alignment

- Align struct fields and tags for readability.

**Example:**
```go
type User struct {
    ID           uuid.UUID `json:"id"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    Email        string    `json:"email"`
    Token        string    `json:"token,omitempty"`
    RefreshToken string    `json:"refresh_token,omitempty"`
}
```

---

## Function Comments

- Use comments to document function purpose and any naming conventions used within.

---

## Prompting Copilot

- When requesting code from Copilot, include naming conventions in your prompt for consistency.

---

## General

- Prefer clarity and explicitness in variable names and code structure.
- Document any additional conventions in this file as they arise.

---