package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/0xMishra/makerble/internal/validator"
)

type PatientModel struct {
	DB *sql.DB
}

type Patient struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	Name           string    `json:"name"`
	Gender         string    `json:"gender"`
	Age            float64   `json:"age"`
	Contact        int64     `json:"contact"`
	Address        string    `json:"address"`
	MedicalHistory string    `json:"medical_history"`
	InsuranceInfo  string    `json:"insurance_info"`
	LastVisit      time.Time `json:"last_visit"`
	Version        int64     `json:"version"`
	DoctorID       int64     `json:"doctor_id"`
}

func ValidatePatient(v *validator.Validator, p *Patient) {
	v.Check(len(p.Name) != 0, "name", "name must be provided")
	v.Check(p.Gender == "male" || p.Gender == "female" || p.Gender == "others", "gender", "gender can only be male, female or others")
	v.Check(p.Age >= 0, "age", "age cannot be negative or zero")
	v.Check(p.Contact >= 10, "contact", "contact number should be of at least 10 digits")

	v.Check(len(p.MedicalHistory) != 0, "medical history", "medical history must be provided")
	v.Check(len(p.InsuranceInfo) != 0, "insurance info", "insurance info must be provided")

	v.Check(p.LastVisit != time.Time{}, "last visit", "last visit must be provided")
	v.Check(p.LastVisit.Before(time.Now()), "last visit", "last visit should be before now")

	v.Check(p.DoctorID >= 0, "doctor id", "doctor's id must be provided")
}

func (m PatientModel) Insert(p *Patient) error {
	query := `
		INSERT INTO patients (name, gender, age, contact, address, medical_history, insurance_info, last_visit, doctor_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_id, version
	`

	args := []any{
		p.Name,
		p.Gender,
		p.Age,
		p.Contact,
		p.Address,
		p.MedicalHistory,
		p.InsuranceInfo,
		p.LastVisit,
		p.DoctorID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&p.ID,
		&p.CreatedAt,
		&p.Version,
	)
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

func (m PatientModel) GetByID(id int64) (*Patient, error) {
	query := `
		SELECT id, created_at, name, gender, age, contact, address, medical_history, insurance_info, last_visit, doctor_id
		FROM patients
		WHERE id = $1
	`
	var p Patient
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&p.Name,
		&p.Gender,
		&p.Age,
		&p.Contact,
		&p.Address,
		&p.MedicalHistory,
		&p.InsuranceInfo,
		&p.LastVisit,
		&p.DoctorID,
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

func (m PatientModel) Update(p *Patient) error {
	query := `
		UPDATE patients
		SET name = $1, gender = $2, age = $3, contact = $4, address = $5, medical_history = $6, insurance_info = $7, last_visit = $8, doctor_id = $9, version = version + 1
		WHERE id = $10  AND version = $11
		RETURNING version
	`

	args := []any{
		p.Name,
		p.Gender,
		p.Age,
		p.Contact,
		p.Address,
		p.MedicalHistory,
		p.InsuranceInfo,
		p.LastVisit,
		p.DoctorID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&p.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict

		default:
			return err
		}
	}

	return nil
}

func (m PatientModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM patients
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
