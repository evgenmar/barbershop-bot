package main

import (
	"barbershop-bot/storage"
	"log"

	tele "gopkg.in/telebot.v3"
)

const (
	stateStart uint8 = iota
	stateUpdName
)
const mainMenuText = "Добрый день. Вы находитесь в главном меню. Выберите действие."

var (
	menuPersonal = &tele.ReplyMarkup{}
	btnUpdName   = menuPersonal.Data("Обновить имя", "upd_name")
	btnUpdPhone  = menuPersonal.Data("Обновить номер телефона", "upd_phone")

	btnUpdPersonal = menuPersonal.Data("Обновить личные данные", "upd_personal_data")

	btnBackToMain = menuPersonal.Data("Вернуться в главное меню", "back_to_main")
)

func init() {
	menuPersonal.Inline(
		menuPersonal.Row(btnUpdName),
		menuPersonal.Row(btnUpdPhone),
		menuPersonal.Row(btnBackToMain),
	)
}

func noAction(tele.Context) error { return nil }

func store(ctx tele.Context) storage.Storage {
	rep, ok := ctx.Get("storage").(storage.Storage)
	if !ok {
		log.Print("can't get storage from Context")
		return nil
	}
	return rep
}

func onUpdPersonal(ctx tele.Context) error {
	return ctx.Edit("Выберите данные, которые вы хотите обновить.", menuPersonal)
}
