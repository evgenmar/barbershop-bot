package entities

import (
	"strconv"

	tele "gopkg.in/telebot.v3"
)

type User struct {
	ID   int64
	Name string

	//Format of phone number is +71234567890.
	Phone string
	tele.StoredMessage
}

func (u User) Recipient() string {
	return strconv.FormatInt(u.StoredMessage.ChatID, 10)
}
