package telegram

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	tm "barbershop-bot/lib/time"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	listOfNecessarySettings = `Прежде чем клиенты получат возможность записаться к Вам на стрижку через этот бот, Вы должны произвести необходимый минимум подготовительных настроек.
Это необходимо для того, чтобы предоставить Вашим клиентам максимально комфортный пользовательский опыт обращения с этим ботом.
Итак, что необходимо сделать:

1. Через меню управления аккаунтом введите свое имя для отображения клиентам.
2. Через меню управления услугами создайте минимум одну услугу, которую вы будете предоставлять клиентам.`

	manageAccountBarber = "В этом меню собраны функции для управления Вашим аккаунтом."
	currentSettings     = `Ваши текущие настройки:
`
	privacyBarber       = "Текст политики конфиденциальности для барберов."
	notUniqueBarberName = `Имя не сохранено. Другой барбер с таким именем уже зарегистрирован в приложении. Введите другое имя.
При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberPhone = `Номер телефона не сохранен. Другой барбер с таким номером уже зарегистрирован в приложении. Введите другой номер.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	deleteAccount      = "В этом меню собраны функции, необходимые для корректного прекращения работы в качестве барбера в этом боте."
	selectLastWorkDate = `Данную функцию следует использовать, если Вы планируете прекратить использовать этот бот в своей работе и хотите, чтобы клиенты перестали использовать бот для записи к Вам на стрижку.
	
Выберите дату последнего рабочего дня.
Для выбора доступны только даты позже последней существующей записи клиента на стрижку.
Если Вы хотите выбрать более раннюю дату, сначала отмените последние записи клиентов на стрижку.
	
ВНИМАНИЕ!!! После установки даты последнего рабочего дня клиенты не смогут записаться на стрижку на более позднюю дату.
По умолчанию установлена бессрочная дата последнего рабочего дня.
Если Вы ранее меняли эту дату и хотите снова установить бессрочную дату, нажмите кнопку в нижней части меню.`
	haveAppointmentAfterDataToSave = `Невозможно сохранить дату, поскольку пока Вы выбирали дату, клиент успел записаться к Вам на стрижку на дату позже выбранной Вами даты.
Пожалуйста проверьте записи клиентов и при необходимости отмените самую позднюю запись. Или выберите более позднюю дату последнего рабочего дня.`
	lastWorkDateSaved                = "Новая дата последнего рабочего дня успешно сохранена."
	lastWorkDateUnchanged            = "Указанная дата совпадает с той, которую вы уже установили ранее"
	lastWorkDateSavedWithoutSсhedule = `Новая дата последнего рабочего дня сохранена.
ВНИМАНИЕ!!! При попытке составить расписание работы вплоть до сохраненной даты произошла ошибка. Расписание не составлено!
Для доступа к записи клиентов на стрижку необходимо составить расписание работы.`
	lastWorkDateNotSavedButScheduleDeleted = `При выполнении команды произошла ошибка, в результате которой Ваше расписание работы после желаемой даты окончания работы было удалено.
Однако сама дата окончания работы не была настроена. Планировшик расписания восстановит удаленную часть расписания во время планового запуска. Однако, дата последнего рабочего дня так и останется не измененной.
Чтобы исправить это, пожалуйста попробуйте еще раз установить интересующую Вас дату последнего рабочего дня.`
	confirmSelfDeletion = `Вы собираетесь отказаться от статуса "барбер".
ВНИМАНИЕ!!! Помимо изменения Вашего статуса также будет удален весь перечень оказываемых Вами услуг и вся история прошедших записей Ваших клиентов. Клиенты больше не смогут записаться к Вам на стрижку через этот бот.
Если Вы уверены, нажмите "Уверен, удалить!". Если передумали, просто вернитесь в главное меню.`
	youHaveActiveSchedule = `Невозможно отказаться от статуса "барбер" прямо сейчас. Предварительно вы должны выполнить следующие действия:
`
	goodbuyBarber = "Ваш статус изменен. Спасибо, что работали с нами!"

	manageServices              = "В этом меню собраны функции для управления услугами. Каждый барбер настраивает свои услуги индивидуально."
	youHaveNoServices           = "У Вас нет ни одной услуги."
	yourServices                = "Список Ваших услуг:"
	continueOldOrMakeNewService = "Ранее Вы уже начали создавать услугу. Хотите продолжить с того места, где остановились? Или хотите начать с самого начала создание новой услуги?"
	readyToCreateService        = "\n\nВы ввели всю необходимую информацию об услуге и можете сохранить ее прямо сейчас, нажав на кнопку \"Сохранить новую услугу\"."
	enterServiceParams          = "Для ввода или изменения параметров услуги выберите соответствующую опцию. Вы также можете покинуть это меню и вернуться к созданию услуги позднее.\n\n"
	enterServiceName            = "Введите название услуги. Название услуги не должно совпадать с названиями других Ваших услуг."
	invalidServiceName          = `Введенное название услуги не соответствует установленным критериям:
	- название услуги может содержать любые буквы, цифры, пробелы, знаки пунктуации, а также знаки + и -;
	- длина названия услуги должна быть не менее 3 символов и не более 35 символов.
Пожалуйста, попробуйте ввести название услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	enterServiceDescription   = "Введите описание услуги."
	invalidServiceDescription = `Введенное описание услуги не соответствует установленным критериям:
	- описание услуги может содержать любые буквы (минимум 7 букв в описании), цифры, пробелы, знаки пунктуации, а также знаки + и -;
	- длина описания услуги должна быть не менее 10 символов и не более 400 символов.
Пожалуйста, попробуйте ввести описание услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	enterServicePrice   = "Введите цену услуги в рублях. Цена услуги должна быть больше нуля. Нужно ввести только число, дополнительные символы, какие-либо сокращения вводить не нужно."
	invalidServicePrice = `Неизвестный формат цены. Введите число, равное стоимости оказания услуги в рублях. Цена услуги должна быть больше нуля. Нужно ввести только число, дополнительные символы, какие-либо сокращения вводить не нужно.
Пожалуйста, попробуйте ввести цену услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	selectServiceDuration = "Выберите продолжительность услуги."
	invalidService        = "Невозможно выполнить команду. Недопустимые параметры услуги."
	nonUniqueServiceName  = "Услуга не сохранена. У Вас уже есть другая услуга с таким же названием. Измените название услуги перед сохранением."
	serviceCreated        = "Услуга успешно создана!"

	continueEditingOrSelectService = "Ранее Вы уже начали редактировать услугу. Хотите продолжить с того места, где остановились? Или хотите заново выбрать услугу и начать ее редактировать?"
	selectServiceToEdit            = "Выберите услугу, которую Вы хотите отредактировать."
	editServiceParams              = "Для изменения параметров услуги выберите соответствующую опцию. Вы также можете покинуть это меню и вернуться к редактированию услуги позднее.\n\n"
	readyToUpdateService           = "\n\nДля того, чтобы внесенные изменения вступили в силу, нажмите на кнопку \"Применить изменения\"."
	serviceUpdated                 = "Услуга успешно изменена!"

	selectServiceToDelete  = "Выберите услугу, которую Вы хотите удалить."
	confirmServiceDeletion = "Вы выбрали для удаления следующую услугу:\n\n%s\n\nПодтвердите удаление услуги. Если передумали, просто вернитесь в главное меню."
	serviceDeleted         = "Услуга успешно удалена!"

	manageBarbers = "В этом меню собраны функции для управления барберами."
	listOfBarbers = "Список всех барберов, зарегистрированных в приложении:"
	addBarber     = `ВНИМАНИЕ!!! Добавленного барбера потом невозможно будет удалить без его личного содействия. Убедитесь, что Вы добавляете нужного человека.

Для добавления нового барбера пришлите в этот чат контакт профиля пользователя телеграм, которого вы хотите сделать барбером.
Подробная инструкция:
1. Зайдите в личный чат с пользователем.
2. В верхней части чата нажмите на поле, отображающее аватар аккаунта пользователя и его имя. Таким образом Вы откроете окно просмотра профиля пользователя.
3. Нажмите на "три точки" в верхнем правом углу дисплея.
4. В открывшемся меню выберите "Поделиться контактом/отправить контакт". Если этой опции нет, попросите пользователя изменить настройки конфиденциальности в профиле телеграм и разрешить другим пользователям делиться его контактом. После того, как пользователь получит статус барбера, он может вернуть любые желаемые настройки профиля.
5. В открывшемся списке Ваших чатов выберите чат с ботом (чат, в котором Вы читаете эту инструкцию).
6. В правом нижнем углу нажмите на значек отправки сообщения.

Если вы передумали добавлять нового барбера, воспользуйтесь командой /start для возврата в главное меню.`
	userIsAlreadyBarber        = "Статус пользователя не изменен, поскольку он уже является барбером."
	addedNewBarberWithSсhedule = `Статус пользователя изменен на "барбер". Для нового барбера составлено расписание работы на ближайшие полгода.
Для доступа к записи клиентов на стрижку новый барбер должен заполнить персональные данные.`
	addedNewBarberWithoutSсhedule = `Статус пользователя изменен на "барбер".
ВНИМАНИЕ!!! При попытке составить расписание работы для нового барбера произошла ошибка. Расписание не составлено!
Для доступа к записи клиентов на стрижку новый барбер должен заполнить персональные данные, а также составить расписание работы.`
	onlyOneBarberExists = "Вы единственный зарегистрированный барбер в приложении. Некого удалять."
	noBarbersToDelete   = `Нет ни одного барбера, которого можно было бы удалить. 
Для того, чтобы барбера можно было удалить, он предварительно должен выполнить следующие действия:
`
	selectBarberToDeletion = `Выберите барбера, которого Вы хотите удалить.
ВНИМАНИЕ!!! При удалении барбера будет также удален весь перечень оказываемых им услуг и вся история прошедших записей клиентов этого барбера. 
Если нужного барбера нет в этом списке, значит он еще не выполнил необходимые действия перед удалением:
`
	preDeletionBarberInstruction = `1. Установить не бессрочную дату последнего рабочего дня. По умолчанию установлена бессрочная дата последнего рабочего дня.
2. Дождаться наступления следующего дня после даты установленной в прошлом пункте.
	
Эти действия необходимы, чтобы гарантировать, что ни один клиент, записавшийся на стрижку, не останется необслуженным в процессе удаления барбера из приложения.`
	barberDeleted            = `Статус указанного пользователя изменен. Пользователь больше не имеет статуса "барбер".`
	barberHaveActiveSchedule = `Невозможно изменить статус барбера, поскольку пока Вы выбирали барбера для удаления, выбранный барбер изменил свою дату последнего рабочего дня на более позднюю.`

	errorBarber = `Произошла ошибка обработки команды. Команда не была выполнена. Если ошибка будет повторяться, возможно, потребуется перезапуск сервиса.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`

	endpntSelectMonthOfLastWorkDate = "select_month_of_last_work_date"
	endpntSelectLastWorkDate        = "select_last_work_date"
	endpntEnterServiceDuration      = "enter_service_duration"
	endpntServiceToEdit             = "service_to_edit"
	endpntEditServiceDuration       = "edit_service_duration"
	endpntServiceToDelete           = "service_to_delete"
	endpntSureToDeleteService       = "sure_to_delete_service"
	endpntBarberToDeletion          = "barber_to_deletion"
)

var (
	markupMainBarber  = &tele.ReplyMarkup{}
	btnSettingsBarber = markupEmpty.Data("Настройки", "settings_barber")

	markupSettingsBarber       = &tele.ReplyMarkup{}
	btnListOfNecessarySettings = markupEmpty.Data("Перечень необходимых настроек", "list_of_necessary_settings")
	btnManageAccountBarber     = markupEmpty.Data("Управление аккаунтом", "manage_account_barber")
	btnManageServices          = markupEmpty.Data("Управление услугами", "manage_services")
	btnManageBarbers           = markupEmpty.Data("Управление барберами", "manage_barbers")

	markupShortSettingsBarber = &tele.ReplyMarkup{}

	markupManageAccountBarber    = &tele.ReplyMarkup{}
	btnShowCurrentSettingsBarber = markupEmpty.Data("Мои текущие настройки", "show_current_settings_barber")
	btnUpdPersonalBarber         = markupEmpty.Data("Обновить персональные данные", "upd_personal_data_barber")
	btnDeleteAccount             = markupEmpty.Data("Удаление аккаунта барбера", "delete_account")

	markupPrivacyBarber       = &tele.ReplyMarkup{}
	btnBarberAgreeWithPrivacy = markupEmpty.Data("Соглашаюсь с политикой конфиденциальности", "barber_agree_with_privacy")

	markupPersonalBarber = &tele.ReplyMarkup{}
	btnUpdNameBarber     = markupEmpty.Data("Обновить имя", "upd_name_barber")
	btnUpdPhoneBarber    = markupEmpty.Data("Обновить номер телефона", "upd_phone_barber")

	markupDeleteAccount     = &tele.ReplyMarkup{}
	btnSetLastWorkDate      = markupEmpty.Data("Установить последний рабочий день", endpntSelectMonthOfLastWorkDate, "0")
	btnSelectLastWorkDate   = markupEmpty.Data("", endpntSelectLastWorkDate)
	btnInfiniteLastWorkDate = markupEmpty.Data("Установить бессрочную дату", endpntSelectLastWorkDate, cfg.InfiniteWorkDate)
	btnSelfDeleteBarber     = markupEmpty.Data(`Отказаться от статуса "барбер"`, "self_delete_barber")

	markupConfirmSelfDeletion = &tele.ReplyMarkup{}
	btnSureToSelfDeleteBarber = markupEmpty.Data("Уверен, удалить!", "sure_to_self_delete_barber")

	markupManageServicesFull = &tele.ReplyMarkup{}
	btnShowMyServices        = markupEmpty.Data("Список моих услуг", "show_my_services")
	btnCreateService         = markupEmpty.Data("Создать услугу", "create_service")
	btnEditService           = markupEmpty.Data("Изменить услугу", "edit_service")
	btnDeleteService         = markupEmpty.Data("Удалить услугу", "delete_services")

	markupManageServicesShort = &tele.ReplyMarkup{}
	markupShowMyServices      = &tele.ReplyMarkup{}

	markupСontinueOldOrMakeNewService = &tele.ReplyMarkup{}
	btnСontinueOldService             = markupEmpty.Data("Продолжить ранее начатое", "continue_old_service")
	btnMakeNewService                 = markupEmpty.Data("Начать заново", "make_new_service")

	markupEnterServiceParams        = &tele.ReplyMarkup{}
	btnEnterServiceName             = markupEmpty.Data("Ввести название услуги", "enter_service_name")
	btnEnterServiceDescription      = markupEmpty.Data("Ввести описание услуги", "enter_service_description")
	btnEnterServicePrice            = markupEmpty.Data("Ввести цену услуги", "enter_service_price")
	btnSelectServiceDurationOnEnter = markupEmpty.Data("Выбрать продолжительность услуги", "select_service_duration_on_enter")
	btnSelectCertainDurationOnEnter = markupEmpty.Data("", endpntEnterServiceDuration)

	markupReadyToCreateService = &tele.ReplyMarkup{}
	btnSaveNewService          = markupEmpty.Data("Сохранить новую услугу", "save_new_service")

	markupEnterServiceName = &tele.ReplyMarkup{}

	markupContinueEditingOrSelectService = &tele.ReplyMarkup{}
	btnСontinueEditingService            = markupEmpty.Data("Продолжить ранее начатое", "continue_editing")
	btnSelectServiceToEdit               = markupEmpty.Data("Начать заново", "select_service_to_edit")
	btnSelectCertainServiceToEdit        = markupEmpty.Data("", endpntServiceToEdit)

	markupEditServiceParams        = &tele.ReplyMarkup{}
	btnEditServiceName             = markupEmpty.Data("Изменить название услуги", "edit_service_name")
	btnEditServiceDescription      = markupEmpty.Data("Изменить описание услуги", "edit_service_description")
	btnEditServicePrice            = markupEmpty.Data("Изменить цену услуги", "edit_service_price")
	btnSelectServiceDurationOnEdit = markupEmpty.Data("Изменить продолжительность услуги", "select_service_duration_on_edit")
	btnSelectCertainDurationOnEdit = markupEmpty.Data("", endpntEditServiceDuration)

	markupReadyToUpdateService = &tele.ReplyMarkup{}
	btnUpdateService           = markupEmpty.Data("Применить изменения", "update_service")

	markupEditServiceName = &tele.ReplyMarkup{}

	btnSelectCertainServiceToDelete = markupEmpty.Data("", endpntServiceToDelete)
	btnSureToDeleteService          = markupEmpty.Data("", endpntSureToDeleteService)

	markupManageBarbers    = &tele.ReplyMarkup{}
	btnShowAllBurbers      = markupEmpty.Data("Список барберов", "show_all_barbers")
	btnAddBarber           = markupEmpty.Data("Добавить барбера", "add_barber")
	btnDeleteBarber        = markupEmpty.Data("Удалить барбера", "delete_barber")
	btnDeleteCertainBarber = markupEmpty.Data("", endpntBarberToDeletion)

	markupBackToMainBarber = &tele.ReplyMarkup{}
	btnBackToMainBarber    = markupEmpty.Data("Вернуться в главное меню", "back_to_main_barber")
)

func init() {
	markupMainBarber.Inline(
		markupEmpty.Row(btnSettingsBarber),
	)

	markupSettingsBarber.Inline(
		markupEmpty.Row(btnListOfNecessarySettings),
		markupEmpty.Row(btnManageAccountBarber),
		markupEmpty.Row(btnManageServices),
		markupEmpty.Row(btnManageBarbers),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupShortSettingsBarber.Inline(
		markupEmpty.Row(btnManageAccountBarber),
		markupEmpty.Row(btnManageServices),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupManageAccountBarber.Inline(
		markupEmpty.Row(btnShowCurrentSettingsBarber),
		markupEmpty.Row(btnUpdPersonalBarber),
		markupEmpty.Row(btnDeleteAccount),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupPrivacyBarber.Inline(
		markupEmpty.Row(btnBarberAgreeWithPrivacy),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupPersonalBarber.Inline(
		markupEmpty.Row(btnUpdNameBarber),
		markupEmpty.Row(btnUpdPhoneBarber),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupDeleteAccount.Inline(
		markupEmpty.Row(btnSetLastWorkDate),
		markupEmpty.Row(btnSelfDeleteBarber),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupConfirmSelfDeletion.Inline(
		markupEmpty.Row(btnSureToSelfDeleteBarber),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupManageServicesFull.Inline(
		markupEmpty.Row(btnShowMyServices),
		markupEmpty.Row(btnCreateService),
		markupEmpty.Row(btnEditService),
		markupEmpty.Row(btnDeleteService),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupManageServicesShort.Inline(
		markupEmpty.Row(btnCreateService),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupShowMyServices.Inline(
		markupEmpty.Row(btnCreateService),
		markupEmpty.Row(btnEditService),
		markupEmpty.Row(btnDeleteService),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupСontinueOldOrMakeNewService.Inline(
		markupEmpty.Row(btnСontinueOldService),
		markupEmpty.Row(btnMakeNewService),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupEnterServiceParams.Inline(
		markupEmpty.Row(btnEnterServiceName),
		markupEmpty.Row(btnEnterServiceDescription),
		markupEmpty.Row(btnEnterServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEnter),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupReadyToCreateService.Inline(
		markupEmpty.Row(btnSaveNewService),
		markupEmpty.Row(btnEnterServiceName),
		markupEmpty.Row(btnEnterServiceDescription),
		markupEmpty.Row(btnEnterServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEnter),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupEnterServiceName.Inline(
		markupEmpty.Row(btnEnterServiceName),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupContinueEditingOrSelectService.Inline(
		markupEmpty.Row(btnСontinueEditingService),
		markupEmpty.Row(btnSelectServiceToEdit),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupEditServiceParams.Inline(
		markupEmpty.Row(btnEditServiceName),
		markupEmpty.Row(btnEditServiceDescription),
		markupEmpty.Row(btnEditServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEdit),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupReadyToUpdateService.Inline(
		markupEmpty.Row(btnUpdateService),
		markupEmpty.Row(btnEditServiceName),
		markupEmpty.Row(btnEditServiceDescription),
		markupEmpty.Row(btnEditServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEdit),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupEditServiceName.Inline(
		markupEmpty.Row(btnEditServiceName),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupManageBarbers.Inline(
		markupEmpty.Row(btnShowAllBurbers),
		markupEmpty.Row(btnAddBarber),
		markupEmpty.Row(btnDeleteBarber),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupBackToMainBarber.Inline(
		markupEmpty.Row(btnBackToMainBarber),
	)
}

func btnDate(date time.Time, endpnt string) tele.Btn {
	return markupEmpty.Data(strconv.Itoa(date.Day()), endpnt, date.Format(time.DateOnly))
}

func btnServiceDuration(dur tm.Duration, endpnt string) tele.Btn {
	return markupEmpty.Data(dur.LongString(), endpnt, strconv.FormatUint(uint64(dur), 10))
}

func markupConfirmServiceDeletion(serviceID int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("Подтвердить удаление", endpntSureToDeleteService, strconv.Itoa(serviceID))),
		markup.Row(btnBackToMainBarber),
	)
	return markup
}

func markupSelectBarberToDeletion(senderID int64, barbers []ent.Barber) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	today := tm.Today()
	var rows []tele.Row
	for _, barber := range barbers {
		if barber.ID != senderID && barber.LastWorkdate.Before(today) {
			rows = append(rows, markup.Row(btnBarber(barber, endpntBarberToDeletion)))
		}
	}
	rows = append(rows, markup.Row(btnBackToMainBarber))
	markup.Inline(rows...)
	return markup
}

func markupSelectLastWorkDate(dateRange ent.DateRange, deltaMonth, maxDeltaMonth byte) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	btnPrevMonth, btnNextMonth := btnsSwitch(deltaMonth, maxDeltaMonth, endpntSelectMonthOfLastWorkDate)
	rowSelectMonth := markup.Row(btnPrevMonth, btnMonth(dateRange.Month()), btnNextMonth)
	rowsSelectDate := rowsSelectLastWorkDate(dateRange)
	rowRestoreDefaultDate := markup.Row(btnInfiniteLastWorkDate)
	rowBackToMainBarber := markup.Row(btnBackToMainBarber)
	var rows []tele.Row
	rows = append(rows, rowSelectMonth, rowWeekdays)
	rows = append(rows, rowsSelectDate...)
	rows = append(rows, rowRestoreDefaultDate, rowBackToMainBarber)
	markup.Inline(rows...)
	return markup
}

func markupSelectServiceDuration(endpnt string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	btnsDurationsToSelect := make([]tele.Btn, 4)
	for duration := 30 * tm.Minute; duration <= 2*tm.Hour; duration += 30 * tm.Minute {
		btnsDurationsToSelect = append(btnsDurationsToSelect, btnServiceDuration(duration, endpnt))
	}
	rows := markup.Split(2, btnsDurationsToSelect)
	rows = append(rows, markup.Row(btnBackToMainBarber))
	markup.Inline(rows...)
	return markup
}

func markupSelectServiceBarber(services []ent.Service, endpnt string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, service := range services {
		row := markup.Row(btnService(service, endpnt))
		rows = append(rows, row)
	}
	rows = append(rows, markup.Row(btnBackToMainBarber))
	markup.Inline(rows...)
	return markup
}

func rowsSelectLastWorkDate(dateRange ent.DateRange) []tele.Row {
	markup := &tele.ReplyMarkup{}
	var btnsDatesToSelect []tele.Btn
	for i := 1; i < dateRange.StartWeekday(); i++ {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	for date := dateRange.StartDate; date.Compare(dateRange.EndDate) <= 0; date = date.Add(24 * time.Hour) {
		btnsDatesToSelect = append(btnsDatesToSelect, btnDate(date, endpntSelectLastWorkDate))
	}
	for i := 7; i > dateRange.EndWeekday(); i-- {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	return markup.Split(7, btnsDatesToSelect)
}
