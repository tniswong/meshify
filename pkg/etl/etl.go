package etl

import (
	"github.com/tniswong/meshify/pkg/twitter"
	"os"
)

// Record is a convenience type alias for map[string]interface{}
type Record map[string]interface{}

// HashtagsToCSV constructs an ETL that fetches n of each hashtag from api and writes them in CSV format to file
//
// file: File where CSV formatted records will be written
// api: twitter api
// n: number of tweets to fetch per hashtag
// hashtags: hashtags to query
func HashtagsToCSV(file *os.File, api *twitter.API, n int, hashtags ...string) ETL {
	return ETL{
		Extractor: NewHashtagExtractor(api, n, hashtags...),
		Loader:    NewCSVLoader(file),
	}
}

// ETL is a generic construct for an ETL. This implementation skips the transform step.
type ETL struct {
	Extractor Extractor
	Loader    Loader
}

// ETL performs the ETL operation
func (h ETL) ETL() error {

	r, err := h.Extractor.Extract()

	if err != nil {
		return err
	}

	return h.Loader.Load(r)

}
