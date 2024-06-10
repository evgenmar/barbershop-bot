package telegram

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	rep "barbershop-bot/repository"
	sched "barbershop-bot/scheduler"
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
		81234567890
		8(123)-456-7890
		+71234567890
		+7 123 456 7890
Номер должен начинаться с "+7" или с "8". Региональный код можно по желанию заключить в скобки, также допустимо разделять пробелами или тире группы цифр.
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
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		log.Print(e.Wrap("can't open the barber's main menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Send(mainBarber, markupMainBarber)
}

func onSettingsBarber(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		log.Print(e.Wrap("can't open the barber settings menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(settingsBarber, markupSettingsBarber)
}

func onUpdPersonalBarber(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		log.Print(e.Wrap("can't open the barber's personal data menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(personalBarber, markupPersonalBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.NewStatus(ent.StateUpdName)}); err != nil {
		log.Print(e.Wrap("can't ask barber to enter name", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(updNameBarber)
}

func onUpdPhoneBarber(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.NewStatus(ent.StateUpdPhone)}); err != nil {
		log.Print(e.Wrap("can't ask barber to enter phone", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(updPhoneBarber)
}

func onManageBarbers(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		log.Print(e.Wrap("can't open the set barbers menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(manageBarbers, markupManageBarbers)
}

func onAddBarber(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.NewStatus(ent.StateAddBarber)}); err != nil {
		log.Print(e.Wrap("can't ask barber to point on user to add", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(addBarber)
}

func onDeleteBarber(ctx tele.Context) error {
	errMsg := "can't suggest actions to delete barber"
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	barberIDs, err := getAllBarberIDs()
	if err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	if len(barberIDs) == 1 {
		return ctx.Edit(noBarbersToDelete, markupBackToMainBarber)
	}
	markupSelectBarber, err := markupSelectBarberToDeletion(ctx, barberIDs)
	if err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(selectBarberToDeletion, markupSelectBarber)
}

func onBackToMainBarber(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		log.Print(e.Wrap("can't go back to the barber's main menu", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Edit(mainBarber, markupMainBarber)
}

func onTextBarber(ctx tele.Context) error {
	status, err := actualizeBarberStatus(ctx.Sender().ID)
	if err != nil {
		log.Print(e.Wrap("can't handle barber's text message", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
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
	errMsg := "can't handle barber's contact message"
	status, err := actualizeBarberStatus(ctx.Sender().ID)
	if err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	switch status.State {
	case ent.StateAddBarber:
		return addNewBarber(ctx, errMsg)
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func actualizeBarberStatus(barberID int64) (status ent.Status, err error) {
	defer func() { err = e.WrapIfErr("can't actualize barber status", err) }()
	status, err = getBarberStatusByID(barberID)
	if err != nil {
		return ent.StatusStart, err
	}
	if !status.IsValid() {
		if err := updBarber(ent.Barber{ID: barberID, Status: ent.StatusStart}); err != nil {
			return ent.StatusStart, err
		}
		return ent.StatusStart, nil
	}
	return status, nil
}

func addNewBarber(ctx tele.Context, errMsg string) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Status: ent.StatusStart}); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	if err := saveNewBarberID(ctx.Message().Contact.UserID); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			return ctx.Send(userIsAlreadyBarber, markupBackToMainBarber)
		}
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	cfg.Barbers.SetIDs(append(cfg.Barbers.IDs(), ctx.Message().Contact.UserID))
	if err := makeBarberSchedule(ctx.Message().Contact.UserID); err != nil {
		log.Print(e.Wrap(errMsg, err))
		return ctx.Send(addedNewBarberWithoutShedule, markupBackToMainBarber)
	}
	return ctx.Send(addedNewBarberWithShedule, markupBackToMainBarber)
}

func getAllBarberIDs() (barberIDs []int64, err error) {
	defer func() { err = e.WrapIfErr("can't get barber IDs", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutRead)
	defer cancel()
	return rep.Rep.FindAllBarberIDs(ctx)
}

func getBarberNameByID(barberID int64) (name string, err error) {
	defer func() { err = e.WrapIfErr("can't get barber name", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutRead)
	barber, err := rep.Rep.GetBarberByID(ctx, barberID)
	cancel()
	return barber.Name, err
}

func getBarberStatusByID(barberID int64) (status ent.Status, err error) {
	defer func() { err = e.WrapIfErr("can't get barber status", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutRead)
	barber, err := rep.Rep.GetBarberByID(ctx, barberID)
	cancel()
	return barber.Status, err
}

func makeBarberSchedule(barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't make barber schedule", err) }()
	return sched.MakeSchedule(barberID)
}

func markupSelectBarberToDeletion(ctx tele.Context, barberIDs []int64) (*tele.ReplyMarkup, error) {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, barberID := range barberIDs {
		if barberID != ctx.Sender().ID {
			barberName, err := getBarberNameByID(barberID)
			if err != nil {
				return &tele.ReplyMarkup{}, e.Wrap("can't make reply markup", err)
			}
			barberIDStr := strconv.FormatInt(barberID, 10)
			row := markup.Row(markup.Data(barberName, endpntBarberToDeletion, barberIDStr))
			rows = append(rows, row)
		}
	}
	rows = append(rows, markup.Row(btnBackToMainBarber))
	markup.Inline(rows...)
	return markup, nil
}

func saveNewBarberID(barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't save new barber ID", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutWrite)
	defer cancel()
	return rep.Rep.CreateBarber(ctx, barberID)
}

func updBarber(barber ent.Barber) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber status", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimoutWrite)
	defer cancel()
	return rep.Rep.UpdateBarber(ctx, barber)
}

func updateBarberName(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Name: ctx.Message().Text, Status: ent.StatusStart}); err != nil {
		if errors.Is(err, rep.ErrInvalidName) {
			log.Print(e.Wrap("invalid barber name", err))
			return ctx.Send(invalidBarberName)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's name must be unique", err))
			return ctx.Send(notUniqueBarberName)
		}
		log.Print(e.Wrap("can't update barber's name", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Send(updNameSuccessBarber, markupPersonalBarber)
}

func updateBarberPhone(ctx tele.Context) error {
	if err := updBarber(ent.Barber{ID: ctx.Sender().ID, Phone: ctx.Message().Text, Status: ent.StatusStart}); err != nil {
		if errors.Is(err, rep.ErrInvalidPhone) {
			log.Print(e.Wrap("invalid barber phone", err))
			return ctx.Send(invalidBarberPhone)
		}
		if errors.Is(err, rep.ErrNonUniqueData) {
			log.Print(e.Wrap("barber's phone must be unique", err))
			return ctx.Send(notUniqueBarberPhone)
		}
		log.Print(e.Wrap("can't update barber's phone", err))
		return ctx.Send(errorBarber, markupBackToMainBarber)
	}
	return ctx.Send(updPhoneSuccessBarber, markupPersonalBarber)
}
