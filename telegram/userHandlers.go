package telegram

import tele "gopkg.in/telebot.v3"

var (
	markupMainUser  = &tele.ReplyMarkup{}
	btnSettingsUser = markupMainBarber.Data("Настройки пользователь", "settings_user")

	markupBackToMainUser = &tele.ReplyMarkup{}
	btnBackToMainUser    = markupBackToMainUser.Data("Вернуться в главное меню", "back_to_main_user")
)

func init() {
	markupMainUser.Inline(
		markupMainUser.Row(btnSettingsUser),
	)

	markupBackToMainUser.Inline(
		markupBackToMainUser.Row(btnBackToMainUser),
	)
}

func onStartUser(c tele.Context) error { //TODO
	return c.Send("wellcome user")
}
