package telegram

import (
	"barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	"barbershop-bot/repository/storage"
	"barbershop-bot/scheduler"
	"context"
	"errors"
	"log"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

const (
	mainBarber = "Добрый день. Вы находитесь в главном меню. Выберите действие."

	settingsBarber = "Вы находитесь меню настроек. Выберите действие."

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
		1234567890
		(123)-456-78-90
		81234567890
		8(123)-456-7890
		+71234567890
		+7 123 456 7890
Пожалуйста, попробуйте ввести номер телефона еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberPhone = `Номер телефона не сохранен. Другой барбер с таким номером уже зарегистрирован в приложении. Введите другой номер.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	manageBarbers = "В этом меню Вы можете добавить нового барбера или удалить существующего. Выберите действие."
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
	noBarbersToDelete      = "Вы единственный зарегистрированный барбер в приложении. Некого удалять."
	endpntBarberToDeletion = "barber_to_deletion"
	selectBarberToDeletion = "Выберите барбера, которого Вы хотите удалить"

	unknownCmdBarber = "Неизвестная команда. Пожалуйста, выберите команду из меню. Для вызова главного меню воспользуйтесь командой /start"
	errorBarber      = `Произошла ошибка обработки команды. Команда не была выполнена. Если ошибка будет повторяться, возможно, потребуется перезапуск сервиса.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`
)

var (
	markupMainBarber  = &tele.ReplyMarkup{}
	btnSettingsBarber = markupMainBarber.Data("Настройки", "settings_barber")

	markupSettingsBarber = &tele.ReplyMarkup{}
	btnUpdPersonalBarber = markupSettingsBarber.Data("Обновить персональные данные", "upd_personal_data_barber")
	btnManageBarbers     = markupSettingsBarber.Data("Управление барберами", "manage_barbers")

	markupPersonalBarber = &tele.ReplyMarkup{}
	btnUpdNameBarber     = markupPersonalBarber.Data("Обновить имя", "upd_name_barber")
	btnUpdPhoneBarber    = markupPersonalBarber.Data("Обновить номер телефона", "upd_phone_barber")

	markupManageBarbers = &tele.ReplyMarkup{}
	btnAddBarber        = markupManageBarbers.Data("Добавить барбера", "add_barber")
	btnDeleteBarber     = markupManageBarbers.Data("Удалить барбера", "delete_barber")
	//btnDeleteExactBarber = markupManageBarbers.Data("", endpntBarberToDeletion)

	markupBackToMainBarber = &tele.ReplyMarkup{}
	btnBackToMainBarber    = markupBackToMainBarber.Data("Вернуться в главное меню", "back_to_main_barber")
)

func init() {
	markupMainBarber.Inline(
		markupMainBarber.Row(btnSettingsBarber),
	)

	markupSettingsBarber.Inline(
		markupSettingsBarber.Row(btnUpdPersonalBarber),
		markupSettingsBarber.Row(btnManageBarbers),
		markupSettingsBarber.Row(btnBackToMainBarber),
	)

	markupPersonalBarber.Inline(
		markupPersonalBarber.Row(btnUpdNameBarber),
		markupPersonalBarber.Row(btnUpdPhoneBarber),
		markupPersonalBarber.Row(btnBackToMainBarber),
	)

	markupManageBarbers.Inline(
		markupManageBarbers.Row(btnAddBarber),
		markupManageBarbers.Row(btnDeleteBarber),
		markupManageBarbers.Row(btnBackToMainBarber),
	)

	markupBackToMainBarber.Inline(
		markupBackToMainBarber.Row(btnBackToMainBarber),
	)
}

func onStartBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't open the barber's main menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Send(mainBarber, markupMainBarber)
}

func onSettingsBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't open the barber settings menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(settingsBarber, markupSettingsBarber)
}

func onUpdPersonalBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't open the barber's personal data menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(personalBarber, markupPersonalBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateUpdName); err != nil {
		log.Print(e.Wrap("can't ask barber to enter name", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(updNameBarber)
}

func onUpdPhoneBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateUpdPhone); err != nil {
		log.Print(e.Wrap("can't ask barber to enter phone", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(updPhoneBarber)
}

func onManageBarbers(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't open the set barbers menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(manageBarbers, markupManageBarbers)
}

func onAddBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateAddBarber); err != nil {
		log.Print(e.Wrap("can't ask barber to point on user to add", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(addBarber)
}

func onDeleteBarber(ctx tele.Context) error {
	errMsg := "can't suggest actions to delete barber"
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	barberIDs, err := getAllBarberIDs(ctx)
	if err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	if len(barberIDs) == 1 {
		return ctx.Edit(noBarbersToDelete, markupBackToMainBarber)
	}
	markupSelectBarber := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, barberID := range barberIDs {
		if barberID != ctx.Sender().ID {
			barberName, err := getBarberNameByID(ctx, barberID)
			if err != nil {
				log.Print(e.Wrap(errMsg, err))
				return ctx.Send(errorBarber, markupBackToMainBarber)
			}
			barberIDStr := strconv.FormatInt(barberID, 10)
			row := markupSelectBarber.Row(markupSelectBarber.Data(barberName, endpntBarberToDeletion, barberIDStr))
			rows = append(rows, row)
		}
	}
	markupSelectBarber.Inline(rows...)
	return ctx.Edit(selectBarberToDeletion, markupSelectBarber)
}

func onBackToMainBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't go back to the barber's main menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(mainBarber, markupMainBarber)
}

func onTextBarber(ctx tele.Context) error {
	state, err := actualizeBarberState(ctx)
	if err != nil {
		log.Print(e.Wrap("can't handle barber's text message", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}

	switch state {
	case stateUpdName:
		if ok := isValidName(ctx.Message().Text); ok {
			if err := updBarberName(ctx); err != nil {
				if errors.Is(err, storage.ErrNonUniqueData) {
					log.Print(e.Wrap("barber's name must be unique", err))
					return ctx.Send(notUniqueBarberName)
				}
				log.Print(e.Wrap("can't update barber's name", err))
				return ctx.Send(errorBarber, markupBackToMainBarber)
			}
			if err := updBarberState(ctx, stateStart); err != nil {
				log.Print(e.Wrap("can't update barber's state", err))
			}
			return ctx.Send(updNameSuccessBarber, markupPersonalBarber)
		}
		return ctx.Send(invalidBarberName)
	case stateUpdPhone:
		if ok := isValidPhone(ctx.Message().Text); ok {
			if err := updBarberPhone(ctx); err != nil {
				if errors.Is(err, storage.ErrNonUniqueData) {
					log.Print(e.Wrap("barber's phone must be unique", err))
					return ctx.Send(notUniqueBarberPhone)
				}
				log.Print(e.Wrap("can't update barber's phone", err))
				return ctx.Send(errorBarber, markupBackToMainBarber)
			}
			if err := updBarberState(ctx, stateStart); err != nil {
				log.Print(e.Wrap("can't update barber's state", err))
			}
			return ctx.Send(updPhoneSuccessBarber, markupPersonalBarber)
		}
		return ctx.Send(invalidBarberPhone)
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func onContactBarber(ctx tele.Context) error {
	errMsg := "can't handle barber's contact message"
	state, err := actualizeBarberState(ctx)
	if err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	switch state {
	case stateAddBarber:
		return addNewBarber(ctx, errMsg)
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func addNewBarber(ctx tele.Context, errMsg string) error {
	isBarberExists, err := isBarberExists(ctx)
	if err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	if isBarberExists {
		return ctx.Send(userIsAlreadyBarber, markupBackToMainBarber)
	}
	if err := saveNewBarberID(ctx); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	barberIDs.setIDs(append(barberIDs.iDs(), ctx.Message().Contact.UserID))
	if err := makeBarberSchedule(ctx, ctx.Message().Contact.UserID); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(addedNewBarberWithoutShedule, markupBackToMainBarber)
	}
	return ctx.Send(addedNewBarberWithShedule, markupBackToMainBarber)
}

func actualizeBarberState(ctx tele.Context) (state state, err error) {
	defer func() { err = e.WrapIfErr("can't actualize barber state", err) }()
	state, ok, err := getBarberState(ctx)
	if err != nil {
		return stateStart, err
	}
	if !ok {
		if err := updBarberState(ctx, stateStart); err != nil {
			return stateStart, err
		}
		return stateStart, nil
	}
	return state, nil
}

func getBarberNameByID(ctxTl tele.Context, barberID int64) (name string, err error) {
	defer func() { err = e.WrapIfErr("can't get barber name", err) }()
	noName := "Барбер без имени"
	rep, err := getRepository(ctxTl)
	if err != nil {
		return noName, err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	barber, err := rep.GetBarberByID(ctxDb, barberID)
	cancel()
	if err != nil {
		return noName, err
	}
	if barber.Name.Valid {
		return barber.Name.String, nil
	}
	return noName, nil
}

// getBarberState returns barbers state. If the state has not expired yet, the second returned value is true.
// If the state has already expired, the second returned value is false.
func getBarberState(ctxTl tele.Context) (state state, notExpired bool, err error) {
	defer func() { err = e.WrapIfErr("can't get barber state", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return stateStart, false, err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	barber, err := rep.GetBarberByID(ctxDb, ctxTl.Sender().ID)
	cancel()
	if err != nil {
		return stateStart, false, err
	}
	return getState(barber.Status)
}

func getAllBarberIDs(ctxTl tele.Context) (barberIDs []int64, err error) {
	defer func() { err = e.WrapIfErr("can't get barber IDs", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return nil, err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	defer cancel()
	return rep.FindAllBarberIDs(ctxDb)
}

func isBarberExists(ctxTl tele.Context) (exists bool, err error) {
	defer func() { err = e.WrapIfErr("can't check if barber exists", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return false, err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	defer cancel()
	return rep.IsBarberExists(ctxDb, ctxTl.Message().Contact.UserID)
}

func makeBarberSchedule(ctx tele.Context, barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't make barber schedule", err) }()
	rep, err := getRepository(ctx)
	if err != nil {
		return err
	}
	return scheduler.MakeSchedule(rep, barberID)
}

func saveNewBarberID(ctxTl tele.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't save new barber ID", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	defer cancel()
	return rep.CreateBarber(ctxDb, ctxTl.Message().Contact.UserID)
}

func updBarberName(ctxTl tele.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber name", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	defer cancel()
	if err := rep.UpdateBarberName(ctxDb, ctxTl.Message().Text, ctxTl.Sender().ID); err != nil {
		return err
	}
	return nil
}

func updBarberPhone(ctxTl tele.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber phone and state", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	defer cancel()
	phone := normalizePhone(ctxTl.Message().Text)
	if err := rep.UpdateBarberPhone(ctxDb, phone, ctxTl.Sender().ID); err != nil {
		return err
	}
	return nil
}

func updBarberState(ctxTl tele.Context, state state) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber state", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	defer cancel()
	if err := rep.UpdateBarberStatus(ctxDb, newStatus(state), ctxTl.Sender().ID); err != nil {
		return err
	}
	return nil
}
