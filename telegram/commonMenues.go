package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	tm "barbershop-bot/lib/time"
	sess "barbershop-bot/sessions"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	mainMenu = "Добрый день. Вы находитесь в главном меню."

	selectDateForAppointment = "Информация об услуге:\n\n%s\n\nВыберите дату. Отображены только даты, на которые возможна запись."
	selectTimeForAppointment = "Информация об услуге:\n\n%s\n\nВыбранная Вами дата: %s\n\nВыберите время. Отображено только время, на которое возможна запись."
	confirmNewAppointment    = "Информация об услуге:\n\n%s\n\nВыбранная Вами дата: %s\nВыбранное время: %s\n\nПодтвердите создание записи или вернитесь в главное меню."
	failedToSaveAppointment  = "К сожалению указанное время уже занято. Не удалось создать запись."

	confirmRescheduleAppointment = "Информация о переносимой записи:\n\n%s\n\nНовая дата: %s\nНовое время: %s\n\nПодтвердите перенос записи или вернитесь в главное меню."

	settingsMenu = "В этом меню собраны функции, обеспечивающие управление и настройки приложения."

	selectPersonalData = "Выберите данные, которые вы хотите обновить."

	updName        = "Введите Ваше имя"
	updNameSuccess = "Имя успешно обновлено. Хотите обновить другие данные?"
	invalidName    = `Введенное имя не соответствует установленным критериям:
- имя может содержать русские и английские буквы (минимум 1 буква в имени), цифры и пробелы;
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

	backToMain = "Вернуться в главное меню"

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
	btnEmpty    = markupEmpty.Data(" ", endpntNoAction)
	btnDash     = markupEmpty.Data("-", endpntNoAction)
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

func btnsSwitchMonth(current tm.Month, period monthRange, endpnt string) (prev, next tele.Btn) {
	if period.firstMonth == period.lastMonth {
		prev = btnEmpty
		next = btnEmpty
		return
	}
	switch current {
	case period.firstMonth:
		prev = btnEmpty
		next = btnNext(endpnt)
	case period.lastMonth:
		prev = btnPrev(endpnt)
		next = btnEmpty
	default:
		prev = btnPrev(endpnt)
		next = btnNext(endpnt)
	}
	return
}

func btnTime(dur tm.Duration, endpnt string) tele.Btn {
	return markupEmpty.Data(dur.ShortString(), endpnt, strconv.FormatUint(uint64(dur), 10))
}

func btnWorkday(workday ent.Workday, endpnt string) tele.Btn {
	return markupEmpty.Data(strconv.Itoa(workday.Date.Day()), endpnt, strconv.Itoa(workday.ID))
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

func markupSelectTimeForAppointment(freeTimes []tm.Duration, endpntTime, endpntMonth, endpntBackToMain string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	rowsSelectTime := rowsSelectTimeForAppointment(freeTimes, endpntTime)
	rowBackToSelectWorkday := markup.Row(markup.Data("Назад к выбору даты", endpntMonth, "0"))
	rowBackToMain := markup.Row(markup.Data(backToMain, endpntBackToMain))
	var rows []tele.Row
	rows = append(rows, rowsSelectTime...)
	rows = append(rows, rowBackToSelectWorkday, rowBackToMain)
	markup.Inline(rows...)
	return markup
}

func markupSelectWorkdayForAppointment(
	dateRange ent.DateRange,
	monthRange monthRange,
	appointment sess.Appointment,
	endpntWorkday string,
	endpntMonth string,
	endpntBackToMain string,
) (*tele.ReplyMarkup, error) {
	markup := &tele.ReplyMarkup{}
	btnPrevMonth, btnNextMonth := btnsSwitchMonth(tm.ParseMonth(dateRange.LastDate), monthRange, endpntMonth)
	rowSelectMonth := markup.Row(btnPrevMonth, btnMonth(dateRange.Month()), btnNextMonth)
	rowsSelectWorkday, err := rowsSelectWorkdayForAppointment(dateRange, appointment, endpntWorkday)
	if err != nil {
		return nil, err
	}
	rowBackToMain := markup.Row(markup.Data(backToMain, endpntBackToMain))
	var rows []tele.Row
	rows = append(rows, rowSelectMonth, rowWeekdays)
	rows = append(rows, rowsSelectWorkday...)
	rows = append(rows, rowBackToMain)
	markup.Inline(rows...)
	return markup, nil
}

func rowsSelectTimeForAppointment(freeTimes []tm.Duration, endpnt string) []tele.Row {
	var btnsTimesToSelect []tele.Btn
	for _, freeTime := range freeTimes {
		btnsTimesToSelect = append(btnsTimesToSelect, btnTime(freeTime, endpnt))
	}
	for i := 1; i <= (4-len(freeTimes)%4)%4; i++ {
		btnsTimesToSelect = append(btnsTimesToSelect, btnEmpty)
	}
	return markupEmpty.Split(4, btnsTimesToSelect)
}

func rowsSelectWorkdayForAppointment(dateRange ent.DateRange, appointment sess.Appointment, endpntWorkday string) ([]tele.Row, error) {
	wds, err := cp.RepoWithContext.GetWorkdaysByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return nil, err
	}
	workdays := make(map[int]ent.Workday)
	for _, wd := range wds {
		workdays[wd.Date.Day()] = wd
	}
	appts, err := cp.RepoWithContext.GetAppointmentsByDateRange(appointment.BarberID, dateRange)
	if err != nil {
		return nil, err
	}
	appointments := make(map[int][]ent.Appointment)
	for _, appt := range appts {
		appointments[appt.WorkdayID] = append(appointments[appt.WorkdayID], appt)
	}
	var btnsWorkdaysToSelect []tele.Btn
	for i := 1; i < dateRange.StartWeekday(); i++ {
		btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
	}
	for date := dateRange.FirstDate; date.Compare(dateRange.LastDate) <= 0; date = date.Add(24 * time.Hour) {
		workday, ok := workdays[date.Day()]
		if !ok {
			btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnDash)
		} else {
			if workdayHaveFreeTimeForAppointment(workday, appointments[workday.ID], appointment) {
				btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnWorkday(workday, endpntWorkday))
			} else {
				btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnDash)
			}
		}
	}
	for i := 7; i > dateRange.EndWeekday(); i-- {
		btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
	}
	return markupEmpty.Split(7, btnsWorkdaysToSelect), nil
}
