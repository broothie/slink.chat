package model

type Type string

func (t Type) Type() Type {
	return t
}

func (t Type) String() string {
	return string(t)
}

type Typer interface {
	Type() Type
}
