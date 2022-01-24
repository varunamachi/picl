package cmn

import "strings"

func EqFold(str string, targets ...string) bool {
	for _, s := range targets {
		if strings.EqualFold(str, s) {
			return true
		}
	}
	return false
}
