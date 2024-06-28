package telegram

import (
	ent "barbershop-bot/entities"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	endpntNoAction = "no_action"
)

var monthNames = []string{
	"Январь",
	"Февраль",
	"Март",
	"Апрель",
	"Май",
	"Июнь",
	"Июль",
	"Август",
	"Сентябрь",
	"Октябрь",
	"Ноябрь",
	"Декабрь",
}

var (
	markupEmpty = &tele.ReplyMarkup{}
	btnEmpty    = markupEmpty.Data("-", endpntNoAction)
	rowWeekdays = tele.Row{
		markupEmpty.Data("Пн", endpntNoAction),
		markupEmpty.Data("Вт", endpntNoAction),
		markupEmpty.Data("Ср", endpntNoAction),
		markupEmpty.Data("Чт", endpntNoAction),
		markupEmpty.Data("Пт", endpntNoAction),
		markupEmpty.Data("Сб", endpntNoAction),
		markupEmpty.Data("Вс", endpntNoAction),
	}
)

func btnBarber(barb ent.Barber, endpnt string) tele.Btn {
	return markupEmpty.Data(barb.Name, endpnt, strconv.FormatInt(barb.ID, 10))
}

func btnDate(date time.Time, endpnt string) tele.Btn {
	return markupEmpty.Data(strconv.Itoa(date.Day()), endpnt, date.Format(time.DateOnly))
}

func btnMonth(month time.Month) tele.Btn {
	return markupEmpty.Data(monthNames[month-1], endpntNoAction)
}

func btnNext(switchTo byte, endpnt string) tele.Btn {
	return markupEmpty.Data("➡", endpnt, strconv.FormatUint(uint64(switchTo), 10))
}

func btnPrev(switchTo byte, endpnt string) tele.Btn {
	return markupEmpty.Data("⬅", endpnt, strconv.FormatUint(uint64(switchTo), 10))
}

func btnService(serv ent.Service, endpnt string) tele.Btn {
	return markupEmpty.Data(serv.BtnSignature(), endpnt, strconv.Itoa(serv.ID))
}

func btnsSwitch(delta, maxDelta byte, endpnt string) (prev, next tele.Btn) {
	if maxDelta == 0 {
		prev = btnEmpty
		next = btnEmpty
		return
	}
	switch delta {
	case 0:
		prev = btnEmpty
		next = btnNext(1, endpnt)
	case maxDelta:
		prev = btnPrev(delta-1, endpnt)
		next = btnEmpty
	default:
		prev = btnPrev(delta-1, endpnt)
		next = btnNext(delta+1, endpnt)
	}
	return
}

func noAction(tele.Context) error { return nil }
