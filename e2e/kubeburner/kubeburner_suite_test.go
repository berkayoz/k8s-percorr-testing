package kubeburner

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubeburner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KubeBurner Suite")
}
