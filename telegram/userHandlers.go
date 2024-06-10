package telegram

import tele "gopkg.in/telebot.v3"

//example
func onUser(c tele.Context) error {
	return c.Send("hello user")
}

func onStartUser(c tele.Context) error { //TODO
	return c.Send("wellcome user")
}
