package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/0xMishra/makerble/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicateEmail = errors.New("duplicate email")

type password struct {
	plaintext *string
	hash      []byte
}

type ReceptionistModel struct {
	DB *sql.DB
}

type Receptionist struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Version    int64     `json:"-"`
	ShiftStart time.Time `json:"shift_start"`
	ShiftEnd   time.Time `json:"shift_end"`
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil

		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePlaintextPassword(v *validator.Validator, plaintext string) {
	v.Check(plaintext != "", "password", "must be provided")
	v.Check(len(plaintext) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(plaintext) <= 72, "password", "must be at most 72 bytes long")
}

func ValidateShift(v *validator.Validator, shiftStart, shiftEnd time.Time) {
	v.Check(shiftEnd.After(shiftStart), "shift", "shift timing should be valid")
}

func ValidateReceptionist(v *validator.Validator, r *Receptionist) {
	v.Check(r.Name != "", "name", "must be provided")
	v.Check(len(r.Name) <= 500, "name", "name must be at most 500 bytes long")

	ValidateEmail(v, r.Email)
	ValidateShift(v, r.ShiftStart, r.ShiftEnd)

	if r.Password.plaintext != nil {
		ValidatePlaintextPassword(v, *r.Password.plaintext)
	}

	if r.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (m ReceptionistModel) Insert(r *Receptionist) error {
	query := `
		INSERT INTO receptionists (name, email, password_hash, shift_start, shift_end)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{r.Name, r.Email, r.Password.hash, r.ShiftStart, r.ShiftEnd}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&r.ID, &r.CreatedAt, &r.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m ReceptionistModel) Update(r *Receptionist) error {
	query := `
		UPDATE receptionists
		SET name = $1, email = $2, password_hash = $3, shift_start = $4, shift_end = $5, version = version + 1
		WHERE id = $6
		RETURNING version
	`

	args := []any{
		&r.Name,
		&r.Email,
		&r.Password.hash,
		&r.ShiftStart,
		&r.ShiftEnd,
		&r.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&r.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m *ReceptionistModel) GetByEmail(email string) (*Receptionist, error) {
	query := `
		SELECT id, created_at, name, email, shift_start, shift_end
		FROM receptionists
		WHERE email = $1
	`

	var r Receptionist
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query).Scan(
		&r.ID,
		&r.Name,
		&r.CreatedAt,
		&r.ShiftStart,
		&r.ShiftEnd,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &r, nil
}
