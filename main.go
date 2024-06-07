package main

import (
	"barbershop-bot/initialization"
	sched "barbershop-bot/scheduler"
	tg "barbershop-bot/telegram"
)

func main() {
	initialization.Globals()
	sched.Cron.Start()
	tg.Bot.Start()
}
