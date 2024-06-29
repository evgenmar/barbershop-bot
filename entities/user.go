package entities

type User struct {
	ID   int64
	Name string

	//Format of phone number is +71234567890.
	Phone string
}
