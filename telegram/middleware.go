package telegram

import (
	"barbershop-bot/repository/storage"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func withStorage(storage storage.Storage) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(ctx tele.Context) error {
			ctx.Set("storage", storage)
			return next(ctx)
		}
	}
}

// Whitelist returns a middleware that skips the update for users
// NOT specified in the chats field.
func whitelist() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		chats := barberIDs.iDs()
		return middleware.Restrict(middleware.RestrictConfig{
			Chats: chats,
			In:    next,
			Out:   func(c tele.Context) error { return nil },
		})(next)
	}
}

// NotInWhitelist returns a middleware that skips the update for users
// specified in the chats field.
func notInWhitelist() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		chats := barberIDs.iDs()
		return middleware.Restrict(middleware.RestrictConfig{
			Chats: chats,
			Out:   next,
			In:    noAction,
		})(next)
	}
}

func onStartRestrict() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		chats := barberIDs.iDs()
		return middleware.Restrict(middleware.RestrictConfig{
			Chats: chats,
			In:    onStartBarber,
			Out:   onStartUser,
		})(next)
	}
}

func onTextRestrict() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		chats := barberIDs.iDs()
		return middleware.Restrict(middleware.RestrictConfig{
			Chats: chats,
			In:    onTextBarber,
			Out:   onStartUser, // TODO
		})(next)
	}
}
