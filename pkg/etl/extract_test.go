package etl_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/tniswong/meshify/pkg/etl"
	"github.com/tniswong/meshify/pkg/twitter"
)

var _ = Describe("Extract", func() {

	Describe("HashtagExtractor.Extract()", func() {

		It("Should return an err when the API call fails", func() {

			// given
			apiErr := errors.New("apiError")
			api := MockTwitterAPI{
				FetchHashtagFn: func(hashtag string, count int, maxID int64) (twitter.SearchAPIResponse, error) {
					return twitter.SearchAPIResponse{}, apiErr
				},
			}
			n := 5
			hashtags := []string{"#IoT"}
			hashtagExtractor := NewHashtagExtractor(api, n, hashtags...)

			// when
			_, err := hashtagExtractor.Extract()

			// then
			Expect(err).To(Equal(apiErr))

		})

		It("Should return ErrIDKeyInvalid when unable to get the ID while processing a Record", func() {

			// given
			api := MockTwitterAPI{
				FetchHashtagFn: func(hashtag string, count int, maxID int64) (twitter.SearchAPIResponse, error) {
					statuses := []map[string]interface{}{
						{
							"invalid": "status",
						},
					}
					return twitter.SearchAPIResponse{Statuses: statuses}, nil
				},
			}
			n := 5
			hashtags := []string{"#IoT"}
			hashtagExtractor := NewHashtagExtractor(api, n, hashtags...)

			// when
			_, err := hashtagExtractor.Extract()

			// then
			Expect(err).To(Equal(ErrIDKeyInvalid))

		})

		It("Should merge records resulting from queries to multiple hashtags", func() {

			// given
			iotTag := "#IoT"
			iotTagStatuses := []map[string]interface{}{
				{"id_str": "12345", "text": "Tweet, Tweet! #IoT"},
				{"id_str": "23456", "text": "RT Tweet, Tweet! #IoT"},
			}
			iotTagQueue := Enqueue(iotTagStatuses...)

			anotherTag := "#AnotherTag"
			anotherTagStatuses := []map[string]interface{}{
				{"id_str": "34567", "text": "#AnotherTag"},
			}
			anotherTagQueue := Enqueue(anotherTagStatuses...)

			api := MockTwitterAPI{
				FetchHashtagFn: func(hashtag string, count int, maxID int64) (twitter.SearchAPIResponse, error) {

					resp := twitter.SearchAPIResponse{}

					switch hashtag {
					case iotTag:
						for x := 0; x < count; x++ {
							if statuses, ok := PopOrClose(iotTagQueue); ok {
								resp.Statuses = append(resp.Statuses, statuses)
							} else {
								break
							}
						}
					case anotherTag:
						for x := 0; x < count; x++ {
							if statuses, ok := PopOrClose(anotherTagQueue); ok {
								resp.Statuses = append(resp.Statuses, statuses)
							} else {
								break
							}
						}
					}

					return resp, nil

				},
			}

			n := 5
			hashtags := []string{iotTag, anotherTag}
			hashtagExtractor := NewHashtagExtractor(api, n, hashtags...)

			// when
			records, err := hashtagExtractor.Extract()

			// then
			Expect(err).To(BeNil())
			Expect(len(records)).To(Equal(len(iotTagStatuses) + len(anotherTagStatuses)))

			uniqueIds := map[string]struct{}{}
			for _, val := range records {
				uniqueIds[val["id_str"].(string)] = struct{}{}
			}

			Expect(len(uniqueIds)).To(Equal(len(iotTagStatuses) + len(anotherTagStatuses)))

		})

		It("Should return a max of n records per hashtag", func() {

			// given
			iotTag := "#IoT"
			iotTagStatuses := []map[string]interface{}{
				{"id_str": "12345", "text": "Tweet, Tweet! #IoT"},
				{"id_str": "23456", "text": "RT Tweet, Tweet! #IoT"},
				{"id_str": "34567", "text": "RT Tweet, Tweet! #IoT"},
				{"id_str": "45678", "text": "RT Tweet, Tweet! #IoT"},
				{"id_str": "56789", "text": "RT Tweet, Tweet! #IoT"},
				{"id_str": "67890", "text": "RT Tweet, Tweet! #IoT"},
				{"id_str": "78901", "text": "RT Tweet, Tweet! #IoT"},
				{"id_str": "89012", "text": "RT Tweet, Tweet! #IoT"},
				{"id_str": "90123", "text": "RT Tweet, Tweet! #IoT"},
			}
			iotTagQueue := Enqueue(iotTagStatuses...)

			api := MockTwitterAPI{
				FetchHashtagFn: func(hashtag string, count int, maxID int64) (twitter.SearchAPIResponse, error) {

					resp := twitter.SearchAPIResponse{}

					switch hashtag {
					case iotTag:
						for x := 0; x < count; x++ {
							if statuses, ok := PopOrClose(iotTagQueue); ok {
								resp.Statuses = append(resp.Statuses, statuses)
							} else {
								break
							}
						}
					}

					return resp, nil

				},
			}

			n := 5
			hashtags := []string{iotTag}
			hashtagExtractor := NewHashtagExtractor(api, n, hashtags...)

			// when
			records, err := hashtagExtractor.Extract()

			// then
			Expect(err).To(BeNil())
			Expect(len(records)).To(Equal(n))

			uniqueIds := map[string]struct{}{}
			for _, val := range records {
				uniqueIds[val["id_str"].(string)] = struct{}{}
			}

			Expect(len(uniqueIds)).To(Equal(n))

		})

	})

})
