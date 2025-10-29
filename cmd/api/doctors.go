package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/0xMishra/makerble/internal/data"
	"github.com/0xMishra/makerble/internal/validator"
)

func (app *application) registerDoctorHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		Specialization string `json:"specialization"`
		Contact        int64  `json:"contact"`
		ShiftStart     string `json:"shift_start"`
		ShiftEnd       string `json:"shift_end"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	sStart, err := app.parseShiftTiming(input.ShiftStart)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	sEnd, err := app.parseShiftTiming(input.ShiftEnd)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	d := &data.Doctor{
		Name:           input.Name,
		Email:          input.Email,
		Specialization: input.Specialization,
		ShiftStart:     sStart,
		ShiftEnd:       sEnd,
	}

	err = d.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateDoctor(v, d); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Doctors.Insert(d)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "email already registered")
			app.failedValidationResponse(w, r, v.Errors)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Tokens.New(d.ID, 3*24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"doctor": d, "token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
