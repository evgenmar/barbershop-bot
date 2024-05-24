package main

import (
	"context"
	"log"

	tele "gopkg.in/telebot.v3"
)

var (
	menuMainBarber = &tele.ReplyMarkup{}
)

func init() {
	menuMainBarber.Inline(menuMainBarber.Row(btnUpdPersonal))
}

func onStartBarber(ctx tele.Context) error {
	return ctx.Send(mainMenuText, menuMainBarber)
}

func onBackToMainBarber(ctx tele.Context) error {
	return ctx.Edit(mainMenuText, menuMainBarber)
}

func onUpdNameBarber(ctx tele.Context) error {
	if err := store(ctx).SaveBarberState(context.TODO(), stateUpdName, ctx.Sender().ID); err != nil {
		log.Print(err)
	}
	return ctx.Send("Как Вас зовут?")
}
