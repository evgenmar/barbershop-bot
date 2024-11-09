package sessions

import (
	"fmt"
	"sync"
	"time"

	ent "github.com/evgenmar/barbershop-bot/entities"
	tm "github.com/evgenmar/barbershop-bot/lib/time"
)

type barberSession struct {
	status
	NewService
	EditedService
	LastWorkDate
	Appointment
	expiresAt int64
}

type barberSessionManager struct {
	sessions map[int64]barberSession
	mutex    sync.RWMutex
}

type EditedService struct {
	ID         int
	OldService Service
	UpdService Service
}

type LastWorkDate struct {
	LastShownMonth tm.Month
}

type NewService struct {
	Service
}

type Service struct {
	Name       string
	Desciption string
	Price      ent.Price
	Duration   tm.Duration
}

var (
	barberSessManager *barberSessionManager
	onceBarber        sync.Once
)

func (s NewService) Info() string {
	name := "не указано"
	description := "не указано"
	price := "не указана"
	duration := "не указана"
	if s.Name != "" {
		name = fmt.Sprintf("*%s*", s.Name)
	}
	if s.Desciption != "" {
		description = fmt.Sprintf("*%s*", s.Desciption)
	}
	if s.Price != 0 {
		price = fmt.Sprintf("*%s*", s.Price.String())
	}
	if s.Duration != 0 {
		duration = fmt.Sprintf("*%s*", s.Duration.LongString())
	}
	format := "На данный момент создаваемая услуга имеет вид:\nНазвание услуги: %s\nОписание услуги: %s\nЦена услуги: %s\nПродолжительность услуги: %s"
	return fmt.Sprintf(format, name, description, price, duration)
}

func (s EditedService) Info() string {
	serviceToDisplay := ent.Service{
		Name:       s.OldService.Name,
		Desciption: s.OldService.Desciption,
		Price:      s.OldService.Price,
		Duration:   s.OldService.Duration,
	}
	if s.UpdService.Name == "" && s.UpdService.Desciption == "" && s.UpdService.Price == 0 && s.UpdService.Duration == 0 {
		return "Выбранная для редактирования услуга имеет вид:\n\n" + serviceToDisplay.Info()
	}
	if s.UpdService.Name != "" {
		serviceToDisplay.Name = s.UpdService.Name
	}
	if s.UpdService.Desciption != "" {
		serviceToDisplay.Desciption = s.UpdService.Desciption
	}
	if s.UpdService.Price != 0 {
		serviceToDisplay.Price = s.UpdService.Price
	}
	if s.UpdService.Duration != 0 {
		serviceToDisplay.Duration = s.UpdService.Duration
	}
	return "Редактируемая услуга с учетом внесенных изменений будет иметь вид:\n\n" + serviceToDisplay.Info()
}

func GetAppointmentBarber(barberID int64) Appointment {
	session := getBarberSessionManager().getSession(barberID)
	return session.Appointment
}

func GetBarberState(barberID int64) State {
	session := getBarberSessionManager().getSession(barberID)
	if !session.status.isValid() {
		return StateStart
	}
	return session.state
}

func GetEditedService(barberID int64) EditedService {
	session := getBarberSessionManager().getSession(barberID)
	return session.EditedService
}

func GetLastWorkDate(barberID int64) LastWorkDate {
	session := getBarberSessionManager().getSession(barberID)
	return session.LastWorkDate
}

func GetNewService(barberID int64) NewService {
	session := getBarberSessionManager().getSession(barberID)
	return session.NewService
}

func UpdateAppointmentAndBarberState(barberID int64, appointment Appointment, state State) {
	session := getBarberSessionManager().getSession(barberID)
	session.Appointment = appointment
	session.status = newStatus(state)
	getBarberSessionManager().updateSession(barberID, session)
}

func UpdateBarberState(barberID int64, state State) {
	session := getBarberSessionManager().getSession(barberID)
	session.status = newStatus(state)
	getBarberSessionManager().updateSession(barberID, session)
}

func UpdateEditedServiceAndState(barberID int64, editedService EditedService, state State) {
	session := getBarberSessionManager().getSession(barberID)
	session.EditedService = editedService
	session.status = newStatus(state)
	getBarberSessionManager().updateSession(barberID, session)
}

func UpdateLastWorkDateAndState(barberID int64, lastWorkDate LastWorkDate, state State) {
	session := getBarberSessionManager().getSession(barberID)
	session.LastWorkDate = lastWorkDate
	session.status = newStatus(state)
	getBarberSessionManager().updateSession(barberID, session)
}

func UpdateNewServiceAndState(barberID int64, newService NewService, state State) {
	session := getBarberSessionManager().getSession(barberID)
	session.NewService = newService
	session.status = newStatus(state)
	getBarberSessionManager().updateSession(barberID, session)
}

func getBarberSessionManager() *barberSessionManager {
	onceBarber.Do(func() {
		barberSessManager = &barberSessionManager{
			sessions: make(map[int64]barberSession),
		}
	})
	return barberSessManager
}

func (m *barberSessionManager) getSession(barberID int64) barberSession {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	session, ok := m.sessions[barberID]
	if !ok {
		return barberSession{}
	}
	return session
}

func (m *barberSessionManager) updateSession(barberID int64, session barberSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	session.expiresAt = time.Now().Add(time.Hour * 72).Unix()
	m.sessions[barberID] = session
}

func (m *barberSessionManager) cleanupBarberSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	now := time.Now().Unix()
	for barberID, session := range m.sessions {
		if session.expiresAt < now {
			delete(m.sessions, barberID)
		}
	}
}
