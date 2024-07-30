package telegram

import (
	cp "barbershop-bot/contextprovider"
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	tm "barbershop-bot/lib/time"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	createServiceFirst                   = "Прежде, чем записывать клиентов, создайте хотя бы одну услугу, которую Вы будете им предоставлять."
	barberSelectServiceForAppointment    = "Выберите услугу, на которую хотите записать клиента."
	informBarberNoFreeTimeForAppointment = `Ваш график полностью занят, нет времени для новой записи.
Попробуйте добавить рабочие часы в свой график или предложите клиенту записаться на услугу меньшей длительности, для которой возможно найдется время в существующем графике.`
	newAppointmentSavedByBarber = "Вы записали клиента на услугу:\n\n%s\n\nВремя записи %s в %s.\n\nЖелаете добавить заметку к только что сделанной записи? Вы можете сделать это позднее, найдя запись в Вашем графике."
	enterNote                   = "Введите заметку к записи клиента."
	invalidNote                 = `Введенная заметка не соответствует установленным критериям:
- заметка может содержать любые буквы, цифры, пробелы, знаки пунктуации, а также знаки + и -;
- длина заметки должна быть не менее 3 символов и не более 100 символов.
Пожалуйста, попробуйте ввести заметку еще раз. При необходимости вернуться в главное меню воспользуйтесь командой /start`
	updNoteSuccess = "Заметка успешно добавлена/обновлена."

	scheduleCalendarIsEmpty = "В Вашем графике работы нет ни одного рабочего дня."
	selectWorkday           = "Выберите рабочий день для просмотра.\nВы также можете добавить новый рабочий день в свой график или, наоборот, сделать день выходным."
	selectAppointment       = "Рабочий день %s с %s до %s.\nВыберите запись для просмотра, редактирования или удаления.\nВы также можете изменить время начала и конца рабочего дня."
	workdayIsFree           = "Рабочий день %s с %s до %s.\nПоскольку на этот рабочий день пока что не запланировано ни одной записи, Вы можете сделать его выходным.\nВы также можете изменить время начала и конца рабочего дня."
	selectWorkdayStartTime  = `Выберите новое время начала рабочего дня. Для выбора доступны варианты:
- не раньше %s;
- не позже начала самой ранней записи клиента;
- не позже, чем за 30 минут до конца рабочего дня.`
	selectWorkdayEndTime = `Выберите новое время конца рабочего дня. Для выбора доступны варианты:
- не позже %s;
- не раньше окончания самой поздней записи клиента;
- не раньше, чем через 30 минут после начала рабочего дня.`
	failToUpdateWorkdayStartTime = "ВНИМАНИЕ!!! Не удалось изменить начало рабочего дня, поскольку появилась новая запись клиента на более раннее время."
	failToUpdateWorkdayEndTime   = "ВНИМАНИЕ!!! Не удалось изменить конец рабочего дня, поскольку появилась новая запись клиента на более позднее время."

	appointmentInfoForBarber               = "Информация о записи:\n\n%s\n\nВремя записи: %s в %s.\n\n%s\n\nВыберите действие."
	appointmentNotRescheduled              = "Перенос записи не выполнен по одной из причин:\n"
	appointmentNotCanceled                 = "Отмена записи не выполнена по одной из причин:\n"
	reasons                                = "- время записи завершилось и запись перенесена в архив;\nили\n- клиент перенес запись на другое время или отменил ее."
	appointmentRescheduledByBarber         = "Уведомляем Вас о переносе записи на новое время.\nИнформация о перенесенной записи:\n\n%s\n\nЗапись перенесена барбером на новое время. Барбер %s ждет Вас %s в %s."
	barberConfirmCancelAppointment         = "Информация о записи:\n\n%s\n\nДата: %s\nВремя: %s\n\nПодтвердите отмену записи или вернитесь в главное меню."
	appointmentCanceledByBarber            = "Уведомляем Вас об отмене записи.\nИнформация об отмененной записи:\n\n%s\n\nЗапись отменена барбером %s. Время отмененной записи: %s в %s."
	appointmentCanceledByBarberWithApology = `Ваша запись на %s в %s отменена в связи с непредвиденными обстоятельствами. Барбер %s приносит свои извинения.
Для уточнения подробностей Вы можете связаться с барбером лично.
%s
Вы также можете записаться повторно через наш бот на любое свободное время.
Информация об отмененной записи:
%s`

	barbersMemo = `Прежде чем клиенты получат возможность записаться к Вам на стрижку через этот бот, Вы должны произвести необходимый минимум подготовительных настроек.
Это необходимо для того, чтобы предоставить Вашим клиентам максимально комфортный пользовательский опыт обращения с этим ботом.
Итак, что необходимо сделать:

1. Через меню управления аккаунтом ознакомьтесь с политикой конфиденциальности и введите свое имя для отображения клиентам.
2. Через меню управления услугами создайте минимум одну услугу, которую вы будете предоставлять клиентам.

Также при необходимости связаться с Вами лично бот может предоставить клиенту ссылку на Ваш аккаунт telegram. Однако, при определенных настройках Вашего профиля эта ссылка может не работать. Чтобы убедиться, что ссылка на Ваш профиль будет работать выполните следующие действия:
1. Зайдите в настройки своего профиля telegram.
2. Выберите "Конфиденциальность".
3. Выберите "Пересылка сообщений".
4. Убедитесь, что в настройке "Кто может ссылаться на мой аккаунт при пересылке сообщений" выбрано "Все".`

	manageBarberAccount = "В этом меню собраны функции для управления Вашим аккаунтом."
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
Для выбора доступны только даты не раньше последней существующей записи клиента на стрижку.
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
- название услуги может содержать любые буквы (минимум 1 буква в названии), цифры, пробелы, знаки пунктуации, а также знаки + и -;
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
Подробная инструкция (для Android):
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

	endpntBarberSelectServiceForAppointment = "barber_select_service_for_appointment"
	endpntBarberSelectMonthForAppointment   = "barber_select_month_for_appointment"
	endpntBarberSelectWorkdayForAppointment = "barber_select_workday_for_appointment"
	endpntBarberSelectTimeForAppointment    = "barber_select_time_for_appointment"

	endpntSelectMonthFromScheduleCalendar   = "select_month_from_schedule_calendar"
	endpntSelectWorkdayFromScheduleCalendar = "select_workday_from_schedule_calendar"

	endpntSelectAppointment                        = "select_appointment"
	endpntAddСertainNonWorkdayFromScheduleCalendar = "add_certain_nonworkday_from_schedule_calendar"
	endpntSelectWorkdayStartTime                   = "select_workday_start_time"
	endpntSelectWorkdayEndTime                     = "select_workday_end_time"

	endpntBarberRescheduleAppointment = "barber_reschedule_appointment"
	endpntBarberCancelAppointment     = "barber_cancel_appointment"

	endpntSelectMonthOfLastWorkDate = "select_month_of_last_work_date"
	endpntSelectLastWorkDate        = "select_last_work_date"

	endpntEnterServiceDuration = "enter_service_duration"
	endpntServiceToEdit        = "service_to_edit"
	endpntEditServiceDuration  = "edit_service_duration"
	endpntServiceToDelete      = "service_to_delete"
	endpntSureToDeleteService  = "sure_to_delete_service"

	endpntBarberToDeletion = "barber_to_deletion"
	endpntBarberBackToMain = "barber_back_to_main"
)

var (
	markupBarberMain              = &tele.ReplyMarkup{}
	btnSignUpClientForAppointment = markupEmpty.Data("Записать клиента на стрижку", "sign_up_client_for_appointment")
	btnMyWorkSchedule             = markupEmpty.Data("Мой график работы", "my_work_schedule")
	btnBarberSettings             = markupEmpty.Data("Настройки", "barber_settings")

	markupBarberConfirmNewAppointment = &tele.ReplyMarkup{}
	btnBarberConfirmNewAppointment    = markupEmpty.Data("Подтвердить запись", "barber_confirm_new_appointment")

	markupBarberConfirmRescheduleAppointment = &tele.ReplyMarkup{}
	btnBarberConfirmRescheduleAppointment    = markupEmpty.Data("Подтвердить перенос записи", "barber_confirm_reschedule_appointment")

	markupBarberConfirmCancelAppointment = &tele.ReplyMarkup{}
	btnBarberConfirmCancelAppointment    = markupEmpty.Data("Подтвердить отмену записи", "barber_confirm_cancel_appointment")

	markupConfirmCancelAppointmentAndApology = &tele.ReplyMarkup{}
	btnConfirmCancelAppointmentAndApology    = markupEmpty.Data("Подтвердить отмену записи", "confirm_cancel_appointment_and_apology")

	markupBarberFailedToSaveOrRescheduleAppointment = &tele.ReplyMarkup{}
	btnBarberSelectAnotherTimeForAppointment        = markupEmpty.Data("Выбрать другое время", "barber_select_another_time_for_appointment")

	markupUpdNote = &tele.ReplyMarkup{}
	btnAddNote    = markupEmpty.Data("Добавить заметку", "add_note")

	btnAddWorkday    = markupEmpty.Data("Добавить рабочий день", "add_workday")
	btnAddNonWorkday = markupEmpty.Data("Сделать день выходным", "add_nonworkday")

	markupWorkdayIsFree      = &tele.ReplyMarkup{}
	btnMakeThisDayNonWorking = markupEmpty.Data("Сделать этот день выходным", endpntAddСertainNonWorkdayFromScheduleCalendar)
	btnUpdWorkdayStartTime   = markupEmpty.Data("Изменить начало рабочего дня", "upd_workday_start_time")
	btnUpdWorkdayEndTime     = markupEmpty.Data("Изменить конец рабочего дня", "upd_workday_end_time")
	btnBackToSelectWorkday   = markupEmpty.Data("Назад к выбору рабочего дня", endpntSelectMonthFromScheduleCalendar, "0")

	btnCancelAppointmentAndApology = markupEmpty.Data("Отменить и извиниться", "cancel_appointment_and_apology")
	btnUpdNote                     = markupEmpty.Data("Добавить/обновить заметку", "upd_note")

	markupBarberSettings   = &tele.ReplyMarkup{}
	btnBarbersMemo         = markupEmpty.Data("Памятка барбера", "barbers_memo")
	btnBarberManageAccount = markupEmpty.Data("Управление аккаунтом", "barber_manage_account")
	btnManageServices      = markupEmpty.Data("Управление услугами", "manage_services")
	btnManageBarbers       = markupEmpty.Data("Управление барберами", "manage_barbers")

	markupShortBarberSettings = &tele.ReplyMarkup{}

	markupBarberManageAccount = &tele.ReplyMarkup{}
	btnBarberCurrentSettings  = markupEmpty.Data("Мои текущие настройки", "barber_current_settings")
	btnBarberUpdPersonal      = markupEmpty.Data("Обновить персональные данные", "barber_upd_personal_data")
	btnDeleteAccount          = markupEmpty.Data("Удаление аккаунта барбера", "delete_account")

	markupBarberPrivacy       = &tele.ReplyMarkup{}
	btnBarberAgreeWithPrivacy = markupEmpty.Data("Соглашаюсь с политикой конфиденциальности", "barber_agree_with_privacy")

	markupBarberPersonal = &tele.ReplyMarkup{}
	btnBarberUpdName     = markupEmpty.Data("Обновить имя", "barber_upd_name")
	btnBarberUpdPhone    = markupEmpty.Data("Обновить номер телефона", "barber_upd_phone")

	markupDeleteAccount     = &tele.ReplyMarkup{}
	btnSetLastWorkDate      = markupEmpty.Data("Установить последний рабочий день", endpntSelectMonthOfLastWorkDate, "0")
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

	markupReadyToCreateService = &tele.ReplyMarkup{}
	btnSaveNewService          = markupEmpty.Data("Сохранить новую услугу", "save_new_service")

	markupEnterServiceName = &tele.ReplyMarkup{}

	markupContinueEditingOrSelectService = &tele.ReplyMarkup{}
	btnСontinueEditingService            = markupEmpty.Data("Продолжить ранее начатое", "continue_editing")
	btnSelectServiceToEdit               = markupEmpty.Data("Начать заново", "select_service_to_edit")

	markupEditServiceParams        = &tele.ReplyMarkup{}
	btnEditServiceName             = markupEmpty.Data("Изменить название услуги", "edit_service_name")
	btnEditServiceDescription      = markupEmpty.Data("Изменить описание услуги", "edit_service_description")
	btnEditServicePrice            = markupEmpty.Data("Изменить цену услуги", "edit_service_price")
	btnSelectServiceDurationOnEdit = markupEmpty.Data("Изменить продолжительность услуги", "select_service_duration_on_edit")

	markupReadyToUpdateService = &tele.ReplyMarkup{}
	btnUpdateService           = markupEmpty.Data("Применить изменения", "update_service")

	markupEditServiceName = &tele.ReplyMarkup{}

	markupManageBarbers = &tele.ReplyMarkup{}
	btnShowAllBurbers   = markupEmpty.Data("Список барберов", "show_all_barbers")
	btnAddBarber        = markupEmpty.Data("Добавить барбера", "add_barber")
	btnDeleteBarber     = markupEmpty.Data("Удалить барбера", "delete_barber")

	markupBarberBackToMain = &tele.ReplyMarkup{}
	btnBarberBackToMain    = markupEmpty.Data(backToMain, endpntBarberBackToMain)
)

func init() {
	markupBarberMain.Inline(
		markupEmpty.Row(btnSignUpClientForAppointment),
		markupEmpty.Row(btnMyWorkSchedule),
		markupEmpty.Row(btnBarberSettings),
	)

	markupBarberConfirmNewAppointment.Inline(
		markupEmpty.Row(btnBarberConfirmNewAppointment),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberFailedToSaveOrRescheduleAppointment.Inline(
		markupEmpty.Row(btnBarberSelectAnotherTimeForAppointment),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberConfirmRescheduleAppointment.Inline(
		markupEmpty.Row(btnBarberConfirmRescheduleAppointment),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberConfirmCancelAppointment.Inline(
		markupEmpty.Row(btnBarberConfirmCancelAppointment),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupConfirmCancelAppointmentAndApology.Inline(
		markupEmpty.Row(btnConfirmCancelAppointmentAndApology),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupUpdNote.Inline(
		markupEmpty.Row(btnAddNote),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupWorkdayIsFree.Inline(
		markupEmpty.Row(btnMakeThisDayNonWorking),
		markupEmpty.Row(btnUpdWorkdayStartTime),
		markupEmpty.Row(btnUpdWorkdayEndTime),
		markupEmpty.Row(btnBackToSelectWorkday),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberSettings.Inline(
		markupEmpty.Row(btnBarbersMemo),
		markupEmpty.Row(btnBarberManageAccount),
		markupEmpty.Row(btnManageServices),
		markupEmpty.Row(btnManageBarbers),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupShortBarberSettings.Inline(
		markupEmpty.Row(btnBarberManageAccount),
		markupEmpty.Row(btnManageServices),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberManageAccount.Inline(
		markupEmpty.Row(btnBarberCurrentSettings),
		markupEmpty.Row(btnBarberUpdPersonal),
		markupEmpty.Row(btnDeleteAccount),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberPrivacy.Inline(
		markupEmpty.Row(btnBarberAgreeWithPrivacy),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberPersonal.Inline(
		markupEmpty.Row(btnBarberUpdName),
		markupEmpty.Row(btnBarberUpdPhone),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupDeleteAccount.Inline(
		markupEmpty.Row(btnSetLastWorkDate),
		markupEmpty.Row(btnSelfDeleteBarber),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupConfirmSelfDeletion.Inline(
		markupEmpty.Row(btnSureToSelfDeleteBarber),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupManageServicesFull.Inline(
		markupEmpty.Row(btnShowMyServices),
		markupEmpty.Row(btnCreateService),
		markupEmpty.Row(btnEditService),
		markupEmpty.Row(btnDeleteService),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupManageServicesShort.Inline(
		markupEmpty.Row(btnCreateService),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupShowMyServices.Inline(
		markupEmpty.Row(btnCreateService),
		markupEmpty.Row(btnEditService),
		markupEmpty.Row(btnDeleteService),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupСontinueOldOrMakeNewService.Inline(
		markupEmpty.Row(btnСontinueOldService),
		markupEmpty.Row(btnMakeNewService),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupEnterServiceParams.Inline(
		markupEmpty.Row(btnEnterServiceName),
		markupEmpty.Row(btnEnterServiceDescription),
		markupEmpty.Row(btnEnterServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEnter),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupReadyToCreateService.Inline(
		markupEmpty.Row(btnSaveNewService),
		markupEmpty.Row(btnEnterServiceName),
		markupEmpty.Row(btnEnterServiceDescription),
		markupEmpty.Row(btnEnterServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEnter),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupEnterServiceName.Inline(
		markupEmpty.Row(btnEnterServiceName),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupContinueEditingOrSelectService.Inline(
		markupEmpty.Row(btnСontinueEditingService),
		markupEmpty.Row(btnSelectServiceToEdit),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupEditServiceParams.Inline(
		markupEmpty.Row(btnEditServiceName),
		markupEmpty.Row(btnEditServiceDescription),
		markupEmpty.Row(btnEditServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEdit),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupReadyToUpdateService.Inline(
		markupEmpty.Row(btnUpdateService),
		markupEmpty.Row(btnEditServiceName),
		markupEmpty.Row(btnEditServiceDescription),
		markupEmpty.Row(btnEditServicePrice),
		markupEmpty.Row(btnSelectServiceDurationOnEdit),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupEditServiceName.Inline(
		markupEmpty.Row(btnEditServiceName),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupManageBarbers.Inline(
		markupEmpty.Row(btnShowAllBurbers),
		markupEmpty.Row(btnAddBarber),
		markupEmpty.Row(btnDeleteBarber),
		markupEmpty.Row(btnBarberBackToMain),
	)

	markupBarberBackToMain.Inline(
		markupEmpty.Row(btnBarberBackToMain),
	)
}

func appointmentHashStr(appt ent.Appointment) string {
	bytes := append(intToBytes(appt.ID), intToBytes(appt.WorkdayID)...)
	bytes = append(bytes, appt.Time.Bytes()...)
	bytes = append(bytes, appt.Duration.Bytes()...)
	return fmt.Sprintf("%x", md5.Sum(bytes))
}

func btnAppointment(appt ent.Appointment) tele.Btn {
	return markupEmpty.Data(
		appt.Time.ShortString()+" - "+(appt.Time+appt.Duration).ShortString(),
		endpntSelectAppointment,
		strings.Join(
			[]string{
				strconv.Itoa(appt.ID),
				appointmentHashStr(appt),
			},
			"|",
		),
	)
}

func btnDate(date time.Time, endpnt string) tele.Btn {
	return markupEmpty.Data(strconv.Itoa(date.Day()), endpnt, date.Format(time.DateOnly))
}

func btnDuration(dur tm.Duration, endpnt string) tele.Btn {
	return markupEmpty.Data(dur.LongString(), endpnt, strconv.FormatUint(uint64(dur), 10))
}

func intToBytes(num int) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(num))
	return bytes
}

func markupBackToWorkdayInfo(workdayID int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data(
			"Назад к рабочему дню",
			endpntSelectWorkdayFromScheduleCalendar,
			strconv.Itoa(workdayID)),
		),
		markup.Row(btnBarberBackToMain),
	)
	return markup
}

func markupConfirmServiceDeletion(serviceID int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("Подтвердить удаление", endpntSureToDeleteService, strconv.Itoa(serviceID))),
		markup.Row(btnBarberBackToMain),
	)
	return markup
}

func markupEditAppointment(workdayID int, userID int64) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	if userID != 0 {
		rows = append(
			rows,
			markup.Row(markup.Data("Перенести и уведомить", endpntBarberRescheduleAppointment)),
			markup.Row(markup.Data("Отменить и уведомить", endpntBarberCancelAppointment)),
			markup.Row(btnCancelAppointmentAndApology),
		)
	} else {
		rows = append(
			rows,
			markup.Row(markup.Data("Перенести", endpntBarberRescheduleAppointment)),
			markup.Row(markup.Data("Отменить", endpntBarberCancelAppointment)),
		)
	}
	rows = append(
		rows,
		markup.Row(btnUpdNote),
		markup.Row(markup.Data(
			"Назад к рабочему дню",
			endpntSelectWorkdayFromScheduleCalendar,
			strconv.Itoa(workdayID)),
		),
		markup.Row(btnBarberBackToMain),
	)
	markup.Inline(rows...)
	return markup
}

func markupSelectAppointment(appointments []ent.Appointment) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	rows = append(rows, rowsSelectAppointment(appointments)...)
	rows = append(
		rows,
		markup.Row(btnUpdWorkdayStartTime),
		markup.Row(btnUpdWorkdayEndTime),
		markup.Row(btnBackToSelectWorkday),
		markup.Row(btnBarberBackToMain),
	)
	markup.Inline(rows...)
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
	rows = append(rows, markup.Row(btnBarberBackToMain))
	markup.Inline(rows...)
	return markup
}

func markupSelectLastWorkDate(dateRange ent.DateRange, monthRange monthRange) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	btnPrevMonth, btnNextMonth := btnsSwitchMonth(tm.ParseMonth(dateRange.LastDate), monthRange, endpntSelectMonthOfLastWorkDate)
	rowSelectMonth := markup.Row(btnPrevMonth, btnMonth(dateRange.Month()), btnNextMonth)
	rowsSelectDate := rowsSelectLastWorkDate(dateRange)
	rowRestoreDefaultDate := markup.Row(btnInfiniteLastWorkDate)
	rowBackToMainBarber := markup.Row(btnBarberBackToMain)
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
		btnsDurationsToSelect = append(btnsDurationsToSelect, btnDuration(duration, endpnt))
	}
	rows := markup.Split(2, btnsDurationsToSelect)
	rows = append(rows, markup.Row(btnBarberBackToMain))
	markup.Inline(rows...)
	return markup
}

func markupSelectWorkdayFromScheduleCalendar(dateRange ent.DateRange, monthRange monthRange, barberID int64) (*tele.ReplyMarkup, error) {
	markup := &tele.ReplyMarkup{}
	btnPrevMonth, btnNextMonth := btnsSwitchMonth(
		tm.ParseMonth(dateRange.LastDate),
		monthRange,
		endpntSelectMonthFromScheduleCalendar,
	)
	rowSelectMonth := markup.Row(btnPrevMonth, btnMonth(dateRange.Month()), btnNextMonth)
	rowsSelectWorkday, err := rowsSelectWorkdayFromScheduleCalendar(dateRange, barberID)
	if err != nil {
		return nil, err
	}
	var rows []tele.Row
	rows = append(rows, rowSelectMonth, rowWeekdays)
	rows = append(rows, rowsSelectWorkday...)
	rows = append(rows, markup.Row(btnAddWorkday), markup.Row(btnAddNonWorkday), markup.Row(btnBarberBackToMain))
	markup.Inline(rows...)
	return markup, nil
}

func markupSelectWorkdayStartOrEndTime(earlestTime, latestTime tm.Duration, endpntTime string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	rows = append(rows, rowsSelectStartOrEndTime(earlestTime, latestTime, endpntTime)...)
	rows = append(rows, markup.Row(btnBackToSelectWorkday), markup.Row(btnBarberBackToMain))
	markup.Inline(rows...)
	return markup
}

func rowsSelectAppointment(appointments []ent.Appointment) []tele.Row {
	var btnsAppointmentsToSelect []tele.Btn
	for _, appointment := range appointments {
		btnsAppointmentsToSelect = append(btnsAppointmentsToSelect, btnAppointment(appointment))
	}
	for i := 1; i <= (3-len(appointments)%3)%3; i++ {
		btnsAppointmentsToSelect = append(btnsAppointmentsToSelect, btnEmpty)
	}
	return markupEmpty.Split(3, btnsAppointmentsToSelect)
}

func rowsSelectLastWorkDate(dateRange ent.DateRange) []tele.Row {
	var btnsDatesToSelect []tele.Btn
	for i := 1; i < dateRange.StartWeekday(); i++ {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	for date := dateRange.FirstDate; date.Compare(dateRange.LastDate) <= 0; date = date.Add(24 * time.Hour) {
		btnsDatesToSelect = append(btnsDatesToSelect, btnDate(date, endpntSelectLastWorkDate))
	}
	for i := 7; i > dateRange.EndWeekday(); i-- {
		btnsDatesToSelect = append(btnsDatesToSelect, btnEmpty)
	}
	return markupEmpty.Split(7, btnsDatesToSelect)
}

func rowsSelectStartOrEndTime(earlestTime, latestTime tm.Duration, endpntTime string) []tele.Row {
	var btnsTimesToSelect []tele.Btn
	timesNumber := 0
	for time := earlestTime; time <= latestTime; time += 30 * tm.Minute {
		btnsTimesToSelect = append(btnsTimesToSelect, btnTime(time, endpntTime))
		timesNumber++
	}
	for i := 1; i <= (4-timesNumber%4)%4; i++ {
		btnsTimesToSelect = append(btnsTimesToSelect, btnEmpty)
	}
	return markupEmpty.Split(4, btnsTimesToSelect)
}

func rowsSelectWorkdayFromScheduleCalendar(dateRange ent.DateRange, barberID int64) ([]tele.Row, error) {
	wds, err := cp.RepoWithContext.GetWorkdaysByDateRange(barberID, dateRange)
	if err != nil {
		return nil, err
	}
	workdays := make(map[int]ent.Workday)
	for _, wd := range wds {
		workdays[wd.Date.Day()] = wd
	}
	var btnsWorkdaysToSelect []tele.Btn
	for i := 1; i < dateRange.StartWeekday(); i++ {
		btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
	}
	for date := dateRange.FirstDate; date.Compare(dateRange.LastDate) <= 0; date = date.Add(24 * time.Hour) {
		workday, ok := workdays[date.Day()]
		if !ok {
			btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnDash)
		} else {
			btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnWorkday(workday, endpntSelectWorkdayFromScheduleCalendar))
		}
	}
	for i := 7; i > dateRange.EndWeekday(); i-- {
		btnsWorkdaysToSelect = append(btnsWorkdaysToSelect, btnEmpty)
	}
	return markupEmpty.Split(7, btnsWorkdaysToSelect), nil
}
