package twitter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTwitter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Twitter Suite")
}
