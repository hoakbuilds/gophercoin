package main

import (
	"log"

	"github.com/imroc/req"
	"github.com/urfave/cli"
)

// RequestURL receives a url in the form of a string and returns
// a map[string]interface{} with the JSON content of that request's
// response
func RequestURL(url string, RESTserver string) ([]byte, error) {
	// use Req object to initiate requests.
	req := req.New()
	req.Get(url)

	// use req package to initiate request.
	r, err := req.Get("http://" + RESTserver + url)

	if err != nil {
		return nil, err
	}

	return r.Bytes(), nil

}

// actionDecorator is used to add additional functionality to
// the command action
func actionDecorator(f func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := f(c); err != nil {
			log.Printf("err: %v", err)
			return err
		}
		return nil
	}
}
