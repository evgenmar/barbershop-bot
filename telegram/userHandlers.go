package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	"barbershop-bot/lib/e"
	rep "barbershop-bot/repository"
	sess "barbershop-bot/sessions"
	"errors"
	"log"

	tele "gopkg.in/telebot.v3"
)

func onStartUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Send(mainMenu, markupMainUser)
}

func onSettingsUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(settingsMenu, markupSettingsUser)
}

func onUpdPersonalUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	userID := ctx.Sender().ID
	user, err := cp.RepoWithContext.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, rep.ErrNoSavedUser) {
			return ctx.Edit(privacyExplanation, markupPrivacyExplanation)
		}
		return logAndMsgErrUser(ctx, "can't show update personal data options for user", err)
	}
	if user.Name == ent.NoName && user.Phone == ent.NoPhone {
		return ctx.Edit(privacyExplanation, markupPrivacyExplanation)
	}
	return ctx.Edit(selectPersonalData, markupPersonalUser)
}

func onPrivacyUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(privacyUser, markupPrivacyUser)
}

func onUserAgreeWithPrivacy(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(selectPersonalData, markupPersonalUser)
}

func onUpdNameUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateUpdName)
	return ctx.Edit(updName)
}

func onUpdPhoneUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateUpdPhone)
	return ctx.Edit(updPhone)
}

func onBackToMainUser(ctx tele.Context) error {
	sess.UpdateUserState(ctx.Sender().ID, sess.StateStart)
	return ctx.Edit(mainMenu, markupMainUser)
}

func onTextUser(ctx tele.Context) error {
	state := sess.GetUserState(ctx.Sender().ID)
	switch state {
	case sess.StateUpdName:
		return onUpdateUserName(ctx)
	case sess.StateUpdPhone:
		return onUpdateUserPhone(ctx)
	default:
		return ctx.Send(unknownCommand)
	}
}

func onUpdateUserName(ctx tele.Context) error {
	errMsg := "can't update user's name"
	userID := ctx.Sender().ID
	text := ctx.Message().Text
	if err := cp.RepoWithContext.CreateUser(ent.User{ID: userID, Name: text}); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			if err := cp.RepoWithContext.UpdateUser(ent.User{ID: userID, Name: text}); err != nil {
				if errors.Is(err, rep.ErrInvalidUser) {
					return ctx.Send(invalidName)
				}
				sess.UpdateUserState(userID, sess.StateStart)
				return logAndMsgErrUser(ctx, errMsg, err)
			}
			sess.UpdateUserState(userID, sess.StateStart)
			return ctx.Send(updNameSuccess, markupPersonalUser)
		}
		if errors.Is(err, rep.ErrInvalidUser) {
			return ctx.Send(invalidName)
		}
		sess.UpdateUserState(userID, sess.StateStart)
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	sess.UpdateUserState(userID, sess.StateStart)
	return ctx.Send(updNameSuccess, markupPersonalUser)
}

func onUpdateUserPhone(ctx tele.Context) error {
	errMsg := "can't update user's phone"
	userID := ctx.Sender().ID
	text := ctx.Message().Text
	if err := cp.RepoWithContext.CreateUser(ent.User{ID: userID, Phone: text}); err != nil {
		if errors.Is(err, rep.ErrAlreadyExists) {
			if err := cp.RepoWithContext.UpdateUser(ent.User{ID: userID, Phone: text}); err != nil {
				if errors.Is(err, rep.ErrInvalidUser) {
					return ctx.Send(invalidPhone)
				}
				sess.UpdateUserState(userID, sess.StateStart)
				return logAndMsgErrUser(ctx, errMsg, err)
			}
			sess.UpdateUserState(userID, sess.StateStart)
			return ctx.Send(updPhoneSuccess, markupPersonalUser)
		}
		if errors.Is(err, rep.ErrInvalidUser) {
			return ctx.Send(invalidPhone)
		}
		sess.UpdateUserState(userID, sess.StateStart)
		return logAndMsgErrUser(ctx, errMsg, err)
	}
	sess.UpdateUserState(userID, sess.StateStart)
	return ctx.Send(updPhoneSuccess, markupPersonalUser)
}

func logAndMsgErrUser(ctx tele.Context, msg string, err error) error {
	log.Print(e.Wrap(msg, err))
	return ctx.Send(errorUser, markupBackToMainUser)
}
