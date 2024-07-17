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

	noWorkingBarbers                = "Извините, услуги временно не предоставляются, так как в настоящий момент в приложении нет ни одного работающего барбера."
	selectBarberForAppointment      = "Выберите барбера, к которому хотите записаться на стрижку."
	selectServiceForAppointment     = "Выберите услугу из списка услуг, предоставляемых барбером %s."
	noFreeTimeForNewAppointmentUser = `Извините, график барбера полностью занят и нет возможности записаться на эту услугу.
Вы можете попробовать связаться с барбером и уточнить у него возможность записи в индивидуальном порядке.
Контакты для связи:
Телефон: %s
[Ссылка на профиль](tg://user?id=%d)`
	selectDateForAppointment  = "Информация о выбранной услуге:\n\n%s\n\nВыберите удобную для Вас дату. Отображены только даты, на которые возможна запись."
	selectTimeForAppointment  = "Информация о выбранной услуге:\n\n%s\n\nВыбранная Вами дата: %s\n\nВыберите удобное для Вас время. Отображено только время, на которое возможна запись."
	confirmNewAppointment     = "Информация о выбранной услуге:\n\n%s\n\nВыбранная Вами дата: %s\nВыбранное время: %s\n\nПодтвердите создание записи или вернитесь в главное меню."
	newAppointmentFailed      = "К сожалению указанное время уже занято. Не удалось создать запись."
	newAppointmentSavedByUser = "Вы записались на услугу:\n\n%s\n\nБарбер %s ждет Вас %s в %s."

	errorUser = `Произошла ошибка обработки команды. Команда не была выполнена. Приносим извинения.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`

	endpntBarberForAppointment         = "barber_for_appointment"
	endpntServiceForAppointmentUser    = "service_for_appointment_user"
	endpntMonthForNewAppointmentUser   = "month_for_new_appointment_user"
	endpntWorkdayForNewAppointmentUser = "workday_for_new_appointment_user"
	endpntTimeForNewAppointmentUser    = "time_for_new_appointment_user"
	endpntBackToMainUser               = "back_to_main_user"
)

var (
	markupMainUser                        = &tele.ReplyMarkup{}
	btnSettingsUser                       = markupEmpty.Data("Настройки", "settings_user")
	btnSignUpForAppointment               = markupEmpty.Data("Записаться на стрижку", "sign_up_for_appointment")
	btnSelectBarberForAppointment         = markupEmpty.Data("", endpntBarberForAppointment)
	btnSelectServiceForAppointmentUser    = markupEmpty.Data("", endpntServiceForAppointmentUser)
	btnSelectMonthForNewAppointmentUser   = markupEmpty.Data("", endpntMonthForNewAppointmentUser)
	btnSelectWorkdayForNewAppointmentUser = markupEmpty.Data("", endpntWorkdayForNewAppointmentUser)
	btnSelectTimeForNewAppointmentUser    = markupEmpty.Data("", endpntTimeForNewAppointmentUser)

	markupConfirmNewAppointmentUser = &tele.ReplyMarkup{}
	btnConfirmNewAppointmentUser    = markupEmpty.Data("Подтвердить запись", "confirm_new_appointment_user")

	markupNewAppointmentFailedUser            = &tele.ReplyMarkup{}
	btnSelectAnotherTimeForNewAppointmentUser = markupEmpty.Data("Выбрать другое время", "select_another_time_for_new_appointment_user")

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
	btnBackToMainUser    = markupEmpty.Data(backToMain, endpntBackToMainUser)

	markupBackToMainUserSend = &tele.ReplyMarkup{}
	btnBackToMainUserSend    = markupEmpty.Data(backToMain, "back_to_main_user_send")
)

func init() {
	markupMainUser.Inline(
		markupEmpty.Row(btnSignUpForAppointment),
		markupEmpty.Row(btnSettingsUser),
	)

	markupConfirmNewAppointmentUser.Inline(
		markupEmpty.Row(btnConfirmNewAppointmentUser),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupNewAppointmentFailedUser.Inline(
		markupEmpty.Row(btnSelectAnotherTimeForNewAppointmentUser),
		markupEmpty.Row(btnBackToMainUser),
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

	markupBackToMainUserSend.Inline(
		markupEmpty.Row(btnBackToMainUserSend),
	)
}

func btnTime(dur tm.Duration, endpnt string) tele.Btn {
	return markupEmpty.Data(dur.ShortString(), endpnt, strconv.FormatUint(uint64(dur), 10))
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

func markupSelectTimeForAppointment(freeTimes []tm.Duration, endpntTime, endpntMonth, endpntBackToMain string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	rowsSelectTime := rowsSelectTimeForAppointment(freeTimes, endpntTime)
	rowBackToSelectWorkday := markup.Row(markup.Data(backToSelectWorkday, endpntMonth, "0"))
	rowBackToMain := markup.Row(markup.Data(backToMain, endpntBackToMain))
	var rows []tele.Row
	rows = append(rows, rowsSelectTime...)
	rows = append(rows, rowBackToSelectWorkday, rowBackToMain)
	markup.Inline(rows...)
	return markup
}

func markupSelectWorkdayForAppointment(
	dateRange ent.DateRange,
	monthRange ent.DateRange,
	appointment sess.Appointment,
	endpntWorkday string,
	endpntMonth string,
	endpntBackToMain string,
) (*tele.ReplyMarkup, error) {
	markup := &tele.ReplyMarkup{}
	btnPrevMonth, btnNextMonth := btnsSwitchMonth(dateRange.EndDate, monthRange, endpntMonth)
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
	for date := dateRange.StartDate; date.Compare(dateRange.EndDate) <= 0; date = date.Add(24 * time.Hour) {
		workday, ok := workdays[date.Day()]
		if !ok {
			btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
		} else {
			if haveFreeTimeForAppointment(workday, appointments[workday.ID], appointment.Duration) {
				btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnWorkday(workday, endpntWorkday))
			} else {
				btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
			}
		}
	}
	for i := 7; i > dateRange.EndWeekday(); i-- {
		btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
	}
	return markupEmpty.Split(7, btnsWorkdaysToSelect), nil
}
