package telegram

import (
	"barbershop-bot/config"
	"barbershop-bot/storage"
	"context"
	"log"

	tele "gopkg.in/telebot.v3"
)

const (
	textMainBarber       = "Добрый день. Вы находитесь в главном меню. Выберите действие."
	textPersonalBarber   = "Выберите данные, которые вы хотите обновить."
	textUpdNameBarber    = "Как Вас зовут?"
	textUpdPhoneBarber   = "" //TODO
	textUnknownCmdBarber = "Неизвестная команда. Пожалуйста, выберите команду из меню. Для вызова главного меню воспользуйтесь командой /start"
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
	updateBarberState(ctx, stateStart, "can't open the barber's main menu")
	return ctx.Send(textMainBarber, markupMainBarber)
}

func onUpdPersonalBarber(ctx tele.Context) error {
	updateBarberState(ctx, stateStart, "can't open the barber's personal data menu")
	return ctx.Edit(textPersonalBarber, markupPersonalBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	updateBarberState(ctx, stateUpdName, "can't ask barber to enter name")
	return ctx.Send(textUpdNameBarber)
}

func onUpdPhoneBarber(ctx tele.Context) error {
	updateBarberState(ctx, stateUpdPhone, "can't ask barber to enter phone")
	return ctx.Send(textUpdPhoneBarber)
}

func onBackToMainBarber(ctx tele.Context) error {
	updateBarberState(ctx, stateStart, "can't go back to the barber's main menu")
	return ctx.Edit(textMainBarber, markupMainBarber)
}

func onTextBarber(ctx tele.Context) error {
	errMsg := "can't handle barber's text message"
	rep := getRepository(ctx, errMsg)
	state, expired := getBarberState(ctx, rep, errMsg)
	if expired {
		updBarberState(ctx, rep, stateStart, errMsg)
		return ctx.Send(textUnknownCmdBarber)
	}

	switch state {
	case stateUpdName:
		errMsg = "can't update barber's name"
		ok, err := isValidName(ctx.Message().Text)
		if err != nil {
			log.Panic(errMsg, err)
		}
		if ok {
			updBarberNameAndState(ctx, rep, stateStart, errMsg) //TODO
		}
		return ctx.Send(ctx.Message().Text) //TODO
	case stateUpdPhone:
		return ctx.Send("Обновляем телефон") //TODO
	default:
		return ctx.Send(textUnknownCmdBarber)
	}
}

// getBarberState returns barbers state. If the state has not expired yet, the second returned value is false.
// If the state has already expired, the second returned value is true.
func getBarberState(ctxTl tele.Context, rep storage.Storage, errMsg string) (state, bool) {
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutRead)
	status, err := rep.GetBarberStatus(ctxDb, ctxTl.Sender().ID)
	cancel()
	if err != nil {
		log.Panicf("%s: %s", errMsg, err)
	}
	state, expired, err := getState(status)
	if err != nil {
		log.Panicf("%s: %s", errMsg, err)
	}
	return state, expired
}

func isValidName(text string) (bool, error) {
	_ = text
	return true, nil //TODO
}

func updBarberNameAndState(ctxTl tele.Context, rep storage.Storage, state state, errMsg string) {
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	if err := rep.UpdateBarberNameAndStatus(ctxDb, ctxTl.Message().Text, newStatus(state), ctxTl.Sender().ID); err != nil {
		log.Panicf("%s: %s", errMsg, err)
	}
	cancel()
}

func updateBarberState(ctxTl tele.Context, state state, errMsg string) {
	rep := getRepository(ctxTl, errMsg)
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	if err := rep.UpdateBarberStatus(ctxDb, newStatus(state), ctxTl.Sender().ID); err != nil {
		log.Panicf("%s: %s", errMsg, err)
	}
	cancel()
}

func updBarberState(ctxTl tele.Context, rep storage.Storage, state state, errMsg string) {
	ctxDb, cancel := context.WithTimeout(context.Background(), config.DbQueryTimoutWrite)
	if err := rep.UpdateBarberStatus(ctxDb, newStatus(state), ctxTl.Sender().ID); err != nil {
		log.Panicf("%s: %s", errMsg, err)
	}
	cancel()
}
