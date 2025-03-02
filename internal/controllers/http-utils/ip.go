package httputils

import (
	"net/http"
	"strings"
)

func GetRequestIP(r *http.Request) string {
	split := strings.Split(r.RemoteAddr, ":")
	return strings.Join(split[:len(split)-1], ":")
}
