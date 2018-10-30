package etl_test

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/tniswong/meshify/pkg/etl"
)

var _ = Describe("Load", func() {

	Describe("CSVLoader.Load()", func() {

		It("Should return an error if the flattener fails", func() {

			// given
			flattenerErr := errors.New("flattener error")

			b := &bytes.Buffer{}
			w := bufio.NewWriter(b)
			csvLoader := NewCSVLoader(w)
			csvLoader.SetFlattener(FlattenerFunc(func(map[string]interface{}) (map[string]interface{}, error) {
				return nil, flattenerErr
			}))

			// when
			err := csvLoader.Load([]Record{
				{"key": "value"},
			})

			// then
			Expect(err).To(Equal(flattenerErr))

		})

		It("Should always write n+1 lines (csv header)", func() {

			b := &bytes.Buffer{}
			writer := bufio.NewWriter(b)
			reader := csv.NewReader(b)

			csvLoader := NewCSVLoader(writer)
			records := []Record{{
				"key": "value1",
			}}

			// when
			err := csvLoader.Load(records)
			Expect(err).To(BeNil())

			// then
			r, err := reader.ReadAll()

			Expect(err).To(BeNil())
			Expect(len(r)).To(Equal(len(records) + 1))

		})

		It("Should sort keys alphabetically, and should order values appropriately", func() {

			b := &bytes.Buffer{}
			writer := bufio.NewWriter(b)
			reader := csv.NewReader(b)

			csvLoader := NewCSVLoader(writer)
			records := []Record{{
				"z": "z value",
				"x": "x value",
				"y": "y value",
				"a": "a value",
				"b": "b value",
				"c": "c value",
			}}

			// when
			err := csvLoader.Load(records)
			Expect(err).To(BeNil())

			// then
			headerLine, err := reader.Read()

			Expect(err).To(BeNil())
			Expect([]string{"a", "b", "c", "x", "y", "z"}).To(Equal(headerLine))

			line1, err := reader.Read()

			Expect(err).To(BeNil())
			Expect([]string{
				"a value",
				"b value",
				"c value",
				"x value",
				"y value",
				"z value",
			}).To(Equal(line1))

		})

	})

})
