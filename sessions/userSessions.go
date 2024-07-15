package sessions

import (
	tm "barbershop-bot/lib/time"
	"sync"
	"time"
)

type NewAppointment struct {
	WorkdayID     int
	ServiceID     int
	Time          tm.Duration
	Duration      tm.Duration
	BarberID      int64
	LastShownDate time.Time
}

type userSession struct {
	status
	newAppointment NewAppointment
	expiresAt      int64
}

type userSessionManager struct {
	sessions map[int64]userSession
	mutex    sync.RWMutex
}

var (
	userSessManager *userSessionManager
	onceUser        sync.Once
)

func GetNewAppointment(userID int64) NewAppointment {
	session := getUserSessionManager().getSession(userID)
	return session.newAppointment
}

func GetUserState(userID int64) State {
	session := getUserSessionManager().getSession(userID)
	if !session.status.isValid() {
		return StateStart
	}
	return session.state
}

func UpdateNewAppointmentAndState(userID int64, newAppointment NewAppointment, state State) {
	session := getUserSessionManager().getSession(userID)
	session.newAppointment = newAppointment
	session.status = newStatus(state)
	getUserSessionManager().updateSession(userID, session)
}

func UpdateUserState(userID int64, state State) {
	session := getUserSessionManager().getSession(userID)
	session.status = newStatus(state)
	getUserSessionManager().updateSession(userID, session)
}

func getUserSessionManager() *userSessionManager {
	onceUser.Do(func() {
		userSessManager = &userSessionManager{
			sessions: make(map[int64]userSession),
		}
	})
	return userSessManager
}

func (m *userSessionManager) getSession(userID int64) userSession {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	session, ok := m.sessions[userID]
	if !ok {
		return userSession{}
	}
	return session
}

func (m *userSessionManager) updateSession(userID int64, session userSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	session.expiresAt = time.Now().Add(time.Hour * 72).Unix()
	m.sessions[userID] = session
}

func (m *userSessionManager) cleanupUserSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	now := time.Now().Unix()
	for userID, session := range m.sessions {
		if session.expiresAt < now {
			delete(m.sessions, userID)
		}
	}
}
