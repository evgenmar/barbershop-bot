package sessions

func CleanupSessions() {
	getBarberSessionManager().cleanupBarberSessions()
	getUserSessionManager().cleanupUserSessions()
}
