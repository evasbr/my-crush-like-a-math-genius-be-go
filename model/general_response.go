package model

type GeneralResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}
