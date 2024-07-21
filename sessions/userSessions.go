package sessions

import (
	"sync"
	"time"
)

type userSession struct {
	status
	Appointment
	expiresAt int64
}

type userSessionManager struct {
	sessions map[int64]userSession
	mutex    sync.RWMutex
}

var (
	userSessManager *userSessionManager
	onceUser        sync.Once
)

func GetAppointmentUser(userID int64) Appointment {
	session := getUserSessionManager().getSession(userID)
	return session.Appointment
}

func GetUserState(userID int64) State {
	session := getUserSessionManager().getSession(userID)
	if !session.status.isValid() {
		return StateStart
	}
	return session.state
}

func UpdateAppointmentAndUserState(userID int64, appointment Appointment, state State) {
	session := getUserSessionManager().getSession(userID)
	session.Appointment = appointment
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
