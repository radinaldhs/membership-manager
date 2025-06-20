package models

type GetMemberResponse struct {
	ID                   int     `json:"IdMMember"`
	RegistrationBranchID int     `json:"IdMCabangDaftar"`
	Name                 string  `json:"Nama"`
	MemberType           string  `json:"TipeMember"`
	Points               float64 `json:"JumlahPoint"`
	CardNumber           string  `json:"NomorKartu"`
	PhoneNumber          string  `json:"Telpon"`
	Address              string  `json:"Alamat"`
	Province             string  `json:"Propinsi"`
	Regency              string  `json:"Kota"`
	Religion             string  `json:"Agama"`
	Email                string  `json:"Email"`
	EmailVerified        bool    `json:"email_verified"`
	DateOfBirth          string  `json:"TglLahir"`
	Gender               string  `json:"Kelamin"`
}

type UpdateMemberRequest struct {
	PhoneNumber string `json:"Telpon" validate:"phone_num"`
	DateOfBirth string `json:"TglLahir"`
	Gender      string `json:"Kelamin"`
	Address     string `json:"Alamat"`
	Province    string `json:"Propinsi"`
	Regency     string `json:"Kota"`
	Religion    string `json:"Agama"`
}
