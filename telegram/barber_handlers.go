package telegram

import (
	"context"
	"log"

	tele "gopkg.in/telebot.v3"
)

const (
	textMainBarber     = "Добрый день. Вы находитесь в главном меню. Выберите действие."
	textPersonalBarber = "Выберите данные, которые вы хотите обновить."
	textUpdNameBarber  = "Как Вас зовут?"
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
	return ctx.Send(textMainBarber, markupMainBarber)
}

func onUpdPersonalBarber(ctx tele.Context) error {
	return ctx.Edit(textPersonalBarber, markupPersonalBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	if err := store(ctx).UpdateBarberStatus(context.TODO(), newStatus(stateUpdName), ctx.Sender().ID); err != nil {
		log.Print(err)
	}
	return ctx.Send(textUpdNameBarber)
}

func onUpdPhoneBarber(ctx tele.Context) error {
	_ = ctx //TODO
	return nil
}

func onBackToMainBarber(ctx tele.Context) error {
	return ctx.Edit(textMainBarber, markupMainBarber)
}
