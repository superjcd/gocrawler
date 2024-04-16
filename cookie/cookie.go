package cookie

import (
	"context"
	"net/http/cookiejar"
)

type CookieGetter interface {
	Get(context.Context) (*cookiejar.Jar, error)
}
