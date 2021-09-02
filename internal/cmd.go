package internal

import "github.com/iancoleman/strcase"

func Commandize(name string) string {
	return strcase.ToKebab(name)
}
