package models

type TeamPlayer struct {
	name    string
	players []Player
}

type Player struct {
	FirstName string
	LastName  string
	FullName  string
	Age       int64
}
