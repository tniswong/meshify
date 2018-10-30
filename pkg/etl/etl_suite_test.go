package etl_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tniswong/meshify/pkg/twitter"
)

func TestEtl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etl Suite")
}

type MockTwitterAPI struct {
	FetchHashtagFn func(hashtag string, count int, maxID int64) (twitter.SearchAPIResponse, error)
}

func (m MockTwitterAPI) FetchHashtag(hashtag string, count int, maxID int64) (twitter.SearchAPIResponse, error) {

	if m.FetchHashtagFn != nil {
		return m.FetchHashtagFn(hashtag, count, maxID)
	}

	return twitter.SearchAPIResponse{}, nil

}

func Enqueue(values ...map[string]interface{}) chan map[string]interface{} {

	queue := make(chan map[string]interface{}, len(values))

	for _, value := range values {
		queue <- value
	}

	return queue

}

func PopOrClose(queue chan map[string]interface{}) (map[string]interface{}, bool) {
	select {
	case result, ok := <-queue:
		return result, ok
	default:
		close(queue)
		return nil, false
	}
}
