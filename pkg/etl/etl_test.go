package etl_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/tniswong/meshify/pkg/etl"
)

var _ = Describe("ETL", func() {

	Describe("ETL()", func() {

		It("Should return any error encountered by the Extractor", func() {

			// given
			extractorErr := errors.New("extractor error")
			e := ETL{
				Loader: NoopLoader,
				Extractor: ExtractorFn(func() ([]Record, error) {
					return nil, extractorErr
				}),
			}

			// when
			err := e.ETL()

			// then
			Expect(err).To(Equal(extractorErr))

		})

		It("Should return any error encountered by the Loader", func() {

			// given
			loaderErr := errors.New("loader error")
			e := ETL{
				Extractor: NoopExtractor,
				Loader: LoaderFunc(func([]Record) error {
					return loaderErr
				}),
			}

			// when
			err := e.ETL()

			// then
			Expect(err).To(Equal(loaderErr))

		})

	})

})
