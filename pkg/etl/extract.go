package etl

import (
	"errors"
	"github.com/tniswong/meshify/pkg/twitter"
	"math"
	"strconv"
	"sync"
)

var (
	// ErrIDKeyInvalid returned by HashtagExtractor.Extract() if tweet can not be identified
	ErrIDKeyInvalid = errors.New("key id_str is either not present or invalid")
)

// Extractor is an interface for ETL extraction
type Extractor interface {
	Extract() ([]Record, error)
}

// ExtractorFn is a function impl of Extractor
type ExtractorFn func() ([]Record, error)

// Extract implements Extractor
func (e ExtractorFn) Extract() ([]Record, error) {
	return e()
}

// NoopExtractor returns a zero-length []Record and nil error
var NoopExtractor = ExtractorFn(func() ([]Record, error) {
	return []Record{}, nil
})

// HashtagFetcher is an interface abstraction of twitter.API
type HashtagFetcher interface {
	FetchHashtag(hashtag string, count int, maxID int64) (twitter.SearchAPIResponse, error)
}

type hashtagWorkerResult struct {
	Records []Record
	Err     error
}

// NewHashtagExtractor returns an Extractor that asynchronously queries the twitter api for n tweets belonging to hashtags
//
// api: twitter api
// n: number of tweets to extract per hashtag
// hashtags: which hashtags to query
func NewHashtagExtractor(api HashtagFetcher, n int, hashtags ...string) Extractor {
	return hashtagExtractor{
		api:      api,
		n:        n,
		hashtags: hashtags,
	}
}

type hashtagExtractor struct {
	api      HashtagFetcher
	hashtags []string
	n        int
}

// Extract will query the twitter api and convert the tweets to []Record.
//
// Each hashtag is queried asynchronously in its own goroutine. Each resulting Record is
// hydrated with an extra key: "hashtag", which contains the value of the hashtag that was queried for.
//
// The []Record result from each async hashtag query is then merged into a single []Record.
func (h hashtagExtractor) Extract() ([]Record, error) {

	workerChan := make(chan hashtagWorkerResult, len(h.hashtags))
	wg := &sync.WaitGroup{}

	// for each hashtag, collect the results asynchronously
	for _, hashtag := range h.hashtags {

		wg.Add(1)
		go h.hashtagWorker(hashtag, wg, workerChan)

	}

	wg.Wait()
	close(workerChan)

	// records holds the records from each worker in a single slice
	var records []Record

	for workerResult := range workerChan {

		if workerResult.Err != nil {
			return nil, workerResult.Err
		}

		// merge each slice of records
		records = append(records, workerResult.Records...)

	}

	return records, nil

}

func (h hashtagExtractor) hashtagWorker(hashtag string, wg *sync.WaitGroup, out chan<- hashtagWorkerResult) {

	defer wg.Done()

	var (
		maxID      int64
		allRecords []Record
	)

	for len(allRecords) <= h.n {

		resp, err := h.api.FetchHashtag(hashtag, h.n, maxID)

		if err != nil {
			out <- hashtagWorkerResult{Err: err}
			return
		}

		// turn the twitter.SearchAPIResponse into []Records
		records, minID, err := processResponse(resp, hashtag)

		if err != nil {
			out <- hashtagWorkerResult{Err: err}
			return
		}

		// no records to append, break the loop
		if len(records) < 1 {
			break
		}

		maxID = minID - 1
		allRecords = append(allRecords, records...)

	}

	// cap at h.n or len(allRecords) to prevent index out of bounds
	lastIndex := int(math.Min(float64(h.n), float64(len(allRecords))))

	out <- hashtagWorkerResult{Records: allRecords[:lastIndex]}

}

func processResponse(searchResult twitter.SearchAPIResponse, hashtag string) ([]Record, int64, error) {

	var (
		records []Record
		minID   int64
	)

	// create a record for each result status
	for _, status := range searchResult.Statuses {

		record := Record{}

		// copy result status to the record
		for k, v := range status {
			record[k] = v
		}

		// hydrate with "hashtag" key
		record["hashtag"] = hashtag
		records = append(records, record)

		recordID, err := getRecordID(record)

		if err != nil {
			return nil, 0, err
		}

		if minID == 0 || recordID < minID {
			minID = recordID
		}

	}

	return records, minID, nil

}

func getRecordID(r Record) (int64, error) {

	if idVal, ok := r["id_str"]; ok {
		if idStr, ok := idVal.(string); ok {
			return strconv.ParseInt(idStr, 10, 64)
		}
	}

	return 0, ErrIDKeyInvalid

}
