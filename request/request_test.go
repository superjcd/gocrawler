package request

import (
	"testing"
)

func TestRequestHashWithSamePayload(t *testing.T) {
	var moreInfos = map[string]string{
		"taskid": "123",
	}
	req := Request{
		URL:    "http://example.com",
		Method: "GET",
		Data:   moreInfos,
	}

	hashed1 := req.Hash("taskid")

	req2 := Request{
		URL:    "http://example.com",
		Method: "GET",
		Data:   moreInfos,
	}
	hashed2 := req2.Hash("taskid")

	if hashed1 != hashed2 {
		t.Errorf("shoud be equal")
	}
}

func TestRequestHashWithDifferentPayload(t *testing.T) {
	var moreInfos = map[string]string{
		"taskid": "123",
	}
	// different taskid
	var moreInfos2 = map[string]string{
		"taskid": "456",
	}
	req := Request{
		URL:    "http://example.com",
		Method: "GET",
		Data:   moreInfos,
	}

	hashed1 := req.Hash("taskid")

	req2 := Request{
		URL:    "http://example.com",
		Method: "GET",
		Data:   moreInfos2,
	}
	hashed2 := req2.Hash("taskid")

	if hashed1 == hashed2 {
		t.Errorf("should not be equal")
	}
}
