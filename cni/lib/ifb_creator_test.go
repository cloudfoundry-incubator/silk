package lib_test

import (
	"errors"
	"net"

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
		cfg.Container.MTU = 1234
		cfg.IFB.DeviceName = "myIfbDeviceName"

	})

	It("creates an IFB device", func() {
		Expect(ifbCreator.Create(cfg)).To(Succeed())

		Expect(fakeNetlinkAdapter.LinkAddCallCount()).To(Equal(1))
		Expect(fakeNetlinkAdapter.LinkAddArgsForCall(0)).To(Equal(&netlink.Ifb{
			LinkAttrs: netlink.LinkAttrs{
				Name:  cfg.IFB.DeviceName,
				Flags: net.FlagUp,
				MTU:   1234,
			},
		}))
	})

	Context("when adding a link fails", func() {

		BeforeEach(func() {
			fakeNetlinkAdapter.LinkAddReturns(errors.New("banana"))
		})

		It("should return a sensible error", func() {
			err := ifbCreator.Create(cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("add a link failed: banana"))
		})
	})
})
