package telegram

import (
	cfg "barbershop-bot/lib/config"

	tele "gopkg.in/telebot.v3"
	mw "gopkg.in/telebot.v3/middleware"
)

// Whitelist returns a middleware that skips the update for users
// NOT specified in the chats field.
func whitelist() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return mw.Restrict(mw.RestrictConfig{
			Chats: cfg.Barbers.IDs(),
			In:    next,
			Out:   noAction,
		})(next)
	}
}

// NotInWhitelist returns a middleware that skips the update for users
// specified in the chats field.
func notInWhitelist() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return mw.Restrict(mw.RestrictConfig{
			Chats: cfg.Barbers.IDs(),
			Out:   next,
			In:    noAction,
		})(next)
	}
}

func onStartRestrict() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return mw.Restrict(mw.RestrictConfig{
			Chats: cfg.Barbers.IDs(),
			In:    onStartBarber,
			Out:   onStartUser,
		})(next)
	}
}

func onTextRestrict() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return mw.Restrict(mw.RestrictConfig{
			Chats: cfg.Barbers.IDs(),
			In:    onTextBarber,
			Out:   onStartUser, // TODO
		})(next)
	}
}
