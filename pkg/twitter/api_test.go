package twitter_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/tniswong/meshify/pkg/twitter"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var _ = Describe("API", func() {

	Describe("API.FetchHashtag()", func() {

		Context("when bearerToken unset", func() {

			It("should return an error if failed to create tokenRequest", func() {

				// given
				tokenRequestErr := errors.New("tokenRequest error")

				api := NewAPI("key", "secret")
				api.SetRequestFactory(func(method string, url string, reader io.Reader) (*http.Request, error) {

					if method == "POST" && url == TokenURL {
						return nil, tokenRequestErr
					}

					return &http.Request{}, nil

				})

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(Equal(tokenRequestErr))

			})

			It("should return an error if failed to execute tokenRequest", func() {

				// given
				tokenRequestExecutionErr := errors.New("tokenRequest execution error")

				api := NewAPI("key", "secret")
				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "POST" && req.URL.String() == TokenURL {
						return nil, tokenRequestExecutionErr
					}

					return NoopDoer(req)

				}))

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(Equal(tokenRequestExecutionErr))

			})

			It("should return an error if decoder fails", func() {

				// given
				tokenRequestDecoderErr := errors.New("tokenRequest decoder error")

				api := NewAPI("key", "secret")
				api.SetClient(NoopDoer)
				api.SetDecoderFactory(func(reader io.Reader) Decoder {
					return DecoderFunc(func(o interface{}) error {
						return tokenRequestDecoderErr
					})
				})

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(Equal(tokenRequestDecoderErr))

			})

		})

		Context("when bearerToken is set", func() {

			It("should return error if failed to create searchRequest", func() {

				// given
				searchRequestErr := errors.New("searchRequest error")

				api := NewAPI("key", "secret")
				api.SetBearerToken("bearerToken")
				api.SetRequestFactory(func(method string, url string, reader io.Reader) (*http.Request, error) {

					if method == "GET" && url == SearchURL {
						return nil, searchRequestErr
					}

					return &http.Request{}, nil

				})

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(Equal(searchRequestErr))

			})

			It("should return an error if failed to execute searchRequest", func() {

				// given
				searchRequestExecutionErr := errors.New("searchRequest execution error")

				api := NewAPI("key", "secret")
				api.SetBearerToken("bearerToken")

				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						return nil, searchRequestExecutionErr
					}

					return NoopDoer(req)

				}))

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(Equal(searchRequestExecutionErr))

			})

			It("should return an error if searchRequest status >= 400", func() {

				// given
				api := NewAPI("key", "secret")
				api.SetBearerToken("bearerToken")

				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						return &http.Response{
							StatusCode: 400,
							Body:       ioutil.NopCloser(strings.NewReader("Error: 400")),
						}, nil
					}

					return NoopDoer(req)

				}))

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(Not(BeNil()))

			})

			It("should return an error if searchRequest decode fails", func() {

				// given
				searchRequestDecoderErr := errors.New("searchRequest decoder error")

				api := NewAPI("key", "secret")
				api.SetBearerToken("bearerToken")
				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						return &http.Response{
							StatusCode: 200,
							Body:       ioutil.NopCloser(strings.NewReader("{}")),
						}, nil
					}

					return NoopDoer(req)

				}))
				api.SetDecoderFactory(func(reader io.Reader) Decoder {
					return DecoderFunc(func(o interface{}) error {
						return searchRequestDecoderErr
					})
				})

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(Equal(searchRequestDecoderErr))

			})

		})

		Describe("TokenRequest", func() {

			It("Should have body of grant_type=client_credentials", func() {

				// given
				var tokenRequest *http.Request

				api := NewAPI("key", "secret")
				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "POST" && req.URL.String() == TokenURL {
						tokenRequest = req
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "access_token": "authCode"
                        }`)),
						}, nil
					}

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "statuses": []
                        }`)),
						}, nil
					}

					return NoopDoer(req)

				}))

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(BeNil())

				reqBody, _ := ioutil.ReadAll(tokenRequest.Body)
				v, err := url.ParseQuery(string(reqBody))
				Expect(err).To(BeNil())
				Expect(v["grant_type"][0]).To(Equal("client_credentials"))

			})

			It("Should have appropriate headers", func() {

				// given
				var tokenRequest *http.Request

				key := "key"
				secret := "secret"

				api := NewAPI(key, secret)
				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "POST" && req.URL.String() == TokenURL {
						tokenRequest = req
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "access_token": "authCode"
                        }`)),
						}, nil
					}

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "statuses": []
                        }`)),
						}, nil
					}

					return NoopDoer(req)

				}))

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(BeNil())
				Expect(tokenRequest.Header.Get("Content-Type")).To(Equal("application/x-www-form-urlencoded"))
				Expect(tokenRequest.Header.Get("Authorization")).To(Equal("Basic " + base64.StdEncoding.EncodeToString([]byte(key+":"+secret))))

			})

		})

		Describe("SearchRequest", func() {

			It("Should have the bearerToken as the Authorization header", func() {

				// given
				var searchRequest *http.Request

				bearerToken := "bearerToken"

				api := NewAPI("key", "secret")
				api.SetBearerToken(bearerToken)
				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						searchRequest = req
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "statuses": []
                        }`)),
						}, nil
					}

					return NoopDoer(req)

				}))

				// when
				_, err := api.FetchHashtag("#IoT", 5, 0)

				// then
				Expect(err).To(BeNil())
				Expect(searchRequest.Header.Get("Authorization")).To(Equal("Bearer " + bearerToken))

			})

			It("Should include the appropriate query params", func() {

				// given
				var searchRequest *http.Request

				api := NewAPI("key", "secret")
				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "POST" && req.URL.String() == TokenURL {
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "access_token": "authCode"
                        }`)),
						}, nil
					}

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						searchRequest = req
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "statuses": []
                        }`)),
						}, nil
					}

					return NoopDoer(req)

				}))

				hashTag := "#IoT"
				count := 5
				maxID := int64(5)

				// when
				_, err := api.FetchHashtag(hashTag, count, maxID)

				// then
				Expect(err).To(BeNil())
				Expect(searchRequest.URL.Query().Get("q")).To(Equal(hashTag))
				Expect(searchRequest.URL.Query().Get("lang")).To(Equal("en"))
				Expect(searchRequest.URL.Query().Get("include_entities")).To(Equal("false"))
				Expect(searchRequest.URL.Query().Get("max_id")).To(Equal(fmt.Sprint(maxID)))

			})

			It("Should NOT include max_id query param if maxID < 1", func() {

				// given
				var searchRequest *http.Request

				api := NewAPI("key", "secret")
				api.SetClient(DoerFunc(func(req *http.Request) (*http.Response, error) {

					if req.Method == "POST" && req.URL.String() == TokenURL {
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "access_token": "authCode"
                        }`)),
						}, nil
					}

					if req.Method == "GET" && strings.HasPrefix(req.URL.String(), SearchURL) {
						searchRequest = req
						return &http.Response{
							StatusCode: 200,
							Body: ioutil.NopCloser(strings.NewReader(`{
                            "statuses": []
                        }`)),
						}, nil
					}

					return NoopDoer(req)

				}))

				hashTag := "#IoT"
				count := 5
				maxID := int64(0)

				// when
				_, err := api.FetchHashtag(hashTag, count, maxID)

				// then
				Expect(err).To(BeNil())
				Expect(searchRequest.URL.Query().Get("max_id")).To(Equal(""))

			})

		})

	})

})
