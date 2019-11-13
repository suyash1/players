package models

type TeamPlayer struct {
	name    string
	players []Player
}

type Player struct {
	FullName string
	Age      int64
	TeamName string
}
