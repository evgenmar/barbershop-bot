package sessions

import (
	ent "barbershop-bot/entities"
	tm "barbershop-bot/lib/time"
	"fmt"
	"sync"
	"time"
)

type BarberSession struct {
	status
	NewService
	expiresAt int64
}

type BarberSessionManager struct {
	sessions map[int64]BarberSession
	mutex    sync.RWMutex
}

type NewService struct {
	Name       string
	Desciption string
	Price      ent.Price
	Duration   tm.Duration
}

var (
	barberSessionManager *BarberSessionManager
	once                 sync.Once
)

func getBarberSessionManager() *BarberSessionManager {
	once.Do(func() {
		barberSessionManager = &BarberSessionManager{
			sessions: make(map[int64]BarberSession),
		}
	})
	return barberSessionManager
}

func (m *BarberSessionManager) get(id int64) BarberSession {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	session, ok := m.sessions[id]
	if !ok {
		return BarberSession{}
	}
	return session
}

func (m *BarberSessionManager) update(id int64, session BarberSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	session.expiresAt = time.Now().Add(time.Hour * 72).Unix()
	m.sessions[id] = session
}

func (m *BarberSessionManager) cleanupBarberSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	now := time.Now().Unix()
	for id, session := range m.sessions {
		if session.expiresAt < now {
			delete(m.sessions, id)
		}
	}
}

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
	format := "На данный момент создаваемая услуга имеет вид:\nНазвание услуги: %s\nОписание услуги: %s\nЦена услуги: %s\nПродолжительность услуги: %s\n"
	return fmt.Sprintf(format, name, description, price, duration)
}

func getBarberSession(id int64) BarberSession {
	return getBarberSessionManager().get(id)
}

func GetBarberState(id int64) State {
	session := getBarberSession(id)
	if !session.status.isValid() {
		session.status = newStatus(StateStart)
		updateBarberSession(id, session)
		return StateStart
	}
	return session.state
}

func GetNewService(id int64) NewService {
	session := getBarberSession(id)
	return session.NewService
}

func updateBarberSession(id int64, session BarberSession) {
	getBarberSessionManager().update(id, session)
}

func UpdateBarberState(id int64, state State) {
	session := getBarberSession(id)
	session.status = newStatus(state)
	updateBarberSession(id, session)
}

func UpdateNewServiceAndState(id int64, nservice NewService, state State) {
	session := getBarberSession(id)
	session.NewService = nservice
	session.status = newStatus(state)
	updateBarberSession(id, session)
}
