package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/0xMishra/makerble/internal/data"
	"github.com/0xMishra/makerble/internal/validator"
)

func (app *application) addPatientHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name          string    `json:"name"`
		Gender        string    `json:"gender"`
		Age           float64   `json:"age"`
		Contact       int64     `json:"contact"`
		Address       string    `json:"address"`
		MedicalHisory string    `json:"medical_history"`
		InsuranceInfo string    `json:"insurance_info"`
		LastVisit     time.Time `json:"last_visit"`
		DoctorID      int64     `json:"doctor_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	patient := &data.Patient{
		Name:           input.Name,
		Gender:         input.Gender,
		Age:            input.Age,
		Contact:        input.Contact,
		Address:        input.Address,
		MedicalHistory: input.MedicalHisory,
		InsuranceInfo:  input.InsuranceInfo,
		LastVisit:      input.LastVisit,
		DoctorID:       input.DoctorID,
	}

	v := validator.New()
	if data.ValidatePatient(v, patient); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Patients.Insert(patient)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/patients/%d", patient.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"patient": patient}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getPatientHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	var patient *data.Patient
	patient, err = app.models.Patients.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"patient": patient}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updatePatientHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	var patient *data.Patient

	patient, err = app.models.Patients.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name           *string    `json:"name"`
		Gender         *string    `json:"gender"`
		Age            *float64   `json:"age"`
		Contact        *int64     `json:"contact"`
		Address        *string    `json:"address"`
		MedicalHistory *string    `json:"medical_history"`
		InsuranceInfo  *string    `json:"insurance_info"`
		LastVisit      *time.Time `json:"last_visit"`
		DoctorID       *int64     `json:"doctor_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidatePatient(v, patient); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if input.Name != nil {
		patient.Name = *input.Name
	}
	if input.Gender != nil {
		patient.Gender = *input.Gender
	}
	if input.Age != nil {
		patient.Age = *input.Age
	}
	if input.Contact != nil {
		patient.Contact = *input.Contact
	}
	if input.Address != nil {
		patient.Address = *input.Address
	}
	if input.MedicalHistory != nil {
		patient.MedicalHistory = *input.MedicalHistory
	}
	if input.InsuranceInfo != nil {
		patient.InsuranceInfo = *input.InsuranceInfo
	}
	if input.LastVisit != nil {
		patient.LastVisit = *input.LastVisit
	}
	if input.DoctorID != nil {
		patient.DoctorID = *input.DoctorID
	}

	err = app.models.Patients.Update(patient)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"patient": patient}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deletePatientHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Patients.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "patient info deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
