package telegram

import (
	ent "barbershop-bot/entities"

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
	newAppointmentSavedByUser = "Вы записались на услугу:\n\n%s\n\nБарбер %s ждет Вас %s в %s."

	youHaveNoAppointments         = "На данный момент у Вас нет записи, которую можно было бы перенести или отменить."
	rescheduleOrCancelAppointment = "Вы записаны к барберу %s %s в %s на услугу:\n\n%s\n\nПри необходимости перенести или отменить запись выберите соотетствующее действие."
	appointmentRescheduledByUser  = "Информация о перенесенной записи:\n\n%s\n\nЗапись перенесена на новое время. Барбер %s ждет Вас %s в %s."
	confirmCancelAppointment      = "Вы записаны к барберу %s %s в %s на услугу:\n\n%s\n\nПодтвердите отмену этой записи или вернитесь в главное меню."
	appointmentCanceled           = "Запись удалена."

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

	markupUserConfirmNewAppointment = &tele.ReplyMarkup{}
	btnUserConfirmNewAppointment    = markupEmpty.Data("Подтвердить запись", "user_confirm_new_appointment")

	markupUserFailedToSaveOrRescheduleAppointment = &tele.ReplyMarkup{}
	btnUserSelectAnotherTimeForAppointment        = markupEmpty.Data("Выбрать другое время", "user_select_another_time_for_appointment")

	markupRescheduleOrCancelAppointment = &tele.ReplyMarkup{}
	btnUserRescheduleAppointment        = markupEmpty.Data("Перенести запись", "user_reschedule_appointment")
	btnUserCancelAppointment            = markupEmpty.Data("Отменить запись", "user_cancel_appointment")

	markupUserConfirmRescheduleAppointment = &tele.ReplyMarkup{}
	btnUserConfirmRescheduleAppointment    = markupEmpty.Data("Подтвердить перенос записи", "user_confirm_reschedule_appointment")

	markupUserConfirmCancelAppointment = &tele.ReplyMarkup{}
	btnUserConfirmCancelAppointment    = markupEmpty.Data("Подтвердить отмену записи", "user_confirm_cancel_appointment")

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

	markupUserConfirmCancelAppointment.Inline(
		markupEmpty.Row(btnUserConfirmCancelAppointment),
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
