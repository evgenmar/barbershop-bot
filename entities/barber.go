package entities

import (
	cfg "barbershop-bot/lib/config"
	"fmt"
	"log"
	"time"
)

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

var endlessWorkdate time.Time

func init() {
	year3000, err := time.ParseInLocation(time.DateOnly, "3000-01-01", cfg.Location)
	if err != nil {
		log.Fatal(err)
	}
	endlessWorkdate = year3000
}

func (b Barber) Info() string {
	var lastWorkDate string
	if b.LastWorkdate.Equal(endlessWorkdate) {
		lastWorkDate = "бессрочно"
	} else {
		lastWorkDate = "до " + b.LastWorkdate.Format("02.01.2006")
	}
	return fmt.Sprintf(`Имя: %s
Tел.: %s
Работает %s`, b.Name, b.Phone, lastWorkDate)
}

func (b Barber) Settings() string {
	return b.Info()
}
