package entities

type Barber struct {
	ID   int64
	Name string

	//Format of phone number is +71234567890
	Phone string
	Status
}

const (
	NoNameBarber  = "Барбер без имени"
	NoPhoneBarber = "Номер не указан"
)
