package leaser_test

import (
	"net"

	"code.cloudfoundry.org/silk/controller/leaser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cidrpool", func() {
	Describe("Size", func() {
		DescribeTable("returns the number of subnets that can be allocated",
			func(subnetRange string, subnetMask, expectedSize int) {
				cidrPool := leaser.NewCIDRPool(subnetRange, subnetMask)
				Expect(cidrPool.Size()).To(Equal(expectedSize))
			},
			Entry("when the range is /16 and mask is /24", "10.255.0.0/16", 24, 255),
			Entry("when the range is /16 and mask is /20", "10.255.0.0/16", 20, 15),
			Entry("when the range is /16 and mask is /16", "10.255.0.0/16", 16, 0),
		)
	})

	Describe("GetAvailable", func() {
		It("returns a subnet from the pool that is not taken", func() {
			subnetRange := "10.255.0.0/12"
			_, network, _ := net.ParseCIDR(subnetRange)
			cidrPool := leaser.NewCIDRPool(subnetRange, 24)

			results := map[string]int{}

			taken := []string{}
			for i := 0; i < 255; i++ {
				s := cidrPool.GetAvailable(taken)
				results[s]++
				taken = append(taken, s)
			}
			Expect(len(results)).To(Equal(255))

			for result, _ := range results {
				_, subnet, err := net.ParseCIDR(result)
				Expect(err).NotTo(HaveOccurred())
				Expect(network.Contains(subnet.IP)).To(BeTrue())
				Expect(subnet.Mask).To(Equal(net.IPMask{255, 255, 255, 0}))
				// first subnet from range is never allocated
				Expect(subnet.IP.To4()).NotTo(Equal(network.IP.To4()))
			}
		})

		Context("when no subnets are available", func() {
			It("returns an error", func() {
				subnetRange := "10.255.0.0/16"
				cidrPool := leaser.NewCIDRPool(subnetRange, 24)
				taken := []string{}
				for i := 0; i < 255; i++ {
					s := cidrPool.GetAvailable(taken)
					taken = append(taken, s)
				}
				s := cidrPool.GetAvailable(taken)
				Expect(s).To(Equal(""))
			})
		})
	})

	Describe("IsMember", func() {
		var cidrPool *leaser.CIDRPool
		BeforeEach(func() {
			subnetRange := "10.255.0.0/16"
			cidrPool = leaser.NewCIDRPool(subnetRange, 24)
		})
		It("returns true if the subnet is a member of the pool", func() {
			Expect(cidrPool.IsMember("10.255.30.0/24")).To(BeTrue())
		})

		Context("when the subnet start is not a match for an entry", func() {
			It("returns false", func() {
				Expect(cidrPool.IsMember("10.255.30.10/24")).To(BeFalse())
			})
		})

		Context("when the subnet size is not a match", func() {
			It("returns false", func() {
				Expect(cidrPool.IsMember("10.255.30.0/20")).To(BeFalse())
			})
		})
	})
})
