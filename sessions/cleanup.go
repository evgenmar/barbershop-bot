package sessions

func CleanupSessions() {
	getBarberSessionManager().cleanupBarberSessions()
}
