package entity

import (
	"strings"
	"time"
	"unicode"
)

type User struct {
	ID               string
	Username         string
	Password         string // hashed (only for registered users)
	CreatedAt        time.Time
	IsGuest          bool
	PushSubscription *PushSubscription
}

func UsernameNormalize(username string) string {
	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" {
		return ""
	}

	translit := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
		'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i",
		'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
		'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
		'у': "u", 'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch",
		'ш': "sh", 'щ': "sch", 'ъ': "", 'ы': "y", 'ь': "",
		'э': "e", 'ю': "yu", 'я': "ya",
	}

	var b strings.Builder

	for _, r := range username {
		if v, ok := translit[r]; ok {
			b.WriteString(v)
			continue
		}

		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}

		// всё остальное — игнорируем (спецсимволы, emoji и т.д.)
		if unicode.IsSpace(r) {
			continue
		}
	}

	return b.String()
}
