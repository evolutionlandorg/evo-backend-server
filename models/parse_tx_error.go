package models

import "github.com/jinzhu/gorm"

type ParseTxError struct {
	gorm.Model
	Tx        string `json:"tx"`
	Chain     string `json:"chain"`
	Error     string `json:"error"`
	ParseFunc string `json:"parse_func"`
	Receipts  string `json:"receipts" sql:"type:text;"`
}
