package models

import "github.com/google/uuid"

type User struct {
	Id    uuid.UUID `json:"id" `
	Email string    `json:"email"`
	// PhoneNumber string `json:"phone_number" db:"phone_number"`
	FirstName    string `json:"first_name"`
	SecondName   string `json:"second_name"`
	Description  string `json:"description"`
	Access       string `json:"access"`
	CompanyName  string `json:"company_name"`
	CompanyInfo  string `json:"company_info"`
	CompanyURL   string `json:"company_url"`
	CompanyLogo  string `json:"company_logo"`
	ProfileImage string `json:"profile_image"`
}
