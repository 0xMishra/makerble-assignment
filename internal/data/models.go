// Package data provides data models used by the API
package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Receptionists ReceptionistModel
	Tokens        TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Receptionists: ReceptionistModel{
			DB: db,
		},
		Tokens: TokenModel{
			DB: db,
		},
	}
}
