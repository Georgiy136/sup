package utils

import "strings"

var replacer = strings.NewReplacer(
	"_", `\_`,
	"[", `\[`,
	"]", `]`,
	"`", "",
	"\\", `\`,
	"*", `\*`,
)

func ReplaceSpecSymbols(input string) string {
	return replacer.Replace(input)
}
