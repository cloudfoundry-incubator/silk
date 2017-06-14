package lib_test

import (
	"code.cloudfoundry.org/silk/cni/lib/fakes"

	"code.cloudfoundry.org/silk/cni/config"

	"code.cloudfoundry.org/silk/cni/lib"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"
)

var _ = Describe("IfbCreator", func() {
	var (
		fakeNetlinkAdapter *fakes.NetlinkAdapter
		ifbCreator         *lib.IFBCreator
		cfg                *config.Config
	)
	BeforeEach(func() {
		fakeNetlinkAdapter = &fakes.NetlinkAdapter{}
		ifbCreator = &lib.IFBCreator{
			NetlinkAdapter: fakeNetlinkAdapter,
		}

		cfg = &config.Config{}

	})
	It("creates an IFB device", func() {
		Expect(ifbCreator.Create(cfg)).To(Succeed())

		Expect(fakeNetlinkAdapter.LinkAddCallCount()).To(Equal(1))
		Expect(fakeNetlinkAdapter.LinkAddArgsForCall(0)).To(Equal(&netlink.Ifb{}))

	})
})
