package cookie

import (
	"net/http/cookiejar"
)

type CoookieGetter interface {
	Get() (*cookiejar.Jar, error)
}
