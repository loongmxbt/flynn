package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/flynn/flynn/pkg/httpclient"
)

type Client struct {
	*httpclient.Client
}

// ErrNotFound is returned when a resource is not found (HTTP status 404).
var ErrNotFound = errors.New("logaggregator: resource not found")

// newClient creates a generic Client object, additional attributes must
// be set by the caller
func newClient(url string, http *http.Client) *Client {
	c := &Client{
		Client: &httpclient.Client{
			ErrNotFound: ErrNotFound,
			URL:         url,
			HTTP:        http,
		},
	}
	return c
}

// NewClient creates a new Client pointing at uri and using key for
// authentication.
func New(uri string) (*Client, error) {
	return NewWithHTTP(uri, http.DefaultClient)
}

func NewWithHTTP(uri string, httpClient *http.Client) (*Client, error) {
	if uri == "" {
		uri = "http://flynn-logaggregator-api.discoverd"
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	return newClient(u.String(), httpClient), nil
}

// GetLog returns a ReadCloser log stream of the log channel with ID channelID.
// Each line returned will be a JSON serialized logaggregator.Message.
//
// If lines is above zero, the number of lines returned will be capped at that
// value. Otherwise, all available logs are returned. If follow is true, new log
// lines are streamed after the buffered log.
func (c *Client) GetLog(channelID string, lines int, follow bool) (io.ReadCloser, error) {
	path := fmt.Sprintf("/log/%s", channelID)
	query := url.Values{}
	if lines > 0 {
		query.Add("lines", strconv.Itoa(lines))
	}
	if follow {
		query.Add("follow", "true")
	}
	if encodedQuery := query.Encode(); encodedQuery != "" {
		path = fmt.Sprintf("%s?%s", path, encodedQuery)
	}
	res, err := c.RawReq("GET", path, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
