package request

import (
	"bytes"
	"crypto/md5"
	"fmt"
)

type Request struct {
	URL    string
	Method string
	Retry  int
	Data   map[string]string
}

func (r *Request) Hash(hashFields ...string) string {
	components := make([][]byte, 2+len(hashFields))
	components[0] = []byte(r.URL)
	components[1] = []byte(r.Method)

	for i, field := range hashFields {
		if fieldValue, ok := r.Data[field]; ok {
			components[i+2] = []byte(fieldValue)
		} else {
			panic(fmt.Errorf("field not in request.Data"))
		}

	}

	hash := md5.Sum(bytes.Join(components, []byte(":")))
	return string(hash[:])
}
