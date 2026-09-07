package validators

import (
	"regexp"
	"strings"
)

var (
	urlRegex = regexp.MustCompile(`(?i)\b((?:[a-z][a-z0-9+.-]*://|www\.)[^\s<>"']+|(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}(?::\d{2,5})?(?:[/?#][^\s<>"']*)?)\b`)
)

func FindURL(text string) (string, bool) {
	if text == "" {
		return "", false
	}

	lowerText := strings.ToLower(text)

	// Проверяем только ссылки с протоколами или www
	if match := urlRegex.FindString(lowerText); match != "" {
		return match, true
	}

	return "", false
}
