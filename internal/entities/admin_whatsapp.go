package entities

type AdminWhatsApp struct {
	PhoneNumber string `db:"phone_number"`
	TextMessage string `db:"text_message"`
	Description string `db:"description"`
}
