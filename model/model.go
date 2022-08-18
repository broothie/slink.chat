package model

import (
	"fmt"
	"strings"

	"github.com/rs/xid"
)

type Type string

func (t Type) NewID() string {
	return fmt.Sprintf("%s_%s", strings.ToLower(string(t[0])), xid.New().String())
}

type Model interface {
	ModelType() Type
}
