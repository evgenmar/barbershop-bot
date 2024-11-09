package main

import (
	"github.com/evgenmar/barbershop-bot/initialization"
	sched "github.com/evgenmar/barbershop-bot/scheduler"
	tg "github.com/evgenmar/barbershop-bot/telegram"
)

func main() {
	initialization.InitGlobals()
	sched.Start()
	tg.Bot.Start()
}
