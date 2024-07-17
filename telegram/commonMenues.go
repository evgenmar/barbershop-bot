package telegram

import (
	ent "barbershop-bot/entities"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	mainMenu = "Добрый день. Вы находитесь в главном меню."

	settingsMenu = "В этом меню собраны функции, обеспечивающие управление и настройки приложения."

	selectPersonalData = "Выберите данные, которые вы хотите обновить."

	updName        = "Введите Ваше имя"
	updNameSuccess = "Имя успешно обновлено. Хотите обновить другие данные?"
	invalidName    = `Введенное имя не соответствует установленным критериям:
- имя может содержать русские и английские буквы, цифры и пробелы;
- длина имени должна быть не менее 2 символов и не более 20 символов.
Пожалуйста, попробуйте ввести имя еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	updPhone        = "Введите свой номер телефона"
	updPhoneSuccess = "Номер телефона успешно обновлен. Хотите обновить другие данные?"
	invalidPhone    = `Неизвестный формат номера телефона. Примеры поддерживаемых форматов:
81234567890
8(123)-456-7890
+71234567890
+7 123 456 7890
Номер должен начинаться с "+7" или с "8". Региональный код можно по желанию заключить в скобки, также допустимо разделять пробелами или тире группы цифр.
Пожалуйста, попробуйте ввести номер телефона еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`

	backToSelectWorkday = "Назад к выбору даты"
	backToMain          = "Вернуться в главное меню"

	unknownCommand = "Неизвестная команда. Пожалуйста, выберите команду из меню. Для вызова главного меню воспользуйтесь командой /start"

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

func btnMonth(month time.Month) tele.Btn {
	return markupEmpty.Data(monthNames[month-1], endpntNoAction)
}

func btnNext(endpnt string) tele.Btn {
	return markupEmpty.Data("➡", endpnt, strconv.FormatInt(1, 10))
}

func btnPrev(endpnt string) tele.Btn {
	return markupEmpty.Data("⬅", endpnt, strconv.FormatInt(-1, 10))
}

func btnService(serv ent.Service, endpnt string) tele.Btn {
	return markupEmpty.Data(serv.BtnSignature(), endpnt, strconv.Itoa(serv.ID))
}

func btnsSwitchMonth(current time.Time, period ent.DateRange, endpnt string) (prev, next tele.Btn) {
	if period.StartDate.Equal(period.EndDate) {
		prev = btnEmpty
		next = btnEmpty
		return
	}
	switch current {
	case period.StartDate:
		prev = btnEmpty
		next = btnNext(endpnt)
	case period.EndDate:
		prev = btnPrev(endpnt)
		next = btnEmpty
	default:
		prev = btnPrev(endpnt)
		next = btnNext(endpnt)
	}
	return
}

func markupSelectService(services []ent.Service, endpntService, endpntBackToMain string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, service := range services {
		row := markup.Row(btnService(service, endpntService))
		rows = append(rows, row)
	}
	rows = append(rows, markup.Row(markup.Data(backToMain, endpntBackToMain)))
	markup.Inline(rows...)
	return markup
}
