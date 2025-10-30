package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"time"

	"github.com/0xMishra/makerble/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(ttl time.Duration, email, role, scope string) (*Token, error) {
	token := &Token{
		Email:  email,
		Role:   role,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	// filling this []byte slice with random entries
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// padding is = at the end of the token that we are avoiding here
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:] // token.Hash is array and that's not acceptable to the pq driver that's why converting to slice

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(ttl time.Duration, email, role, scope string) (*Token, error) {
	token, err := generateToken(ttl, email, role, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, email, role, expiry, scope)
		VALUES ($1, $2, $3, $4, $5)
	`

	args := []any{token.Hash, token.Email, token.Role, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, email string) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND email = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, email)

	return err
}

func (m TokenModel) GetUserForToken(token string) (*Token, error) {
	query := `
		SELECT email, role, expiry FROM tokens
		WHERE token = $1
	`

	var t Token
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, token).Scan(
		&t.Email,
		&t.Role,
		&t.Expiry,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound

		default:
			return nil, err
		}
	}

	return &t, nil
}
