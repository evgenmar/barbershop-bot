package telegram

import tele "gopkg.in/telebot.v3"

const (
	privacyExplanation = `Мы просим Вас оставить свое имя и номер телефона, если вы хотите, чтобы в случае возникновения непредвиденных обстоятельств (заболел барбер, отключили свет/воду и т.п.) Вам позвонили на телефон и предупредили.
Если Вам будет достаточно, чтобы с Вами связались через telegram, оставлять персональные данные не нужно.
Перед тем как оставить свои персональные данные, ознакомьтесь с политикой конфиденциальности.`
	privacyUser = "Текст политики конфиденциальности для клиентов."

	errorUser = `Произошла ошибка обработки команды. Команда не была выполнена. Приносим извинения.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`
)

var (
	markupMainUser  = &tele.ReplyMarkup{}
	btnSettingsUser = markupEmpty.Data("Настройки", "settings_user")

	markupSettingsUser = &tele.ReplyMarkup{}
	btnUpdPersonalUser = markupEmpty.Data("Обновить персональные данные", "upd_personal_data_user")

	markupPrivacyExplanation = &tele.ReplyMarkup{}
	btnPrivacyUser           = markupEmpty.Data("Политика конфиденциальности", "privacy_policy_user")

	markupPrivacyUser       = &tele.ReplyMarkup{}
	btnUserAgreeWithPrivacy = markupEmpty.Data("Соглашаюсь с политикой конфиденциальности", "user_agree_with_privacy")

	markupPersonalUser = &tele.ReplyMarkup{}
	btnUpdNameUser     = markupEmpty.Data("Обновить имя", "upd_name_user")
	btnUpdPhoneUser    = markupEmpty.Data("Обновить номер телефона", "upd_phone_user")

	markupBackToMainUser = &tele.ReplyMarkup{}
	btnBackToMainUser    = markupEmpty.Data("Вернуться в главное меню", "back_to_main_user")
)

func init() {
	markupMainUser.Inline(
		markupEmpty.Row(btnSettingsUser),
	)

	markupSettingsUser.Inline(
		markupEmpty.Row(btnUpdPersonalUser),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupPrivacyExplanation.Inline(
		markupEmpty.Row(btnPrivacyUser),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupPrivacyUser.Inline(
		markupEmpty.Row(btnUserAgreeWithPrivacy),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupPersonalUser.Inline(
		markupEmpty.Row(btnUpdNameUser),
		markupEmpty.Row(btnUpdPhoneUser),
		markupEmpty.Row(btnBackToMainUser),
	)

	markupBackToMainUser.Inline(
		markupEmpty.Row(btnBackToMainUser),
	)
}
