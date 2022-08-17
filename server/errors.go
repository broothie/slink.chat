package server

import (
	"github.com/broothie/slink.chat/util"
	"github.com/samber/lo"
)

func errorMap(errors ...error) util.Map {
	return util.Map{"errors": lo.Map(errors, func(err error, _ int) string { return err.Error() })}
}
