package telegram

import (
	"barbershop-bot/storage"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

// NotInWhitelist returns a middleware that skips the update for users
// specified in the chats field.
func notInWhitelist(chats ...int64) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return middleware.Restrict(middleware.RestrictConfig{
			Chats: chats,
			Out:   next,
			In:    noAction,
		})(next)
	}
}

func withStorage(storage storage.Storage) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(ctx tele.Context) error {
			ctx.Set("storage", storage)
			return next(ctx)
		}
	}
}
