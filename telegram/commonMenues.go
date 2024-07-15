package telegram

import (
	ent "barbershop-bot/entities"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	mainMenu = "Добрый день. Вы находитесь в главном меню."

	settingsMenu = "В этом меню собраны функции, обеспечивающие управление и настройки приложения."

	selectPersonalData = "Выберите данные, которые вы хотите обновить."

	updName        = "Введите Ваше имя"
	updNameSuccess = "Имя успешно обновлено. Хотите обновить другие данные?"
	invalidName    = `Введенное имя не соответствует установленным критериям:
- имя может содержать русские и английские буквы, цифры и пробелы;
- длина имени должна быть не менее 2 символов и не более 20 символов.
Пожалуйста, попробуйте ввести имя еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	updPhone        = "Введите свой номер телефона"
	updPhoneSuccess = "Номер телефона успешно обновлен. Хотите обновить другие данные?"
	invalidPhone    = `Неизвестный формат номера телефона. Примеры поддерживаемых форматов:
81234567890
8(123)-456-7890
+71234567890
+7 123 456 7890
Номер должен начинаться с "+7" или с "8". Региональный код можно по желанию заключить в скобки, также допустимо разделять пробелами или тире группы цифр.
Пожалуйста, попробуйте ввести номер телефона еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`

	unknownCommand = "Неизвестная команда. Пожалуйста, выберите команду из меню. Для вызова главного меню воспользуйтесь командой /start"

	endpntNoAction = "no_action"
)

var monthNames = []string{
	"Январь",
	"Февраль",
	"Март",
	"Апрель",
	"Май",
	"Июнь",
	"Июль",
	"Август",
	"Сентябрь",
	"Октябрь",
	"Ноябрь",
	"Декабрь",
}

var (
	markupEmpty = &tele.ReplyMarkup{}
	btnEmpty    = markupEmpty.Data("-", endpntNoAction)
	rowWeekdays = tele.Row{
		markupEmpty.Data("Пн", endpntNoAction),
		markupEmpty.Data("Вт", endpntNoAction),
		markupEmpty.Data("Ср", endpntNoAction),
		markupEmpty.Data("Чт", endpntNoAction),
		markupEmpty.Data("Пт", endpntNoAction),
		markupEmpty.Data("Сб", endpntNoAction),
		markupEmpty.Data("Вс", endpntNoAction),
	}
)

func btnBarber(barb ent.Barber, endpnt string) tele.Btn {
	return markupEmpty.Data(barb.Name, endpnt, strconv.FormatInt(barb.ID, 10))
}

func btnMonth(month time.Month) tele.Btn {
	return markupEmpty.Data(monthNames[month-1], endpntNoAction)
}

func btnNext(switchTo byte, endpnt string) tele.Btn {
	return markupEmpty.Data("➡", endpnt, strconv.FormatUint(uint64(switchTo), 10))
}

func btnPrev(switchTo byte, endpnt string) tele.Btn {
	return markupEmpty.Data("⬅", endpnt, strconv.FormatUint(uint64(switchTo), 10))
}

func btnService(serv ent.Service, endpnt string) tele.Btn {
	return markupEmpty.Data(serv.BtnSignature(), endpnt, strconv.Itoa(serv.ID))
}

// btnsSwitch returns prev and next buttons for switching in range from 0 to maxDelta.
func btnsSwitch(delta, maxDelta byte, endpnt string) (prev, next tele.Btn) {
	if maxDelta == 0 {
		prev = btnEmpty
		next = btnEmpty
		return
	}
	switch delta {
	case 0:
		prev = btnEmpty
		next = btnNext(1, endpnt)
	case maxDelta:
		prev = btnPrev(delta-1, endpnt)
		next = btnEmpty
	default:
		prev = btnPrev(delta-1, endpnt)
		next = btnNext(delta+1, endpnt)
	}
	return
}

func noAction(tele.Context) error { return nil }
