package config_test

import (
	"net"

	"code.cloudfoundry.org/silk/cni/config"
	"github.com/containernetworking/cni/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ipam config generation", func() {
	It("returns IPAM config object", func() {

		generator := config.IPAMConfigGenerator{}
		ipamConfig, err := generator.GenerateConfig("10.255.30.0/24", "some-network-name", "/some/data/dir")
		Expect(err).NotTo(HaveOccurred())

		subnetAsIPNet, err := types.ParseCIDR("10.255.30.0/24")
		Expect(err).NotTo(HaveOccurred())

		startIP := net.ParseIP("10.255.30.1").To4()
		endIP := net.ParseIP("10.255.30.255").To4()

		Expect(ipamConfig).To(Equal(
			&config.HostLocalIPAM{
				CNIVersion: "0.3.1",
				Name:       "some-network-name",
				IPAM: config.IPAMConfig{
					Type: "host-local",
					Ranges: []config.RangeSet{
						[]config.Range{
							{
								Subnet:     types.IPNet(*subnetAsIPNet),
								RangeStart: startIP,
								RangeEnd:   endIP,
								Gateway:    endIP,
							},
						}},
					Routes:  []*types.Route{},
					DataDir: "/some/data/dir/ipam",
				},
			}))
	})
	Context("when the subnet is invalid", func() {
		It("returns an error", func() {
			generator := config.IPAMConfigGenerator{}
			_, err := generator.GenerateConfig("10.255.30.0/33", "some-network-name", "/some/data/dir")
			Expect(err).To(MatchError("invalid subnet: invalid CIDR address: 10.255.30.0/33"))
		})
	})
})
