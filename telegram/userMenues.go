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
	privacyExplanation = `Мы просим Вас оставить свое имя и номер телефона, если вы хотите, чтобы в случае возникновения непредвиденных обстоятельств (заболел барбер, отключили свет/воду и т.п.) Вам позвонили на телефон и предупредили.
Если Вам будет достаточно, чтобы с Вами связались через telegram, оставлять персональные данные не нужно.
Перед тем как оставить свои персональные данные, ознакомьтесь с политикой конфиденциальности.`
	privacyUser = "Текст политики конфиденциальности для клиентов."

	noWorkingBarbers            = "Извините, услуги временно не предоставляются, так как в настоящий момент в приложении нет ни одного работающего барбера."
	selectBarberForAppointment  = "Выберите барбера, к которому хотите записаться на стрижку."
	selectServiceForAppointment = "Выберите услугу из списка услуг, предоставляемых барбером %s."
	noFreeTimeForAppointment    = `Извините, график барбера полностью занят и нет возможности записаться на эту услугу.
Вы можете попробовать связаться с барбером и уточнить у него возможность записи в индивидуальном порядке.
Контакты для связи:
Телефон: %s
[Ссылка на профиль](tg://user?id=%d)`
	selectDateForAppointment = "Информация о выбранной услуге:\n\n%s\n\nВыберите удобную для Вас дату. Отображены только даты, на которые возможна запись."

	errorUser = `Произошла ошибка обработки команды. Команда не была выполнена. Приносим извинения.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`

	endpntBarberForAppointment  = "barber_for_appointment"
	endpntServiceForAppointment = "service_for_appointment"
	endpntMonthForAppointment   = "month_for_appointment"
	endpntWorkdayForAppointment = "workday_for_appointment"
)

var (
	markupMainUser                 = &tele.ReplyMarkup{}
	btnSettingsUser                = markupEmpty.Data("Настройки", "settings_user")
	btnSignUpForAppointment        = markupEmpty.Data("Записаться на стрижку", "sign_up_for_appointment")
	btnSelectBarberForAppointment  = markupEmpty.Data("", endpntBarberForAppointment)
	btnSelectServiceForAppointment = markupEmpty.Data("", endpntServiceForAppointment)
	btnSelectMonthForAppointment   = markupEmpty.Data("", endpntMonthForAppointment)
	btnSelectWorkdayForAppointment = markupEmpty.Data("", endpntWorkdayForAppointment)

	markupSettingsUser = &tele.ReplyMarkup{}
	btnUpdPersonalUser = markupEmpty.Data("Обновить персональные данные", "upd_personal_data_user")

	markupPrivacyExplanation = &tele.ReplyMarkup{}
	btnPrivacyUser           = markupEmpty.Data("Политика конфиденциальности", "privacy_policy_user")

	markupPrivacyUser       = &tele.ReplyMarkup{}
	btnUserAgreeWithPrivacy = markupEmpty.Data("Соглашаюсь с политикой конфиденциальности", "user_agree_with_privacy")

	markupPersonalUser = &tele.ReplyMarkup{}
	btnUpdNameUser     = markupEmpty.Data("Обновить имя", "upd_name_user")
	btnUpdPhoneUser    = markupEmpty.Data("Обновить номер телефона", "upd_phone_user")

	markupBackToMainUser = &tele.ReplyMarkup{}
	btnBackToMainUser    = markupEmpty.Data("Вернуться в главное меню", "back_to_main_user")
)

func init() {
	markupMainUser.Inline(
		markupEmpty.Row(btnSignUpForAppointment),
		markupEmpty.Row(btnSettingsUser),
	)

	markupSettingsUser.Inline(
		markupEmpty.Row(btnUpdPersonalUser),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupPrivacyExplanation.Inline(
		markupEmpty.Row(btnPrivacyUser),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupPrivacyUser.Inline(
		markupEmpty.Row(btnUserAgreeWithPrivacy),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupPersonalUser.Inline(
		markupEmpty.Row(btnUpdNameUser),
		markupEmpty.Row(btnUpdPhoneUser),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupBackToMainUser.Inline(
		markupEmpty.Row(btnBackToMainUser),
	)
}

func btnWorkday(workday ent.Workday, endpnt string) tele.Btn {
	return markupEmpty.Data(strconv.Itoa(workday.Date.Day()), endpnt, strconv.Itoa(workday.ID))
}

func markupSelectBarberForAppointment(barbers []ent.Barber) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, barber := range barbers {
		rows = append(rows, markup.Row(btnBarber(barber, endpntBarberForAppointment)))
	}
	rows = append(rows, markup.Row(btnBackToMainUser))
	markup.Inline(rows...)
	return markup
}

func markupSelectWorkdayForAppointment(dateRange ent.DateRange, firstMonthEnd, lastMonthEnd time.Time, newAppointment sess.NewAppointment) (*tele.ReplyMarkup, error) {
	markup := &tele.ReplyMarkup{}
	btnPrevMonth, btnNextMonth := btnsSwitchMonth(dateRange.EndDate, firstMonthEnd, lastMonthEnd, endpntMonthForAppointment)
	rowSelectMonth := markup.Row(btnPrevMonth, btnMonth(dateRange.Month()), btnNextMonth)
	rowsSelectWorkday, err := rowsSelectWorkdayForAppointment(dateRange, newAppointment)
	if err != nil {
		return nil, err
	}
	rowBackToMainUser := markup.Row(btnBackToMainUser)
	var rows []tele.Row
	rows = append(rows, rowSelectMonth, rowWeekdays)
	rows = append(rows, rowsSelectWorkday...)
	rows = append(rows, rowBackToMainUser)
	markup.Inline(rows...)
	return markup, nil
}

func markupSelectServiceForAppointment(services []ent.Service) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, service := range services {
		row := markup.Row(btnService(service, endpntServiceForAppointment))
		rows = append(rows, row)
	}
	rows = append(rows, markup.Row(btnBackToMainUser))
	markup.Inline(rows...)
	return markup
}

func rowsSelectWorkdayForAppointment(dateRange ent.DateRange, newAppointment sess.NewAppointment) ([]tele.Row, error) {
	markup := &tele.ReplyMarkup{}
	wds, err := cp.RepoWithContext.GetWorkdaysByDateRange(newAppointment.BarberID, dateRange)
	if err != nil {
		return nil, err
	}
	workdays := make(map[int]ent.Workday)
	for _, wd := range wds {
		workdays[wd.Date.Day()] = wd
	}
	appts, err := cp.RepoWithContext.GetAppointmentsByDateRange(newAppointment.BarberID, dateRange)
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
	for date := dateRange.StartDate; date.Compare(dateRange.EndDate) <= 0; date = date.Add(24 * time.Hour) {
		workday, ok := workdays[date.Day()]
		if !ok {
			btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
		} else {
			if haveFreeTimeForAppointment(workday, appointments[workday.ID], newAppointment.Duration) {
				btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnWorkday(workday, endpntWorkdayForAppointment))
			} else {
				btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
			}
		}
	}
	for i := 7; i > dateRange.EndWeekday(); i-- {
		btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
	}
	return markup.Split(7, btnsWorkdaysToSelect), nil
}

func haveFreeTimeForAppointment(workday ent.Workday, appointments []ent.Appointment, duration tm.Duration) bool {
	var analyzedTime tm.Duration
	if workday.Date.Equal(tm.Today()) {
		currentDayTime := tm.CurrentDayTime()
		if currentDayTime > workday.StartTime {
			analyzedTime = currentDayTime
		} else {
			analyzedTime = workday.StartTime
		}
	} else {
		analyzedTime = workday.StartTime
	}
	for _, appointment := range appointments {
		if (appointment.Time - analyzedTime) >= duration {
			return true
		}
		analyzedTime = appointment.Time + appointment.Duration
	}
	return (workday.EndTime - analyzedTime) >= duration
}
