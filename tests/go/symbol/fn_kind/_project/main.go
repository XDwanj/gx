package demo

const Limit = 10

type User struct{}

type Store interface {
	Load() User
}

type Status int

func BuildUser() User {
	return User{}
}

func (User) Load() User {
	return User{}
}
