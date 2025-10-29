package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/0xMishra/makerble/internal/validator"
)

var ErrDuplicateEmail = errors.New("duplicate email")

type ReceptionistModel struct {
	DB *sql.DB
}

type Receptionist struct {
	ID         int64              `json:"id"`
	CreatedAt  time.Time          `json:"created_at"`
	Name       string             `json:"name"`
	Email      string             `json:"email"`
	Password   validator.Password `json:"-"`
	Version    int64              `json:"-"`
	ShiftStart time.Time          `json:"shift_start"`
	ShiftEnd   time.Time          `json:"shift_end"`
}

func ValidateReceptionist(v *validator.Validator, r *Receptionist) {
	v.Check(r.Name != "", "name", "must be provided")
	v.Check(len(r.Name) <= 500, "name", "name must be at most 500 bytes long")

	validator.ValidateEmail(v, r.Email)
	validator.ValidateShift(v, r.ShiftStart, r.ShiftEnd)

	if r.Password.Plaintext != nil {
		validator.ValidatePlaintextPassword(v, *r.Password.Plaintext)
	}

	if r.Password.Hash == nil {
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

	args := []any{r.Name, r.Email, r.Password.Hash, r.ShiftStart, r.ShiftEnd}

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
		r.Name,
		r.Email,
		r.Password.Hash,
		r.ShiftStart,
		r.ShiftEnd,
		r.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&r.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
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

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
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
