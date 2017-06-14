package lib_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("TokenBucketFilter Setup", func() {

	var (
	// cfg                *config.Config
	// fakeNetlinkAdapter *fakes.NetlinkAdapter
	// tbf                TokenBucketFilter
	// fakeLink           netlink.Link
	)

	BeforeEach(func() {
		// fakeLink = &netlink.Bridge{
		// 	LinkAttrs: netlink.LinkAttrs{
		// 		Name: "my-fake-bridge",
		// 	},
		// }
		// fakeNetlinkAdapter = &fakes.NetlinkAdapter{}
		// fakeNetlinkAdapter.LinkByNameReturns(fakeLink, nil)
		// tbf = TokenBucketFilter{
		// 	NetlinkAdapter: fakeNetlinkAdapter,
		// }
		// cfg = &config.Config{}
		// cfg.Host.DeviceName = "host-device"
	})

	It("adds the ifb device", func() {
	})

})
