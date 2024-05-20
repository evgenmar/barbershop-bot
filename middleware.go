package main

import (
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

// NotInWhitelist returns a middleware that skips the update for users
// specified in the chats field.
func NotInWhitelist(chats ...int64) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return middleware.Restrict(middleware.RestrictConfig{
			Chats: chats,
			Out:   next,
			In:    func(c tele.Context) error { return nil },
		})(next)
	}
}
