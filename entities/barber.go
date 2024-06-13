package entities

import "time"

type Barber struct {
	ID   int64
	Name string

	//Format of phone number is +71234567890.
	Phone string

	//LastWorkdate is a date in local time zone with HH:MM:SS set to 00:00:00. Default is '3000-01-01'.
	LastWorkdate time.Time
	Status
}

const (
	NoNameBarber  = "Барбер без имени"
	NoPhoneBarber = "Номер не указан"
)
