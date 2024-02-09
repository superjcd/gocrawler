package cookie

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/go-redis/redis"
)

type CoookieGetter interface {
	Get() (*cookiejar.Jar, error)
}

type RedisCookieGetter struct {
	CookieUrl *url.URL
	RCli      *redis.Client
	Key       string
}

func (c *RedisCookieGetter) Get() (*cookiejar.Jar, error) {
	jar, _ := cookiejar.New(nil)
	cookieStr, err := c.RCli.SRandMember(c.Key).Result()
	if err != nil {

		return nil, err
	}

	var cookieData map[string]string
	json.Unmarshal([]byte(cookieStr), &cookieData)

	cookies := make([]*http.Cookie, 0)

	for k, v := range cookieData {
		cookie := &http.Cookie{Name: k, Value: v}
		cookies = append(cookies, cookie)
	}
	jar.SetCookies(c.CookieUrl, cookies)
	return jar, nil
}

func NewRedisCookieGetter(url *url.URL, cli *redis.Client, key string) (*RedisCookieGetter, error) {
	cg := &RedisCookieGetter{CookieUrl: url, RCli: cli, Key: key}

	return cg, nil
}
