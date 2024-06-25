package sessions

import (
	"sync"
)

type BarberSession struct {
	Status
}

type BarberSessionManager struct {
	sessions map[int64]BarberSession
	mutex    sync.RWMutex
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
		return BarberSession{Status: StatusStart}
	}
	return session
}

// Update updates only non-niladic fields of session. Niladic fields remains unchanged.
func (m *BarberSessionManager) update(id int64, session BarberSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	existingSession, ok := m.sessions[id]
	if !ok {
		existingSession = BarberSession{Status: StatusStart}
	}
	if session.State != 0 {
		existingSession.Status = session.Status
	}
	m.sessions[id] = existingSession
}

// TODO: make it unexported
func GetBarberSession(id int64) BarberSession {
	return getBarberSessionManager().get(id)
}

func GetBarberState(id int64) State {
	session := GetBarberSession(id)
	if !session.Status.isValid() {
		UpdateBarberSession(id, BarberSession{Status: StatusStart})
		return StateStart
	}
	return session.State
}

// TODO: make it unexported
// Update updates only non-niladic fields of session. Niladic fields remains unchanged.
func UpdateBarberSession(id int64, session BarberSession) {
	getBarberSessionManager().update(id, session)
}

func UpdateBarberState(id int64, state State) {
	UpdateBarberSession(id, BarberSession{Status: NewStatus(state)})
}
