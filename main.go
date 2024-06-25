package main

import (
	"barbershop-bot/initialization"
	sched "barbershop-bot/scheduler"
	tg "barbershop-bot/telegram"
)

func main() {
	initialization.InitGlobals()
	sched.Start()
	tg.Bot.Start()
}
