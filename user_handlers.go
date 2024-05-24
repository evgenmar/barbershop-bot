package main

import tele "gopkg.in/telebot.v3"

//example
func onUser(c tele.Context) error {
	return c.Send("hello user")
}

//example
func onStartUser(c tele.Context) error {
	return c.Send("wellcome user")
}
