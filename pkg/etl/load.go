package etl

import (
	"encoding/csv"
	"fmt"
	"github.com/jeremywohl/flatten"
	"io"
	"sort"
)

// WriteFlusher is an interface for writing records to a CSV file
type WriteFlusher interface {
	Write([]string) error
	Flush()
}

// Flattener is an interface for flattening arbitrary depth map structures
type Flattener interface {
	Flatten(map[string]interface{}) (map[string]interface{}, error)
}

// FlattenerFunc is a function impl of Flattener
type FlattenerFunc func(map[string]interface{}) (map[string]interface{}, error)

// Flatten implements Flattener
func (f FlattenerFunc) Flatten(m map[string]interface{}) (map[string]interface{}, error) {
	return f(m)
}

// DefaultFlattener this is the default flattener implementation
var DefaultFlattener = FlattenerFunc(func(m map[string]interface{}) (map[string]interface{}, error) {
	return flatten.Flatten(m, "", flatten.DotStyle)
})

// Loader is a interface for ETL Loading
type Loader interface {
	Load([]Record) error
}

// LoaderFunc is a function impl of Loader
type LoaderFunc func([]Record) error

// Load implements Loader
func (l LoaderFunc) Load(r []Record) error {
	return l(r)
}

// NoopLoader returns a nil error
var NoopLoader = LoaderFunc(func(r []Record) error {
	return nil
})

// NewCSVLoader is a constructor for CSVLoader
//
// out: Writer where CSV formatted records will be written
func NewCSVLoader(out io.Writer) *CSVLoader {
	return &CSVLoader{
		flatter:      DefaultFlattener,
		writeFlusher: csv.NewWriter(out),
	}
}

// CSVLoader loads records in CSV format and writes them to the provided file
type CSVLoader struct {
	flatter      Flattener
	writeFlusher WriteFlusher
}

// SetFlattener sets the flattener for testing purposes
func (c *CSVLoader) SetFlattener(f Flattener) {
	c.flatter = f
}

// SetRecordWriter

// Load will load the records in CSV format to the provided file
//
// The keys for the resulting CSV will be sorted alphabetically. Keys will automatically be flattened for records with
// nested object structures
func (c *CSVLoader) Load(records []Record) error {

	flattenedRecords, err := c.flattenRecords(records)

	if err != nil {
		return err
	}

	uniqueKeys := uniqueKeysForRecords(flattenedRecords)
	indexLookup := indexLookupForKeys(uniqueKeys)

	csvHeader := make([]string, len(uniqueKeys))

	for _, k := range uniqueKeys {
		csvHeader[indexLookup[k]] = k
	}

	c.writeFlusher.Write(csvHeader)

	for _, record := range flattenedRecords {

		csvRecord := make([]string, len(uniqueKeys))

		for k, v := range record {
			if v != nil {
				csvRecord[indexLookup[k]] = fmt.Sprintf("%v", v)
			}
		}

		c.writeFlusher.Write(csvRecord)

	}

	c.writeFlusher.Flush()

	return nil

}

func (c CSVLoader) flattenRecords(records []Record) ([]Record, error) {

	var flattened []Record

	for _, record := range records {

		flattenedRecord, err := c.flatter.Flatten(record)

		if err != nil {
			return nil, err
		}

		flattened = append(flattened, flattenedRecord)

	}

	return flattened, nil

}

func uniqueKeysForRecords(records []Record) []string {

	var result []string

	// keys may vary between records, so we must track keys we've already encountered
	uniqueKeys := map[string]struct{}{}

	// for each record
	for _, record := range records {

		// iterate over record keys
		for key := range record {

			// if key hasn't been seen yet
			if _, ok := uniqueKeys[key]; !ok {

				// track it, and add it to the result slice
				uniqueKeys[key] = struct{}{}
				result = append(result, key)

			}
		}
	}

	return result

}

func indexLookupForKeys(keys []string) map[string]int {

	keyIndices := map[string]int{}
	sort.Strings(keys)

	for i, key := range keys {
		keyIndices[key] = i
	}

	return keyIndices

}
