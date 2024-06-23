package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	rep "barbershop-bot/repository"
	sched "barbershop-bot/scheduler"
	"errors"
	"log"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	mainBarber = "Добрый день. Вы находитесь в главном меню. Выберите действие."

	settingsBarber = "Вы находитесь меню настроек. Выберите действие."

	manageAccountBarber = "В этом меню собраны функции для управления Вашим аккаунтом. Выберите действие."
	currentSettings     = `Ваши текущие настройки:
`
	personalBarber       = "Выберите данные, которые вы хотите обновить."
	updNameBarber        = "Введите Ваше имя"
	updNameSuccessBarber = "Имя успешно обновлено. Хотите обновить другие данные?"
	invalidBarberName    = `Введенное имя не соответствует установленным критериям:
		- имя может содержать русские и английские буквы, цифры и пробелы;
		- длина имени должна быть не менее 2 символов и не более 20 символов.
Пожалуйста, попробуйте ввести имя еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberName = `Имя не сохранено. Другой барбер с таким именем уже зарегистрирован в приложении. Введите другое имя.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	updPhoneBarber        = "Введите свой номер телефона"
	updPhoneSuccessBarber = "Номер телефона успешно обновлен. Хотите обновить другие данные?"
	invalidBarberPhone    = `Неизвестный формат номера телефона. Примеры поддерживаемых форматов:
		81234567890
		8(123)-456-7890
		+71234567890
		+7 123 456 7890
Номер должен начинаться с "+7" или с "8". Региональный код можно по желанию заключить в скобки, также допустимо разделять пробелами или тире группы цифр.
Пожалуйста, попробуйте ввести номер телефона еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberPhone = `Номер телефона не сохранен. Другой барбер с таким номером уже зарегистрирован в приложении. Введите другой номер.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	endpntSelectMonthOfLastWorkDate = "select_month_of_last_work_date"
	endpntSelectLastWorkDate        = "select_last_work_date"
	selectLastWorkDate              = `Данную функцию следует использовать, если Вы планируете прекратить использовать этот бот в своей работе и хотите, чтобы клиенты перестали использовать бот для записи к Вам на стрижку.
	
Выберите дату последнего рабочего дня.
Для выбора доступны только даты позже последней существующей записи клиента на стрижку.
Если Вы хотите выбрать более раннюю дату, сначала отмените последние записи клиентов на стрижку.
	
ВНИМАНИЕ!!! После установки даты последнего рабочего дня клиенты не смогут записаться на стрижку на более позднюю дату.
По умолчанию установлена бессрочная дата последнего рабочего дня.
Если Вы ранее меняли эту дату и хотите снова установить бессрочную дату, нажмите кнопку в нижней части меню.`
	haveAppointmentAfterDataToSave = `Невозможно сохранить дату, поскольку пока Вы выбирали дату, клиент успел записаться к Вам на стрижку на дату позже выбранной Вами даты.
Пожалуйста проверьте записи клиентов и при необходимости отмените самую позднюю запись. Или выберите более позднюю дату последнего рабочего дня.`
	lastWorkDateSaved               = "Новая дата последнего рабочего дня успешно сохранена."
	lastWorkDateUnchanged           = "Указанная дата совпадает с той, которую вы уже установили ранее"
	lastWorkDateSavedWithoutShedule = `Новая дата последнего рабочего дня сохранена.
ВНИМАНИЕ!!! При попытке составить расписание работы вплоть до сохраненной даты произошла ошибка. Расписание не составлено!
Для доступа к записи клиентов на стрижку необходимо составить расписание работы.`
	confirmSelfDeletion = `Вы собираетесь отказаться от статуса "барбер".
ВНИМАНИЕ!!! Помимо изменения Вашего статуса также будет удален весь перечень оказываемых Вами услуг и вся история прошедших записей Ваших клиентов. Клиенты больше не смогут записаться к Вам на стрижку через этот бот.
Если Вы уверены, нажмите "Уверен, удалить!". Если передумали, просто вернитесь в главное меню.`
	youHaveActiveSchedule = `Невозможно отказаться от статуса "барбер" прямо сейчас. Предварительно вы должны выполнить следующие действия:
`
	goodbuyBarber = "Ваш статус изменен. Спасибо, что работали с нами!"

	manageBarbers = "В этом меню Вы можете добавить нового барбера или удалить существующего. Выберите действие."
	listOfBarbers = "Список всех барберов, зарегистрированных в приложении:"
	addBarber     = `Для добавления нового барбера пришлите в этот чат контакт профиля пользователя телеграм, которого вы хотите сделать барбером.
Подробная инструкция:
1. Зайдите в личный чат с пользователем.
2. В верхней части чата нажмите на поле, отображающее аватар аккаунта пользователя и его имя. Таким образом Вы откроете окно просмотра профиля пользователя.
3. Нажмите на "три точки" в верхнем правом углу дисплея.
4. В открывшемся меню выберите "Поделиться контактом".
5. В открывшемся списке Ваших чатов выберите чат с ботом (чат, в котором Вы читаете эту инструкцию).
6. В правом нижнем углу нажмите на значек отправки сообщения.`
	userIsAlreadyBarber       = "Статус пользователя не изменен, поскольку он уже является барбером."
	addedNewBarberWithShedule = `Статус пользователя изменен на "барбер". Для нового барбера составлено расписание работы на ближайшие полгода.
Для доступа к записи клиентов на стрижку новый барбер должен заполнить персональные данные.`
	addedNewBarberWithoutShedule = `Статус пользователя изменен на "барбер".
ВНИМАНИЕ!!! При попытке составить расписание работы для нового барбера произошла ошибка. Расписание не составлено!
Для доступа к записи клиентов на стрижку новый барбер должен заполнить персональные данные, а также составить расписание работы.`
	onlyOneBarberExists = "Вы единственный зарегистрированный барбер в приложении. Некого удалять."
	noBarbersToDelete   = `Нет ни одного барбера, которого можно было бы удалить. 
Для того, чтобы барбера можно было удалить, он предварительно должен выполнить следующие действия:
`
	selectBarberToDeletion = `Выберите барбера, которого Вы хотите удалить.
ВНИМАНИЕ!!! При удалении барбера будет также удален весь перечень оказываемых им услуг и вся история прошедших записей клиентов этого барбера. 
Если нужного барбера нет в этом списке, значит он еще не выполнил необходимые действия перед удалением:
`
	preDeletionBarberInstruction = `1. Установить не бессрочную дату последнего рабочего дня. По умолчанию установлена бессрочная дата последнего рабочего дня.
2. Дождаться наступления следующего дня после даты установленной в прошлом пункте.
	
Эти действия необходимы, чтобы гарантировать, что ни один клиент, записавшийся на стрижку, не останется необслуженным в процессе удаления барбера из приложения.`
	endpntBarberToDeletion   = "barber_to_deletion"
	barberDeleted            = `Статус указанного пользователя изменен. Пользователь больше не имеет статуса "барбер".`
	barberHaveActiveSchedule = `Невозможно изменить статус барбера, поскольку пока Вы выбирали барбера для удаления, выбранный барбер изменил свою дату последнего рабочего дня на более позднюю.`

	unknownCmdBarber = "Неизвестная команда. Пожалуйста, выберите команду из меню. Для вызова главного меню воспользуйтесь командой /start"
	errorBarber      = `Произошла ошибка обработки команды. Команда не была выполнена. Если ошибка будет повторяться, возможно, потребуется перезапуск сервиса.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`
)

var (
	markupMainBarber  = &tele.ReplyMarkup{}
	btnSettingsBarber = markupMainBarber.Data("Настройки", "settings_barber")

	markupSettingsBarber   = &tele.ReplyMarkup{}
	btnManageAccountBarber = markupSettingsBarber.Data("Управление аккаунтом", "manage_account_barber")
	btnManageBarbers       = markupSettingsBarber.Data("Управление барберами", "manage_barbers")

	markupManageAccountBarber    = &tele.ReplyMarkup{}
	btnShowCurrentSettingsBarber = markupManageAccountBarber.Data("Посмотреть мои текущие настройки", "show_current_settings_barber")
	btnUpdPersonalBarber         = markupManageAccountBarber.Data("Обновить персональные данные", "upd_personal_data_barber")
	btnSetLastWorkDate           = markupManageAccountBarber.Data("Установить последний рабочий день", endpntSelectMonthOfLastWorkDate, "0")
	btnSelectLastWorkDate        = markupManageAccountBarber.Data("", endpntSelectLastWorkDate)
	btnSelfDeleteBarber          = markupManageAccountBarber.Data(`Отказаться от статуса "барбер"`, "self_delete_barber")

	markupPersonalBarber = &tele.ReplyMarkup{}
	btnUpdNameBarber     = markupPersonalBarber.Data("Обновить имя", "upd_name_barber")
	btnUpdPhoneBarber    = markupPersonalBarber.Data("Обновить номер телефона", "upd_phone_barber")

	markupConfirmSelfDeletion = &tele.ReplyMarkup{}
	btnSureToDelete           = markupConfirmSelfDeletion.Data("Уверен, удалить!", "sure_to_delete")

	markupManageBarbers    = &tele.ReplyMarkup{}
	btnShowAllBurbers      = markupManageBarbers.Data("Посмотреть список барберов", "show_all_barbers")
	btnAddBarber           = markupManageBarbers.Data("Добавить барбера", "add_barber")
	btnDeleteBarber        = markupManageBarbers.Data("Удалить барбера", "delete_barber")
	btnDeleteCertainBarber = markupManageBarbers.Data("", endpntBarberToDeletion)

	markupBackToMainBarber = &tele.ReplyMarkup{}
	btnBackToMainBarber    = markupBackToMainBarber.Data("Вернуться в главное меню", "back_to_main_barber")
)

func init() {
	markupMainBarber.Inline(
		markupMainBarber.Row(btnSettingsBarber),
	)

	markupSettingsBarber.Inline(
		markupSettingsBarber.Row(btnManageAccountBarber),
		markupSettingsBarber.Row(btnManageBarbers),
		markupSettingsBarber.Row(btnBackToMainBarber),
	)

	markupManageAccountBarber.Inline(
		markupManageAccountBarber.Row(btnShowCurrentSettingsBarber),
		markupManageAccountBarber.Row(btnUpdPersonalBarber),
		markupManageAccountBarber.Row(btnSetLastWorkDate),
		markupManageAccountBarber.Row(btnSelfDeleteBarber),
		markupManageAccountBarber.Row(btnBackToMainBarber),
	)

	markupPersonalBarber.Inline(
		markupPersonalBarber.Row(btnUpdNameBarber),
		markupPersonalBarber.Row(btnUpdPhoneBarber),
		markupPersonalBarber.Row(btnBackToMainBarber),
	)

	markupConfirmSelfDeletion.Inline(
		markupConfirmSelfDeletion.Row(btnSureToDelete),
		markupConfirmSelfDeletion.Row(btnBackToMainBarber),
	)

	markupManageBarbers.Inline(
		markupManageBarbers.Row(btnShowAllBurbers),
		markupManageBarbers.Row(btnAddBarber),
		markupManageBarbers.Row(btnDeleteBarber),
		markupManageBarbers.Row(btnBackToMainBarber),
	)

	markupBackToMainBarber.Inline(
		markupBackToMainBarber.Row(btnBackToMainBarber),
	)
}

func onStartBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, "can't open the barber's main menu", err)
	}
	return ctx.Send(mainBarber, markupMainBarber)
}

func onSettingsBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, "can't open the barber settings menu", err)
	}
	return ctx.Edit(settingsBarber, markupSettingsBarber)
}

func onManageAccountBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, "can't open the barber's manage account menu", err)
	}
	return ctx.Edit(manageAccountBarber, markupManageAccountBarber)
}

func onShowCurrentSettingsBarber(ctx tele.Context) error {
	errMsg := "can't show current barber settings"
	barberID := ctx.Sender().ID
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	return ctx.Edit(currentSettings+barber.Settings(), markupBackToMainBarber)
}

func onUpdPersonalBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, "can't open the barber's personal data menu", err)
	}
	return ctx.Edit(personalBarber, markupPersonalBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.NewStatus(ent.StateUpdName)}); err != nil {
		return logAndMsgErrBarber(ctx, "can't ask barber to enter name", err)
	}
	return ctx.Edit(updNameBarber)
}

func onUpdPhoneBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.NewStatus(ent.StateUpdPhone)}); err != nil {
		return logAndMsgErrBarber(ctx, "can't ask barber to enter phone", err)
	}
	return ctx.Edit(updPhoneBarber)
}

func onSetLastWorkDate(ctx tele.Context) error {
	errMsg := "can't open select last work date menu"
	barberID := ctx.Sender().ID
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	latestAppointmentDate, err := cp.RepoWithContext.GetLatestAppointmentDate(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	var firstDisplayedDateRange ent.DateRange
	relativeFirstDisplayedMonth := 0
	for relativeFirstDisplayedMonth <= cfg.MaxAppointmentBookingMonths {
		firstDisplayedDateRange = ent.Month(byte(relativeFirstDisplayedMonth))
		if latestAppointmentDate.Compare(firstDisplayedDateRange.EndDate) <= 0 {
			if latestAppointmentDate.After(firstDisplayedDateRange.StartDate) {
				firstDisplayedDateRange.StartDate = latestAppointmentDate
			}
			break
		}
		relativeFirstDisplayedMonth++
	}
	deltaDisplayedMonth, err := strconv.Atoi(ctx.Callback().Data)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	markupSelectDate := markupSelectLastWorkDate(firstDisplayedDateRange, relativeFirstDisplayedMonth, deltaDisplayedMonth)
	return ctx.Edit(selectLastWorkDate, markupSelectDate)
}

func onSelectLastWorkDate(ctx tele.Context) error {
	errMsg := "can't save last work date"
	dateToSave, err := time.ParseInLocation(time.DateOnly, ctx.Callback().Data, cfg.Location)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberID := ctx.Sender().ID
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	switch dateToSave.Compare(barber.LastWorkdate) {
	case 0:
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Status: ent.StatusStart}); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		return ctx.Edit(lastWorkDateUnchanged, markupBackToMainBarber)
	case 1:
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave, Status: ent.StatusStart}); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := sched.MakeSchedule(barberID); err != nil {
			log.Print(e.Wrap(errMsg, err))
			return ctx.Send(lastWorkDateSavedWithoutShedule, markupBackToMainBarber)
		}
		return ctx.Edit(lastWorkDateSaved, markupBackToMainBarber)
	case -1:
		latestWorkDate, err := cp.RepoWithContext.GetLatestWorkDate(barberID)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		dateRangeToDelete := ent.DateRange{StartDate: dateToSave.Add(24 * time.Hour), EndDate: latestWorkDate}
		err = cp.RepoWithContext.DeleteWorkdaysByDateRange(barberID, dateRangeToDelete)
		if err != nil && !errors.Is(err, rep.ErrInvalidDateRange) {
			if errors.Is(err, rep.ErrAppointmentsExists) {
				return ctx.Edit(haveAppointmentAfterDataToSave, markupBackToMainBarber)
			}
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, LastWorkdate: dateToSave, Status: ent.StatusStart}); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		return ctx.Edit(lastWorkDateSaved, markupBackToMainBarber)
	default:
		return nil
	}
}

func onSelfDeleteBarber(ctx tele.Context) error {
	errMsg := "can't provide options for barber self deletion"
	barberID := ctx.Sender().ID
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberToDelete, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if barberToDelete.LastWorkdate.Before(tm.Today()) {
		return ctx.Edit(confirmSelfDeletion, markupConfirmSelfDeletion)
	}
	return ctx.Edit(youHaveActiveSchedule+preDeletionBarberInstruction, markupBackToMainBarber)
}

func onSureToDelete(ctx tele.Context) error {
	errMsg := "can't self delete barber"
	barberIDToDelete := ctx.Sender().ID
	if err := cp.RepoWithContext.DeleteAppointmentsBeforeDate(barberIDToDelete, tm.Today()); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if err := cp.RepoWithContext.DeleteBarberByID(barberIDToDelete); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	cfg.Barbers.RemoveID(barberIDToDelete)
	return ctx.Edit(goodbuyBarber)
}

func onManageBarbers(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, "can't open the set barbers menu", err)
	}
	return ctx.Edit(manageBarbers, markupManageBarbers)
}

func onShowAllBarbers(ctx tele.Context) error {
	errMsg := "can't show all barbers"
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberIDs := cfg.Barbers.IDs()
	barbersInfo := ""
	for _, barberID := range barberIDs {
		barber, err := cp.RepoWithContext.GetBarberByID(barberID)
		if err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		barbersInfo = barbersInfo + "\n\n" + barber.Info()
	}
	return ctx.Edit(listOfBarbers+barbersInfo, markupBackToMainBarber)
}

func onAddBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.NewStatus(ent.StateAddBarber)}); err != nil {
		return logAndMsgErrBarber(ctx, "can't ask barber to point on user to add", err)
	}
	return ctx.Edit(addBarber)
}

func onDeleteBarber(ctx tele.Context) error {
	errMsg := "can't suggest actions to delete barber"
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberIDs := cfg.Barbers.IDs()
	if len(barberIDs) == 1 {
		return ctx.Edit(onlyOneBarberExists, markupBackToMainBarber)
	}
	markupSelectBarber, notEmpty, err := markupSelectBarberToDeletion(ctx, barberIDs)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	if !notEmpty {
		return ctx.Edit(noBarbersToDelete+preDeletionBarberInstruction, markupSelectBarber)
	}
	return ctx.Edit(selectBarberToDeletion+preDeletionBarberInstruction, markupSelectBarber)
}

func onDeleteCertainBarber(ctx tele.Context) error {
	errMsg := "can't delete barber"
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberIDToDelete, err := strconv.ParseInt(ctx.Callback().Data, 10, 64)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	barberToDelete, err := cp.RepoWithContext.GetBarberByID(barberIDToDelete)
	if err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	today := tm.Today()
	if barberToDelete.LastWorkdate.Before(today) {
		if err := cp.RepoWithContext.DeleteAppointmentsBeforeDate(barberIDToDelete, today); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		if err := cp.RepoWithContext.DeleteBarberByID(barberIDToDelete); err != nil {
			return logAndMsgErrBarber(ctx, errMsg, err)
		}
		cfg.Barbers.RemoveID(barberIDToDelete)
		return ctx.Edit(barberDeleted, markupBackToMainBarber)
	}
	return ctx.Edit(barberHaveActiveSchedule, markupBackToMainBarber)
}

func onBackToMainBarber(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, "can't go back to the barber's main menu", err)
	}
	return ctx.Edit(mainBarber, markupMainBarber)
}

func onTextBarber(ctx tele.Context) error {
	status, err := actualizeBarberStatus(ctx.Sender().ID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't handle barber's text message", err)
	}
	switch status.State {
	case ent.StateUpdName:
		return updateBarberName(ctx)
	case ent.StateUpdPhone:
		return updateBarberPhone(ctx)
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func onContactBarber(ctx tele.Context) error {
	status, err := actualizeBarberStatus(ctx.Sender().ID)
	if err != nil {
		return logAndMsgErrBarber(ctx, "can't handle barber's contact message", err)
	}
	switch status.State {
	case ent.StateAddBarber:
		return addNewBarber(ctx)
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func actualizeBarberStatus(barberID int64) (status ent.Status, err error) {
	defer func() { err = e.WrapIfErr("can't actualize barber status", err) }()
	barber, err := cp.RepoWithContext.GetBarberByID(barberID)
	if err != nil {
		return ent.StatusStart, err
	}
	if !barber.Status.IsValid() {
		if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: barberID, Status: ent.StatusStart}); err != nil {
			return ent.StatusStart, err
		}
		return ent.StatusStart, nil
	}
	return barber.Status, nil
}

func addNewBarber(ctx tele.Context) error {
	errMsg := "can't add new barber"
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	newBarberID := ctx.Message().Contact.UserID
	if err := cp.RepoWithContext.CreateBarber(newBarberID); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			return ctx.Send(userIsAlreadyBarber, markupBackToMainBarber)
		}
		return logAndMsgErrBarber(ctx, errMsg, err)
	}
	cfg.Barbers.AddID(newBarberID)
	if err := sched.MakeSchedule(newBarberID); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(addedNewBarberWithoutShedule, markupBackToMainBarber)
	}
	return ctx.Send(addedNewBarberWithShedule, markupBackToMainBarber)
}

func logAndMsgErrBarber(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorBarber, markupBackToMainBarber)
}

func markupSelectBarberToDeletion(ctx tele.Context, barberIDs []int64) (*tele.ReplyMarkup, bool, error) {
	today := tm.Today()
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, barberID := range barberIDs {
		if barberID != ctx.Sender().ID {
			barber, err := cp.RepoWithContext.GetBarberByID(barberID)
			if err != nil {
				return &tele.ReplyMarkup{}, false, e.Wrap("can't make reply markup", err)
			}
			if barber.LastWorkdate.Before(today) {
				barberIDString := strconv.FormatInt(barberID, 10)
				row := markup.Row(markup.Data(barber.Name, endpntBarberToDeletion, barberIDString))
				rows = append(rows, row)
			}
		}
	}
	notEmpty := len(rows) > 0
	rows = append(rows, markup.Row(btnBackToMainBarber))
	markup.Inline(rows...)
	return markup, notEmpty, nil
}

func markupSelectLastWorkDate(firstDisplayedDateRange ent.DateRange, relativeFirstDisplayedMonth, deltaDisplayedMonth int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var btnPrevMonth, btnNextMonth tele.Btn
	var displayedDateRange ent.DateRange
	if deltaDisplayedMonth == 0 {
		displayedDateRange = firstDisplayedDateRange
		btnPrevMonth = btnEmpty
		btnNextMonth = markup.Data(next, endpntSelectMonthOfLastWorkDate, strconv.Itoa(1))
	} else {
		displayedDateRange = ent.Month(byte(relativeFirstDisplayedMonth + deltaDisplayedMonth))
		btnPrevMonth = markup.Data(prev, endpntSelectMonthOfLastWorkDate, strconv.Itoa(deltaDisplayedMonth-1))
		relativeMaxDisplayedMonth := int(cfg.ScheduledWeeks) * 7 / 30
		if relativeFirstDisplayedMonth+deltaDisplayedMonth == relativeMaxDisplayedMonth {
			btnNextMonth = btnEmpty
		} else {
			btnNextMonth = markup.Data(next, endpntSelectMonthOfLastWorkDate, strconv.Itoa(deltaDisplayedMonth+1))
		}
	}
	rowSelectMonth := markup.Row(btnPrevMonth, btnMonth(displayedDateRange.Month()), btnNextMonth)
	var btnsDatesToSelect []tele.Btn
	for i := 1; i < displayedDateRange.StartWeekday(); i++ {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	for date := displayedDateRange.StartDate; date.Compare(displayedDateRange.EndDate) <= 0; date = date.Add(24 * time.Hour) {
		btnDateToSelect := markup.Data(strconv.Itoa(date.Day()), endpntSelectLastWorkDate, date.Format(time.DateOnly))
		btnsDatesToSelect = append(btnsDatesToSelect, btnDateToSelect)
	}
	for i := 7; i > displayedDateRange.EndWeekday(); i-- {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	rowsSelectDate := markup.Split(7, btnsDatesToSelect)
	rowRestoreDefaultDate := markup.Row(markup.Data("Установить бессрочную дату окончания работы", endpntSelectLastWorkDate, cfg.InfiniteWorkDate))
	rowBackToMainBarber := markup.Row(btnBackToMainBarber)
	var rows []tele.Row
	rows = append(rows, rowSelectMonth, rowWeekdays)
	rows = append(rows, rowsSelectDate...)
	rows = append(rows, rowRestoreDefaultDate, rowBackToMainBarber)
	markup.Inline(rows...)
	return markup
}

func updateBarberName(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Name: ctx.Message().Text, Status: ent.StatusStart}); err != nil {
		if errors.Is(err, rep.ErrInvalidName) {
			log.Print(e.Wrap("invalid barber name", err))
			return ctx.Send(invalidBarberName)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's name must be unique", err))
			return ctx.Send(notUniqueBarberName)
		}
		return logAndMsgErrBarber(ctx, "can't update barber's name", err)
	}
	return ctx.Send(updNameSuccessBarber, markupPersonalBarber)
}

func updateBarberPhone(ctx tele.Context) error {
	if err := cp.RepoWithContext.UpdateBarber(ent.Barber{ID: ctx.Sender().ID, Phone: ctx.Message().Text, Status: ent.StatusStart}); err != nil {
		if errors.Is(err, rep.ErrInvalidPhone) {
			log.Print(e.Wrap("invalid barber phone", err))
			return ctx.Send(invalidBarberPhone)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's phone must be unique", err))
			return ctx.Send(notUniqueBarberPhone)
		}
		return logAndMsgErrBarber(ctx, "can't update barber's phone", err)
	}
	return ctx.Send(updPhoneSuccessBarber, markupPersonalBarber)
}
