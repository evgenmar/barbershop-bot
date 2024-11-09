package entities

import (
	"fmt"
	"log"
	"time"

	cfg "github.com/evgenmar/barbershop-bot/lib/config"
	tm "github.com/evgenmar/barbershop-bot/lib/time"

	tele "gopkg.in/telebot.v3"
)

type Barber struct {
	ID   int64
	Name string

	//Format of phone number is +71234567890.
	Phone string

	//LastWorkdate is a date in local time zone with HH:MM:SS set to 00:00:00. Default is '3000-01-01'.
	LastWorkdate time.Time
	tele.StoredMessage
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

func (b Barber) Contacts() string {
	return fmt.Sprintf("Контакты для связи:\nТелефон: %s\n[Ссылка на профиль](tg://user?id=%d)", b.Phone, b.ID)
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
