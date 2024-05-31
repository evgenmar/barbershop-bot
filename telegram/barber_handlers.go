package telegram

import (
	"barbershop-bot/config"
	"barbershop-bot/lib/e"
	"barbershop-bot/storage"
	"context"
	"errors"
	"fmt"
	"log"

	tele "gopkg.in/telebot.v3"
)

var ID int64

const (
	mainBarber     = "Добрый день. Вы находитесь в главном меню. Выберите действие."
	personalBarber = "Выберите данные, которые вы хотите обновить."

	updNameBarber        = "Как Вас зовут?"
	updNameSuccessBarber = "Имя успешно обновлено. Хотите обновить другие данные?"
	updNameFailBarber    = `Введенное имя не соответствует установленным критериям:
		- имя может содержать русские и английские буквы, цифры и пробелы;
		- длина имени должна быть не менее 2 символов и не более 20 символов.
Пожалуйста, попробуйте ввести имя еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberName = `Имя не сохранено. Другой барбер с таким именем уже зарегистрирован в приложении. Введите другое имя.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	updPhoneBarber        = "Введите свой номер телефона"
	updPhoneSuccessBarber = "Номер телефона успешно обновлен. Хотите обновить другие данные?"
	updPhoneFailBarber    = `Неизвестный формат номера телефона. Примеры поддерживаемых форматов:
		1234567890
		(123)-456-78-90
		81234567890
		8(123)-456-7890
		+71234567890
		+7 123 456 7890
Пожалуйста, попробуйте ввести номер телефона еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberPhone = `Номер телефона не сохранен. Другой барбер с таким номером уже зарегистрирован в приложении. Введите другой номер.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	unknownCmdBarber = "Неизвестная команда. Пожалуйста, выберите команду из меню. Для вызова главного меню воспользуйтесь командой /start"
	errorBarber      = "Произошла ошибка обработки команды. Если ошибка будет повторяться, возможно, потребуется перезапуск сервиса"
)

var (
	markupMainBarber     = &tele.ReplyMarkup{}
	btnUpdPersonalBarber = markupPersonalBarber.Data("Обновить личные данные", "upd_personal_data_barber")

	markupPersonalBarber = &tele.ReplyMarkup{}
	btnUpdNameBarber     = markupPersonalBarber.Data("Обновить имя", "upd_name_barber")
	btnUpdPhoneBarber    = markupPersonalBarber.Data("Обновить номер телефона", "upd_phone_barber")
	btnBackToMainBarber  = markupPersonalBarber.Data("Вернуться в главное меню", "back_to_main_barber")
)

func init() {
	markupMainBarber.Inline(
		markupMainBarber.Row(btnUpdPersonalBarber),
	)
	markupPersonalBarber.Inline(
		markupPersonalBarber.Row(btnUpdNameBarber),
		markupPersonalBarber.Row(btnUpdPhoneBarber),
		markupPersonalBarber.Row(btnBackToMainBarber),
	)
}

func onStartBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't open the barber's main menu", err))
		return ctx.Send(errorBarber)
	}
	return ctx.Send(mainBarber, markupMainBarber)
}

func onUpdPersonalBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't open the barber's personal data menu", err))
		return ctx.Send(errorBarber)
	}
	return ctx.Edit(personalBarber, markupPersonalBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateUpdName); err != nil {
		log.Print(e.Wrap("can't ask barber to enter name", err))
		return ctx.Send(errorBarber)
	}
	return ctx.Send(updNameBarber)
}

func onUpdPhoneBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateUpdPhone); err != nil {
		log.Print(e.Wrap("can't ask barber to enter phone", err))
		return ctx.Send(errorBarber)
	}
	return ctx.Send(updPhoneBarber)
}

func onBackToMainBarber(ctx tele.Context) error {
	if err := updBarberState(ctx, stateStart); err != nil {
		log.Print(e.Wrap("can't go back to the barber's main menu", err))
		return ctx.Send(errorBarber)
	}
	return ctx.Edit(mainBarber, markupMainBarber)
}

func onTextBarber(ctx tele.Context) error {
	state, err := actualizeBarberState(ctx)
	if err != nil {
		log.Print(e.Wrap("can't handle barber's text message", err))
		return ctx.Send(errorBarber)
	}

	switch state {
	case stateStart:
		return ctx.Send(unknownCmdBarber)
	case stateUpdName:
		if ok := isValidName(ctx.Message().Text); ok {
			if err := updBarberNameAndState(ctx, stateStart); err != nil {
				if errors.Is(err, storage.ErrNonUniqueData) {
					log.Print(e.Wrap("barber's name must be unique", err))
					return ctx.Send(notUniqueBarberName)
				}
				log.Print(e.Wrap("can't update barber's name", err))
				return ctx.Send(errorBarber)
			}
			return ctx.Send(updNameSuccessBarber, markupPersonalBarber)
		}
		return ctx.Send(updNameFailBarber)
	case stateUpdPhone:
		if ok := isValidPhone(ctx.Message().Text); ok {
			if err := updBarberPhoneAndState(ctx, stateStart); err != nil {
				if errors.Is(err, storage.ErrNonUniqueData) {
					log.Print(e.Wrap("barber's phone must be unique", err))
					return ctx.Send(notUniqueBarberPhone)
				}
				log.Print(e.Wrap("can't update barber's phone", err))
				return ctx.Send(errorBarber)
			}
			return ctx.Send(updPhoneSuccessBarber, markupPersonalBarber)
		}
		return ctx.Send(updPhoneFailBarber)
	default:
		return ctx.Send(unknownCmdBarber)
	}
}

func onContactBarber(ctx tele.Context) error {
	fmt.Println(*ctx.Message().Contact) //TODO
	return nil
}

// getBarberState returns barbers state. If the state has not expired yet, the second returned value is true.
// If the state has already expired, the second returned value is false.
func getBarberState(ctxTl tele.Context, rep storage.Storage) (state, bool, error) {
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	status, err := rep.GetBarberStatus(ctxDb, ctxTl.Sender().ID)
	cancel()
	if err != nil {
		return stateStart, false, err
	}
	return getState(status)
}

func updBarberNameAndState(ctxTl tele.Context, state state) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber name and state", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	defer cancel()
	if err := rep.UpdateBarberNameAndStatus(ctxDb, ctxTl.Message().Text, newStatus(state), ctxTl.Sender().ID); err != nil {
		return err
	}
	return nil
}

func updBarberPhoneAndState(ctxTl tele.Context, state state) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber phone and state", err) }()
	rep, err := getRepository(ctxTl)
	if err != nil {
		return err
	}
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	defer cancel()
	phone := normalizePhone(ctxTl.Message().Text)
	if err := rep.UpdateBarberPhoneAndStatus(ctxDb, phone, newStatus(state), ctxTl.Sender().ID); err != nil {
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

func actualizeBarberState(ctx tele.Context) (state state, err error) {
	defer func() { err = e.WrapIfErr("can't actualize barber state", err) }()
	rep, err := getRepository(ctx)
	if err != nil {
		return stateStart, err
	}
	state, ok, err := getBarberState(ctx, rep)
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
