package data

import (
	"database/sql"
	"time"
)

type PatientModel struct {
	DB *sql.DB
}

type Patient struct {
	ID            int64     `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	Name          string    `json:"name"`
	Gender        string    `json:"gender"`
	Age           float64   `json:"age"`
	Contact       int64     `json:"contact"`
	Address       string    `json:"address"`
	MedicalHisory string    `json:"medical_history"`
	InsuranceInfo string    `json:"insurance_info"`
	LastVisit     time.Time `json:"last_visit"`
	Version       int64     `json:"version"`
	DoctorID      int64     `json:"doctor_id"`
}

func (m PatientModel) Insert(p *Patient) error {
	return nil
}

func (m PatientModel) GetByID(id int64) (*Patient, error) {
	return &Patient{}, nil
}

func (m PatientModel) Update(id int64) error {
	return nil
}

func (m PatientModel) Delete(id int64) error {
	return nil
}
