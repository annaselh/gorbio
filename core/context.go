package core

type Context struct {
}

type MenuItem struct {
	Name     string
	Icon     string
	Children []MenuItem
}
