package sessions

import (
	"sync"
	"time"
)

type userSession struct {
	status
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

func GetUserState(id int64) State {
	session := getUserSessionManager().getSession(id)
	if !session.status.isValid() {
		return StateStart
	}
	return session.state
}

func UpdateUserState(id int64, state State) {
	session := getUserSessionManager().getSession(id)
	session.status = newStatus(state)
	getUserSessionManager().updateSession(id, session)
}

func getUserSessionManager() *userSessionManager {
	onceUser.Do(func() {
		userSessManager = &userSessionManager{
			sessions: make(map[int64]userSession),
		}
	})
	return userSessManager
}

func (m *userSessionManager) getSession(id int64) userSession {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	session, ok := m.sessions[id]
	if !ok {
		return userSession{}
	}
	return session
}

func (m *userSessionManager) updateSession(id int64, session userSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	session.expiresAt = time.Now().Add(time.Hour * 72).Unix()
	m.sessions[id] = session
}

func (m *userSessionManager) cleanupUserSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	now := time.Now().Unix()
	for id, session := range m.sessions {
		if session.expiresAt < now {
			delete(m.sessions, id)
		}
	}
}
