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
Перед тем как оставить свои персональные данные, ознакомьтесь с политикой конфиденциальности.

Если Вам будет достаточно, чтобы с Вами связались через telegram, оставлять персональные данные не обязательно.`
	privacyUser = "Текст политики конфиденциальности для клиентов."

	appointmentAlreadyExists           = "Вы уже записаны на услугу:\n\n%s\n\nБарбер %s ждет Вас %s в %s."
	noWorkingBarbers                   = "Извините, услуги временно не предоставляются, так как в настоящий момент в приложении нет ни одного работающего барбера."
	selectBarberForAppointment         = "Выберите барбера, к которому хотите записаться на стрижку."
	userSelectServiceForAppointment    = "Выберите услугу из списка услуг, предоставляемых барбером %s."
	informUserNoFreeTimeForAppointment = `Извините, график барбера полностью занят.
Вы можете попробовать связаться с барбером и уточнить у него возможность записи в индивидуальном порядке.
Контакты для связи:
Телефон: %s
[Ссылка на профиль](tg://user?id=%d)`
	selectDateForAppointment  = "Информация об услуге:\n\n%s\n\nВыберите дату. Отображены только даты, на которые возможна запись."
	selectTimeForAppointment  = "Информация об услуге:\n\n%s\n\nВыбранная Вами дата: %s\n\nВыберите время. Отображено только время, на которое возможна запись."
	confirmNewAppointment     = "Информация об услуге:\n\n%s\n\nВыбранная Вами дата: %s\nВыбранное время: %s\n\nПодтвердите создание записи или вернитесь в главное меню."
	failedToSaveAppointment   = "К сожалению указанное время уже занято. Не удалось создать запись."
	newAppointmentSavedByUser = "Вы записались на услугу:\n\n%s\n\nБарбер %s ждет Вас %s в %s."

	youHaveNoAppointments         = "На данный момент у Вас нет записи, которую можно было бы перенести или отменить."
	rescheduleOrCancelAppointment = "Вы записаны к барберу %s %s в %s на услугу:\n\n%s\n\nПри необходимости перенести или отменить запись выберите соотетствующее действие."
	confirmRescheduleAppointment  = "Информация о переносимой записи:\n\n%s\n\nНовая дата: %s\nНовое время: %s\n\nПодтвердите перенос записи или вернитесь в главное меню."
	appointmentRescheduledByUser  = "Информация о перенесенной записи:\n\n%s\n\nЗапись перенесена на новое время. Барбер %s ждет Вас %s в %s."
	failedToRescheduleAppointment = "К сожалению указанное время уже занято. Не удалось перенести запись."

	errorUser = `Произошла ошибка обработки команды. Команда не была выполнена. Приносим извинения.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`

	endpntBarberForAppointment            = "barber_for_appointment"
	endpntUserSelectServiceForAppointment = "user_select_service_for_appointment"
	endpntUserSelectMonthForAppointment   = "user_select_month_for_appointment"
	endpntUserSelectWorkdayForAppointment = "user_select_workday_for_appointment"
	endpntUserSelectTimeForAppointment    = "user_select_time_for_appointment"
	endpntUserBackToMain                  = "user_back_to_main"
)

var (
	markupUserMain                   = &tele.ReplyMarkup{}
	btnUserSettings                  = markupEmpty.Data("Настройки", "user_settings")
	btnSignUpForAppointment          = markupEmpty.Data("Записаться на стрижку", "sign_up_for_appointment")
	btnRescheduleOrCancelAppointment = markupEmpty.Data("Перенести/отменить запись", "reschedule_or_cancel_appointment")

	btnSelectBarberForAppointment      = markupEmpty.Data("", endpntBarberForAppointment)
	btnUserSelectServiceForAppointment = markupEmpty.Data("", endpntUserSelectServiceForAppointment)
	btnUserSelectMonthForAppointment   = markupEmpty.Data("", endpntUserSelectMonthForAppointment)
	btnUserSelectWorkdayForAppointment = markupEmpty.Data("", endpntUserSelectWorkdayForAppointment)
	btnUserSelectTimeForAppointment    = markupEmpty.Data("", endpntUserSelectTimeForAppointment)

	markupUserConfirmNewAppointment = &tele.ReplyMarkup{}
	btnUserConfirmNewAppointment    = markupEmpty.Data("Подтвердить запись", "user_confirm_new_appointment")

	markupUserFailedToSaveOrRescheduleAppointment = &tele.ReplyMarkup{}
	btnUserSelectAnotherTimeForAppointment        = markupEmpty.Data("Выбрать другое время", "user_select_another_time_for_appointment")

	markupRescheduleOrCancelAppointment = &tele.ReplyMarkup{}
	btnUserRescheduleAppointment        = markupEmpty.Data("Перенести запись", "user_reschedule_appointment")
	btnUserCancelAppointment            = markupEmpty.Data("Отменить запись", "user_cancel_appointment")

	markupUserConfirmRescheduleAppointment = &tele.ReplyMarkup{}
	btnUserConfirmRescheduleAppointment    = markupEmpty.Data("Подтвердить перенос записи", "user_confirm_reschedule_appointment")

	markupUserSettings = &tele.ReplyMarkup{}
	btnUserUpdPersonal = markupEmpty.Data("Обновить персональные данные", "user_upd_personal_data")

	markupPrivacyExplanation = &tele.ReplyMarkup{}
	btnUserPrivacy           = markupEmpty.Data("Политика конфиденциальности", "user_privacy_policy")

	markupUserPrivacy       = &tele.ReplyMarkup{}
	btnUserAgreeWithPrivacy = markupEmpty.Data("Соглашаюсь с политикой конфиденциальности", "user_agree_with_privacy")

	markupUserPersonal = &tele.ReplyMarkup{}
	btnUserUpdName     = markupEmpty.Data("Обновить имя", "user_upd_name")
	btnUserUpdPhone    = markupEmpty.Data("Обновить номер телефона", "user_upd_phone")

	markupUserBackToMain = &tele.ReplyMarkup{}
	btnUserBackToMain    = markupEmpty.Data(backToMain, endpntUserBackToMain)

	markupUserBackToMainSend = &tele.ReplyMarkup{}
	btnUserBackToMainSend    = markupEmpty.Data(backToMain, "user_back_to_main_send")
)

func init() {
	markupUserMain.Inline(
		markupEmpty.Row(btnSignUpForAppointment),
		markupEmpty.Row(btnRescheduleOrCancelAppointment),
		markupEmpty.Row(btnUserSettings),
	)

	markupUserConfirmNewAppointment.Inline(
		markupEmpty.Row(btnUserConfirmNewAppointment),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupUserFailedToSaveOrRescheduleAppointment.Inline(
		markupEmpty.Row(btnUserSelectAnotherTimeForAppointment),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupRescheduleOrCancelAppointment.Inline(
		markupEmpty.Row(btnUserRescheduleAppointment),
		markupEmpty.Row(btnUserCancelAppointment),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupUserConfirmRescheduleAppointment.Inline(
		markupEmpty.Row(btnUserConfirmRescheduleAppointment),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupUserSettings.Inline(
		markupEmpty.Row(btnUserUpdPersonal),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupPrivacyExplanation.Inline(
		markupEmpty.Row(btnUserPrivacy),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupUserPrivacy.Inline(
		markupEmpty.Row(btnUserAgreeWithPrivacy),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupUserPersonal.Inline(
		markupEmpty.Row(btnUserUpdName),
		markupEmpty.Row(btnUserUpdPhone),
		markupEmpty.Row(btnUserBackToMain),
	)

	markupUserBackToMain.Inline(
		markupEmpty.Row(btnUserBackToMain),
	)

	markupUserBackToMainSend.Inline(
		markupEmpty.Row(btnUserBackToMainSend),
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
	rows = append(rows, markup.Row(btnUserBackToMain))
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
			if haveFreeTimeForAppointment(workday, appointments[workday.ID], appointment) {
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
