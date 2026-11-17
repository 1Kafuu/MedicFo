package helper

import (
	"encoding/json"
)

// WebResponse: amplop baku untuk semua respons (pola Modul 2).
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta: keterangan paginasi — diisi khusus endpoint list.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// errBody: isi bagian errors pada respons gagal.
type errBody struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// Ok menghasilkan (JSON body, HTTP status) untuk respons sukses.
func Ok(status int, message string, data any, meta *Meta) (string, int) {
	return marshal(WebResponse{Success: true, Message: message, Data: data, Meta: meta}), status
}

// Fail menghasilkan (JSON body, HTTP status) untuk respons gagal eksplisit.
func Fail(status int, code, detail string) (string, int) {
	return marshal(WebResponse{
		Success: false,
		Errors:  errBody{Code: code, Detail: detail},
	}), status
}

// FromError menerjemahkan error service menjadi envelope + status.
func FromError(err error) (string, int) {
	ae := AsAppError(err)
	return Fail(ae.Status, ae.Code, ae.Detail)
}

func marshal(e WebResponse) string {
	b, _ := json.Marshal(e) // struct kita sendiri, marshal tidak bisa gagal
	return string(b)
}
