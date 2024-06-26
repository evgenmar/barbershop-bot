package telegram

import (
	tm "barbershop-bot/lib/time"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

const (
	mainBarber = "Добрый день. Вы находитесь в главном меню."

	settingsBarber = "В этом меню собраны функции, обеспечивающие управление и настройки приложения."

	manageAccountBarber = "В этом меню собраны функции для управления Вашим аккаунтом."
	currentSettings     = `Ваши текущие настройки:
`
	personalBarber       = "Выберите данные, которые вы хотите обновить."
	updNameBarber        = "Введите Ваше имя"
	updNameSuccessBarber = "Имя успешно обновлено. Хотите обновить другие данные?"
	invalidBarberName    = `Введенное имя не соответствует установленным критериям:
		- имя может содержать русские и английские буквы, цифры и пробелы;
		- длина имени должна быть не менее 2 символов и не более 20 символов.
Пожалуйста, попробуйте ввести имя еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberName = `Имя не сохранено. Другой барбер с таким именем уже зарегистрирован в приложении. Введите другое имя.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	updPhoneBarber        = "Введите свой номер телефона"
	updPhoneSuccessBarber = "Номер телефона успешно обновлен. Хотите обновить другие данные?"
	invalidBarberPhone    = `Неизвестный формат номера телефона. Примеры поддерживаемых форматов:
		81234567890
		8(123)-456-7890
		+71234567890
		+7 123 456 7890
Номер должен начинаться с "+7" или с "8". Региональный код можно по желанию заключить в скобки, также допустимо разделять пробелами или тире группы цифр.
Пожалуйста, попробуйте ввести номер телефона еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	notUniqueBarberPhone = `Номер телефона не сохранен. Другой барбер с таким номером уже зарегистрирован в приложении. Введите другой номер.
При необходимости вернуться в главное меню воспользуйтесь командой /start`

	deleteAccount                   = "В этом меню собраны функции, необходимые для корректного прекращения работы в качестве барбера в этом боте."
	endpntSelectMonthOfLastWorkDate = "select_month_of_last_work_date"
	endpntSelectLastWorkDate        = "select_last_work_date"
	selectLastWorkDate              = `Данную функцию следует использовать, если Вы планируете прекратить использовать этот бот в своей работе и хотите, чтобы клиенты перестали использовать бот для записи к Вам на стрижку.
	
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
	readyToCreateService        = `Вы ввели всю необходимую информацию об услуге и можете сохранить ее прямо сейчас, нажав на кнопку "Сохранить новую услугу".`
	enterServiceParams          = `Для ввода или изменения параметров услуги выберите соответствующую опцию.
Вы также можете покинуть это меню и вернуться к созданию услуги позднее.`
	enterServiceName   = "Введите название услуги. Название услуги не должно совпадать с названиями других Ваших услуг."
	invalidServiceName = `Введенное название услуги не соответствует установленным критериям:
	- название услуги может содержать русские и английские буквы, цифры, пробелы, запятые, точки, кавычки, а также знаки +, -, (, ), /, !;
	- длина названия услуги должна быть не менее 3 символов и не более 50 символов.
Пожалуйста, попробуйте ввести название услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	enterServiceDescription   = "Введите описание услуги."
	invalidServiceDescription = `Введенное описание услуги не соответствует установленным критериям:
	- описание услуги может содержать русские и английские буквы (минимум 7 букв в описании), цифры, пробелы, запятые, точки, кавычки, а также знаки +, -, (, ), /, !;
	- длина описания услуги должна быть не менее 10 символов и не более 400 символов.
Пожалуйста, попробуйте ввести описание услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	enterServicePrice   = "Введите цену услуги в рублях. Цена услуги должна быть больше нуля. Нужно ввести только число, дополнительные символы, какие-либо сокращения вводить не нужно."
	invalidServicePrice = `Неизвестный формат цены. Введите число, равное стоимости оказания услуги в рублях. Цена услуги должна быть больше нуля. Нужно ввести только число, дополнительные символы, какие-либо сокращения вводить не нужно.
Пожалуйста, попробуйте ввести цену услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	selectServiceDuration = "Выберите продолжительность услуги."
	endpntServiceDuration = "service_duration"
	invalidService        = "Недопустимые параметры услуги. Попробуйте создать услугу заново с другими параметрами."
	nonUniqueServiceName  = "Услуга не сохранена. У Вас уже есть другая услуга с таким же названием. Измените название услуги перед сохранением."
	serviceCreated        = "Услуга успешно создана!"

	manageBarbers = "В этом меню собраны функции для управления барберами."
	listOfBarbers = "Список всех барберов, зарегистрированных в приложении:"
	addBarber     = `Для добавления нового барбера пришлите в этот чат контакт профиля пользователя телеграм, которого вы хотите сделать барбером.
Подробная инструкция:
1. Зайдите в личный чат с пользователем.
2. В верхней части чата нажмите на поле, отображающее аватар аккаунта пользователя и его имя. Таким образом Вы откроете окно просмотра профиля пользователя.
3. Нажмите на "три точки" в верхнем правом углу дисплея.
4. В открывшемся меню выберите "Поделиться контактом".
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
	endpntBarberToDeletion   = "barber_to_deletion"
	barberDeleted            = `Статус указанного пользователя изменен. Пользователь больше не имеет статуса "барбер".`
	barberHaveActiveSchedule = `Невозможно изменить статус барбера, поскольку пока Вы выбирали барбера для удаления, выбранный барбер изменил свою дату последнего рабочего дня на более позднюю.`

	unknownCmdBarber = "Неизвестная команда. Пожалуйста, выберите команду из меню. Для вызова главного меню воспользуйтесь командой /start"
	errorBarber      = `Произошла ошибка обработки команды. Команда не была выполнена. Если ошибка будет повторяться, возможно, потребуется перезапуск сервиса.
Пожалуйста, перейдите в главное меню и попробуйте выполнить команду заново.`
)

var (
	markupMainBarber  = &tele.ReplyMarkup{}
	btnSettingsBarber = markupMainBarber.Data("Настройки", "settings_barber")

	markupSettingsBarber   = &tele.ReplyMarkup{}
	btnManageAccountBarber = markupSettingsBarber.Data("Управление аккаунтом", "manage_account_barber")
	btnManageServices      = markupSettingsBarber.Data("Управление услугами", "manage_services")
	btnManageBarbers       = markupSettingsBarber.Data("Управление барберами", "manage_barbers")

	markupManageAccountBarber    = &tele.ReplyMarkup{}
	btnShowCurrentSettingsBarber = markupManageAccountBarber.Data("Мои текущие настройки", "show_current_settings_barber")
	btnUpdPersonalBarber         = markupManageAccountBarber.Data("Обновить персональные данные", "upd_personal_data_barber")
	btnDeleteAccount             = markupManageAccountBarber.Data("Удаление аккаунта барбера", "delete_account")

	markupPersonalBarber = &tele.ReplyMarkup{}
	btnUpdNameBarber     = markupPersonalBarber.Data("Обновить имя", "upd_name_barber")
	btnUpdPhoneBarber    = markupPersonalBarber.Data("Обновить номер телефона", "upd_phone_barber")

	markupDeleteAccount   = &tele.ReplyMarkup{}
	btnSetLastWorkDate    = markupDeleteAccount.Data("Установить последний рабочий день", endpntSelectMonthOfLastWorkDate, "0")
	btnSelectLastWorkDate = markupDeleteAccount.Data("", endpntSelectLastWorkDate)
	btnSelfDeleteBarber   = markupDeleteAccount.Data(`Отказаться от статуса "барбер"`, "self_delete_barber")

	markupConfirmSelfDeletion = &tele.ReplyMarkup{}
	btnSureToDelete           = markupConfirmSelfDeletion.Data("Уверен, удалить!", "sure_to_delete")

	markupManageServices = &tele.ReplyMarkup{}
	btnShowMyServices    = markupManageServices.Data("Список моих услуг", "show_my_services")
	btnCreateService     = markupManageServices.Data("Создать услугу", "create_service")
	btnEditService       = markupManageServices.Data("Изменить услугу", "edit_service")
	btnDeleteService     = markupManageServices.Data("Удалить услугу", "delete_services")

	markupShowMyServices = &tele.ReplyMarkup{}
	markupShowNoServices = &tele.ReplyMarkup{}

	markupСontinueOldOrMakeNewService = &tele.ReplyMarkup{}
	btnСontinueOldService             = markupСontinueOldOrMakeNewService.Data("Продолжить ранее начатое", "continue_old_service")
	btnMakeNewService                 = markupСontinueOldOrMakeNewService.Data("Начать заново", "make_new_service")

	markupEnterServiceParams   = &tele.ReplyMarkup{}
	btnEnterServiceName        = markupEnterServiceParams.Data("Ввести название услуги", "enter_service_name")
	btnEnterServiceDescription = markupEnterServiceParams.Data("Ввести описание услуги", "enter_service_description")
	btnEnterServicePrice       = markupEnterServiceParams.Data("Ввести цену услуги", "enter_service_price")
	btnSelectServiceDuration   = markupEnterServiceParams.Data("Выбрать продолжительность услуги", "select_service_duration")
	btnSelectCertainDuration   = markupEnterServiceParams.Data("", endpntServiceDuration)

	markupReadyToCreateService = &tele.ReplyMarkup{}
	btnSaveNewService          = markupReadyToCreateService.Data("Сохранить новую услугу", "save_new_service")

	markupEnterServiceName = &tele.ReplyMarkup{}

	markupManageBarbers    = &tele.ReplyMarkup{}
	btnShowAllBurbers      = markupManageBarbers.Data("Список барберов", "show_all_barbers")
	btnAddBarber           = markupManageBarbers.Data("Добавить барбера", "add_barber")
	btnDeleteBarber        = markupManageBarbers.Data("Удалить барбера", "delete_barber")
	btnDeleteCertainBarber = markupManageBarbers.Data("", endpntBarberToDeletion)

	markupBackToMainBarber = &tele.ReplyMarkup{}
	btnBackToMainBarber    = markupBackToMainBarber.Data("Вернуться в главное меню", "back_to_main_barber")
)

func init() {
	markupMainBarber.Inline(
		markupMainBarber.Row(btnSettingsBarber),
	)

	markupSettingsBarber.Inline(
		markupSettingsBarber.Row(btnManageAccountBarber),
		markupSettingsBarber.Row(btnManageServices),
		markupSettingsBarber.Row(btnManageBarbers),
		markupSettingsBarber.Row(btnBackToMainBarber),
	)

	markupManageAccountBarber.Inline(
		markupManageAccountBarber.Row(btnShowCurrentSettingsBarber),
		markupManageAccountBarber.Row(btnUpdPersonalBarber),
		markupManageAccountBarber.Row(btnDeleteAccount),
		markupManageAccountBarber.Row(btnBackToMainBarber),
	)

	markupPersonalBarber.Inline(
		markupPersonalBarber.Row(btnUpdNameBarber),
		markupPersonalBarber.Row(btnUpdPhoneBarber),
		markupPersonalBarber.Row(btnBackToMainBarber),
	)

	markupDeleteAccount.Inline(
		markupDeleteAccount.Row(btnSetLastWorkDate),
		markupDeleteAccount.Row(btnSelfDeleteBarber),
		markupDeleteAccount.Row(btnBackToMainBarber),
	)

	markupConfirmSelfDeletion.Inline(
		markupConfirmSelfDeletion.Row(btnSureToDelete),
		markupConfirmSelfDeletion.Row(btnBackToMainBarber),
	)

	markupManageServices.Inline(
		markupManageServices.Row(btnShowMyServices),
		markupManageServices.Row(btnCreateService),
		markupManageServices.Row(btnEditService),
		markupManageServices.Row(btnDeleteService),
		markupManageServices.Row(btnBackToMainBarber),
	)

	markupShowNoServices.Inline(
		markupShowNoServices.Row(btnCreateService),
		markupShowNoServices.Row(btnBackToMainBarber),
	)

	markupShowMyServices.Inline(
		markupShowMyServices.Row(btnCreateService),
		markupShowMyServices.Row(btnEditService),
		markupShowMyServices.Row(btnDeleteService),
		markupShowMyServices.Row(btnBackToMainBarber),
	)

	markupСontinueOldOrMakeNewService.Inline(
		markupСontinueOldOrMakeNewService.Row(btnСontinueOldService),
		markupСontinueOldOrMakeNewService.Row(btnMakeNewService),
		markupСontinueOldOrMakeNewService.Row(btnBackToMainBarber),
	)

	markupEnterServiceParams.Inline(
		markupEnterServiceParams.Row(btnEnterServiceName),
		markupEnterServiceParams.Row(btnEnterServiceDescription),
		markupEnterServiceParams.Row(btnEnterServicePrice),
		markupEnterServiceParams.Row(btnSelectServiceDuration),
		markupEnterServiceParams.Row(btnBackToMainBarber),
	)

	markupReadyToCreateService.Inline(
		markupReadyToCreateService.Row(btnSaveNewService),
		markupReadyToCreateService.Row(btnEnterServiceName),
		markupReadyToCreateService.Row(btnEnterServiceDescription),
		markupReadyToCreateService.Row(btnEnterServicePrice),
		markupReadyToCreateService.Row(btnSelectServiceDuration),
		markupReadyToCreateService.Row(btnBackToMainBarber),
	)

	markupEnterServiceName.Inline(
		markupEnterServiceName.Row(btnEnterServiceName),
		markupEnterServiceName.Row(btnBackToMainBarber),
	)

	markupManageBarbers.Inline(
		markupManageBarbers.Row(btnShowAllBurbers),
		markupManageBarbers.Row(btnAddBarber),
		markupManageBarbers.Row(btnDeleteBarber),
		markupManageBarbers.Row(btnBackToMainBarber),
	)

	markupBackToMainBarber.Inline(
		markupBackToMainBarber.Row(btnBackToMainBarber),
	)
}

func btnServiceDuration(dur tm.Duration) tele.Btn {
	return markupEmpty.Data(dur.LongString(), endpntServiceDuration, strconv.FormatUint(uint64(dur), 10))
}
