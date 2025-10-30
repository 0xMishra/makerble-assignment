package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/0xMishra/makerble/internal/data"
	"github.com/0xMishra/makerble/internal/validator"
)

func (app *application) registerReceptionistHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string `json:"name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		ShiftStart string `json:"shift_start"`
		ShiftEnd   string `json:"shift_end"`
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

	rec := &data.Receptionist{
		Name:       input.Name,
		Email:      input.Email,
		ShiftStart: sStart,
		ShiftEnd:   sEnd,
	}

	err = rec.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateReceptionist(v, rec); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Receptionists.Insert(rec)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Tokens.New(3*24*time.Hour, rec.Email, "receptionist", data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"receptionist": rec, "token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
