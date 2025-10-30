package data

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/0xMishra/makerble/internal/validator"
)

type DoctorModel struct {
	DB *sql.DB
}

type Doctor struct {
	ID             int64              `json:"ID"`
	CreatedAt      time.Time          `json:"created_at"`
	Name           string             `json:"name"`
	Email          string             `json:"email"`
	Password       validator.Password `json:"password"`
	Version        int64              `json:"version"`
	Specialization string             `json:"specialization"`
	Contact        int64              `json:"contact"`
	ShiftStart     time.Time          `json:"shift_start"`
	ShiftEnd       time.Time          `json:"shift_end"`
}

func ValidateDoctor(v *validator.Validator, d *Doctor) {
	v.Check(d.Name != "", "name", "must be provided")
	v.Check(len(d.Name) <= 500, "name", "name must be at most 500 bytes long")
	v.Check(len(d.Specialization) >= 4, "specialization", "doctor must have some specialization")
	v.Check(math.Floor(math.Log10(math.Abs(float64(d.Contact))))+1 == 10, "contact", "doctor's contact number should be at least 10 digits long")

	validator.ValidateEmail(v, d.Email)
	validator.ValidateShift(v, d.ShiftStart, d.ShiftEnd)

	if d.Password.Plaintext != nil {
		validator.ValidatePlaintextPassword(v, *d.Password.Plaintext)
	}

	if d.Password.Hash == nil {
		panic("missing password hash for user")
	}
}

func (m DoctorModel) Insert(d *Doctor) error {
	query := `
		INSERT INTO doctors (name, email, password_hash, specialization, contact, shift_start, shift_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, version

	`
	args := []any{
		d.Name,
		d.Email,
		d.Password.Hash,
		d.Specialization,
		d.Contact,
		d.ShiftStart,
		d.ShiftEnd,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&d.ID, &d.CreatedAt, &d.Version)
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

func (m DoctorModel) GetByEmail(email string) (*Doctor, error) {
	query := `
		SELECT id, created_at, name, email, specialization, contact, shift_start, shift_end
		FROM doctors
		WHERE email = $1
	`

	var d Doctor
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&d.ID,
		&d.CreatedAt,
		&d.Name,
		&d.Email,
		&d.Specialization,
		&d.Contact,
		&d.ShiftStart,
		&d.ShiftEnd,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound

		default:
			return nil, err
		}
	}

	return nil, nil
}

func (m DoctorModel) Update(d *Doctor) error {
	query := `
		UPDATE doctors
		SET name = $1, email = $2, password_hash = $3, specialization = $4, contact = $5, shift_start = $6, shift_end = $7, version = version + 1
		WHERE id = $8
		RETURNING version
	`

	args := []any{
		d.Name,
		d.Email,
		d.Password.Hash,
		d.Specialization,
		d.Contact,
		d.ShiftStart,
		d.ShiftEnd,
		d.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&d.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound

		default:
			return err
		}
	}

	return nil
}
