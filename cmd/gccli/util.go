package main

import (
	"github.com/imroc/req"
)

// RequestURL receives a url in the form of a string and returns
// a map[string]interface{} with the JSON content of that request's
// response
func RequestURL(url string) (map[string]interface{}, error) {
	// use Req object to initiate requests.
	req := req.New()
	req.Get(url)

	// use req package to initiate request.
	r, err := req.Get(url)

	if err != nil {
		return nil, err
	}

	var res map[string]interface{}

	r.ToJSON(&res) // response => struct/map

	return res, nil

}
