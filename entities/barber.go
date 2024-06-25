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
}

const (
	NoNameBarber  = "Барбер без имени"
	NoPhoneBarber = "Номер не указан"
)

var infiniteWorkDate time.Time

func init() {
	infWorkDate, err := time.ParseInLocation(time.DateOnly, cfg.InfiniteWorkDate, cfg.Location)
	if err != nil {
		log.Fatal(err)
	}
	infiniteWorkDate = infWorkDate
}

func (b Barber) PuplicInfo() string {
	var lastWorkDate string
	if b.LastWorkdate.Equal(infiniteWorkDate) {
		lastWorkDate = "бессрочно"
	} else {
		lastWorkDate = "до " + b.LastWorkdate.Format("02.01.2006")
	}
	return fmt.Sprintf(`Имя: %s
Tел.: %s
Работает %s`, b.Name, b.Phone, lastWorkDate)
}

func (b Barber) PersonalInfo() string {
	return b.PuplicInfo()
}
