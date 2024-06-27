package telegram

import (
	ent "barbershop-bot/entities"
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
	- длина названия услуги должна быть не менее 3 символов и не более 35 символов.
Пожалуйста, попробуйте ввести название услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	enterServiceDescription   = "Введите описание услуги."
	invalidServiceDescription = `Введенное описание услуги не соответствует установленным критериям:
	- описание услуги может содержать русские и английские буквы (минимум 7 букв в описании), цифры, пробелы, запятые, точки, кавычки, а также знаки +, -, (, ), /, !;
	- длина описания услуги должна быть не менее 10 символов и не более 400 символов.
Пожалуйста, попробуйте ввести описание услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	enterServicePrice   = "Введите цену услуги в рублях. Цена услуги должна быть больше нуля. Нужно ввести только число, дополнительные символы, какие-либо сокращения вводить не нужно."
	invalidServicePrice = `Неизвестный формат цены. Введите число, равное стоимости оказания услуги в рублях. Цена услуги должна быть больше нуля. Нужно ввести только число, дополнительные символы, какие-либо сокращения вводить не нужно.
Пожалуйста, попробуйте ввести цену услуги еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	selectServiceDuration      = "Выберите продолжительность услуги."
	endpntEnterServiceDuration = "enter_service_duration"
	invalidService             = "Невозможно выполнить команду. Недопустимые параметры услуги."
	nonUniqueServiceName       = "Услуга не сохранена. У Вас уже есть другая услуга с таким же названием. Измените название услуги перед сохранением."
	serviceCreated             = "Услуга успешно создана!"

	continueEditingOrSelectService = "Ранее Вы уже начали редактировать услугу. Хотите продолжить с того места, где остановились? Или хотите заново выбрать услугу и начать ее редактировать?"
	selectServiceToEdit            = "Выберите услугу, которую Вы хотите отредактировать."
	endpntServiceToEdit            = "service_to_edit"
	editServiceParams              = `Для изменения параметров услуги выберите соответствующую опцию.
Вы также можете покинуть это меню и вернуться к редактированию услуги позднее.`
	readyToUpdateService      = `Для того, чтобы внесенные изменения вступили в силу, нажмите на кнопку "Применить изменения".`
	endpntEditServiceDuration = "edit_service_duration"
	serviceUpdated            = "Услуга успешно изменена!"

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
	btnSettingsBarber = markupEmpty.Data("Настройки", "settings_barber")

	markupSettingsBarber   = &tele.ReplyMarkup{}
	btnManageAccountBarber = markupEmpty.Data("Управление аккаунтом", "manage_account_barber")
	btnManageServices      = markupEmpty.Data("Управление услугами", "manage_services")
	btnManageBarbers       = markupEmpty.Data("Управление барберами", "manage_barbers")

	markupManageAccountBarber    = &tele.ReplyMarkup{}
	btnShowCurrentSettingsBarber = markupEmpty.Data("Мои текущие настройки", "show_current_settings_barber")
	btnUpdPersonalBarber         = markupEmpty.Data("Обновить персональные данные", "upd_personal_data_barber")
	btnDeleteAccount             = markupEmpty.Data("Удаление аккаунта барбера", "delete_account")

	markupPersonalBarber = &tele.ReplyMarkup{}
	btnUpdNameBarber     = markupEmpty.Data("Обновить имя", "upd_name_barber")
	btnUpdPhoneBarber    = markupEmpty.Data("Обновить номер телефона", "upd_phone_barber")

	markupDeleteAccount   = &tele.ReplyMarkup{}
	btnSetLastWorkDate    = markupEmpty.Data("Установить последний рабочий день", endpntSelectMonthOfLastWorkDate, "0")
	btnSelectLastWorkDate = markupEmpty.Data("", endpntSelectLastWorkDate)
	btnSelfDeleteBarber   = markupEmpty.Data(`Отказаться от статуса "барбер"`, "self_delete_barber")

	markupConfirmSelfDeletion = &tele.ReplyMarkup{}
	btnSureToDelete           = markupEmpty.Data("Уверен, удалить!", "sure_to_delete")

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
		markupEmpty.Row(btnManageAccountBarber),
		markupEmpty.Row(btnManageServices),
		markupEmpty.Row(btnManageBarbers),
		markupEmpty.Row(btnBackToMainBarber),
	)

	markupManageAccountBarber.Inline(
		markupEmpty.Row(btnShowCurrentSettingsBarber),
		markupEmpty.Row(btnUpdPersonalBarber),
		markupEmpty.Row(btnDeleteAccount),
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
		markupEmpty.Row(btnSureToDelete),
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

func btnServiceDuration(dur tm.Duration, endpnt string) tele.Btn {
	return markupEmpty.Data(dur.LongString(), endpnt, strconv.FormatUint(uint64(dur), 10))
}

func btnServiceToEdit(serv ent.Service) tele.Btn {
	return markupEmpty.Data(serv.BtnSignature(), endpntServiceToEdit, strconv.Itoa(serv.ID))
}
