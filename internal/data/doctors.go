package data

import (
	"database/sql"
	"time"
)

type DoctorModel struct {
	DB *sql.DB
}

type Doctor struct {
	ID             int64     `json:"ID"`
	CreatedAt      time.Time `json:"created_at"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Password       password  `json:"password"`
	Version        int64     `json:"version"`
	Specialization string    `json:"specialization"`
	Contact        int64     `json:"contact"`
	ShiftStart     time.Time `json:"shift_start"`
	ShiftEnd       time.Time `json:"shift_end"`
}

func (m DoctorModel) Insert(p *Doctor) error {
	return nil
}

func (m DoctorModel) GetByEmail(email string) (*Doctor, error) {
	return &Doctor{}, nil
}

func (m DoctorModel) Update(id int64) error {
	return nil
}
