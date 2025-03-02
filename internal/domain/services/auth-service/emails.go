package authservice

import (
	"auth/internal/domain"
	"fmt"
	"time"
)

func newEmailContentRefreshFromNewIPWarning(newIP string) domain.EmailContent {
	timeStr := time.Now().In(time.UTC).Format(time.DateTime) + " (UTC)"
	bodyFormat := "Обнаружен вход в ваш аккаунт с нового ip-адреса.\n" +
		"Время: %s\n" +
		"IP-адрес: %s"
	return domain.EmailContent{
		Subject:     "Вход в аккаунт с нового IP",
		ContentType: "text/plain",
		Body:        fmt.Sprintf(bodyFormat, timeStr, newIP),
	}

}
