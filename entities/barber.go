package entities

import (
	cfg "barbershop-bot/lib/config"
	tm "barbershop-bot/lib/time"
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
	NoName  = "Без имени"
	NoPhone = "Номер не указан"
)

var infiniteWorkDate time.Time

func init() {
	infWorkDate, err := time.ParseInLocation(time.DateOnly, cfg.InfiniteWorkDate, cfg.Location)
	if err != nil {
		log.Fatal(err)
	}
	infiniteWorkDate = infWorkDate
}

func (b Barber) Info() string {
	var lastWorkDate string
	if b.LastWorkdate.Equal(infiniteWorkDate) {
		lastWorkDate = "бессрочно"
	} else {
		lastWorkDate = "до " + tm.ShowDate(b.LastWorkdate)
	}
	return fmt.Sprintf("Имя: %s\nTел.: %s\nРаботает %s", b.Name, b.Phone, lastWorkDate)
}

func (b Barber) PublicInfo() string {
	return fmt.Sprintf("%s\n[Ссылка на профиль](tg://user?id=%d)", b.Info(), b.ID)
}
