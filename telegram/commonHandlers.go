package telegram

import (
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	prev           = "⬅"
	next           = "➡"
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

func btnMonth(month time.Month) tele.Btn {
	return markupEmpty.Data(monthNames[month-1], endpntNoAction)
}

func noAction(tele.Context) error { return nil }
