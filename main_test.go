package batcher

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var (
	ctrl *gomock.Controller
)

var _ = BeforeSuite(func() {
	ctrl = gomock.NewController(GinkgoT())
})

func TestDataloader(t *testing.T) {
	defer GinkgoRecover()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataloader Suite")
}
