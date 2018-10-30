package twitter

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
)

const (
	// MaxPerRequest is the maximum number of tweets returned per api request
	MaxPerRequest int = 100

	// BaseURL is the twitter API base url
	BaseURL string = "https://api.twitter.com/"

	// TokenURL is the twitter token API url
	TokenURL = BaseURL + "oauth2/token"

	// SearchURL is the twitter search API url
	SearchURL = BaseURL + "1.1/search/tweets.json"
)

// Decoder is a convenience interface for testing purposes
type Decoder interface {
	Decode(interface{}) error
}

// DecoderFunc is a func impl of Decoder
type DecoderFunc func(interface{}) error

// Decode implements Decoder
func (d DecoderFunc) Decode(o interface{}) error {
	return d(o)
}

// Doer is a convenience interface for testing purposes
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// DoerFunc is a func impl of Doer
type DoerFunc func(*http.Request) (*http.Response, error)

// Do implements Doer
func (d DoerFunc) Do(r *http.Request) (*http.Response, error) {
	return d(r)
}

// NoopDoer is a test helper that returns &http.Response{}, nil
var NoopDoer = DoerFunc(func(*http.Request) (*http.Response, error) {
	return &http.Response{}, nil
})

// SearchAPIResponse represents the responses from the twitter search API
type SearchAPIResponse struct {
	Statuses []map[string]interface{}
}

type tokenAPIResponse struct {
	AccessToken string `json:"access_token"`
}

// NewAPI is a constructor for API
//
// key: Consumer API Key
// secret: Consumer API Secret Key
func NewAPI(key string, secret string) *API {
	return &API{
		key:    key,
		secret: secret,
		client: &http.Client{},
		decoderFactory: func(r io.Reader) Decoder {
			return json.NewDecoder(r)
		},
		requestFactory: func(method string, url string, reader io.Reader) (*http.Request, error) {
			return http.NewRequest(method, url, reader)
		},
	}
}

// API provides access the Twitter API
type API struct {
	key            string
	secret         string
	bearerToken    string
	client         Doer
	decoderFactory func(io.Reader) Decoder
	requestFactory func(string, string, io.Reader) (*http.Request, error)
}

// SetClient setter for client. This is for testing purposes
func (a *API) SetClient(client Doer) {
	a.client = client
}

// SetDecoderFactory setter for decoderFactory. This is for testing purposes
func (a *API) SetDecoderFactory(decoderFactory func(io.Reader) Decoder) {
	a.decoderFactory = decoderFactory
}

// SetRequestFactory setter for requestFactory. This is for testing purposes
func (a *API) SetRequestFactory(requestFactory func(string, string, io.Reader) (*http.Request, error)) {
	a.requestFactory = requestFactory
}

// SetBearerToken setter for bearerToken. This is for testing purposes
func (a *API) SetBearerToken(bearerToken string) {
	a.bearerToken = bearerToken
}

// FetchHashtag will query the Twitter Search API for a hashtag
//
// hashtag: hashtag to query
// count: number of records to retrieve
// maxId: maxId for the query (see: https://developer.twitter.com/en/docs/tweets/timelines/guides/working-with-timelines)
func (a *API) FetchHashtag(hashtag string, count int, maxID int64) (SearchAPIResponse, error) {

	if a.bearerToken == "" {

		bearerToken, err := a.newBearerToken()

		if err != nil {
			return SearchAPIResponse{}, err
		}

		a.bearerToken = bearerToken

	}

	auth := authorization(a.bearerToken)
	req, err := a.searchRequest(auth, hashtag, count, maxID)

	if err != nil {
		return SearchAPIResponse{}, err
	}

	resp, err := a.client.Do(req)

	if err != nil {
		return SearchAPIResponse{}, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return SearchAPIResponse{}, err
	}

	if resp.StatusCode >= 400 {
		return SearchAPIResponse{}, errors.New(string(bodyBytes))
	}

	var result SearchAPIResponse

	d := a.decoderFactory(bytes.NewBuffer(bodyBytes))
	err = d.Decode(&result)

	if err != nil {
		return SearchAPIResponse{}, err
	}

	return result, nil

}

func (a API) newBearerToken() (string, error) {

	auth := tokenAuthorization(a.key, a.secret)

	req, err := a.tokenRequest(auth)
	if err != nil {
		return "", err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}

	d := a.decoderFactory(resp.Body)
	respBody := tokenAPIResponse{}

	err = d.Decode(&respBody)
	if err != nil {
		return "", err
	}

	return respBody.AccessToken, nil

}

func (a API) searchRequest(auth string, q string, count int, maxID int64) (*http.Request, error) {

	req, err := a.requestFactory("GET", SearchURL, nil)

	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("q", q)
	params.Add("lang", "en")
	params.Add("count", fmt.Sprintf("%v", int(math.Min(float64(count), float64(MaxPerRequest)))))
	params.Add("include_entities", "false")

	if maxID > 0 {
		params.Add("max_id", fmt.Sprintf("%v", maxID))
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Add("Authorization", auth)

	return req, nil

}

func authorization(token string) string {

	sb := strings.Builder{}
	sb.WriteString("Bearer ")
	sb.WriteString(token)

	return sb.String()

}

func (a API) tokenRequest(auth string) (*http.Request, error) {

	reqBody := url.Values{
		"grant_type": []string{"client_credentials"},
	}

	req, err := a.requestFactory("POST", TokenURL, strings.NewReader(reqBody.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", auth)

	return req, nil

}

func tokenAuthorization(key string, secret string) string {

	auth := strings.Builder{}
	auth.WriteString(key)
	auth.WriteString(":")
	auth.WriteString(secret)

	b64auth := base64.StdEncoding.EncodeToString([]byte(auth.String()))

	header := strings.Builder{}
	header.WriteString("Basic ")
	header.WriteString(b64auth)

	return header.String()

}
