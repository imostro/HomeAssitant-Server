package main

type PasswordRequest struct {
	Password string `json:"password"`
	Cert string		`json:"cert"`
}
