package core

type Migration struct {
	ID   string
	Up   func(tx DBTx) error
	Down func(tx DBTx) error
}
