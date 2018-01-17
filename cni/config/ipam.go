package config

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/ip"
)

type RangeSet []Range

type Range struct {
	RangeStart net.IP      `json:"rangeStart,omitempty"` // The first ip, inclusive
	RangeEnd   net.IP      `json:"rangeEnd,omitempty"`   // The last ip, inclusive
	Subnet     types.IPNet `json:"subnet"`
	Gateway    net.IP      `json:"gateway,omitempty"`
}

type IPAMConfig struct {
	*Range
	Name       string
	Type       string         `json:"type"`
	Routes     []*types.Route `json:"routes"`
	DataDir    string         `json:"dataDir"`
	ResolvConf string         `json:"resolvConf"`
	Ranges     []RangeSet     `json:"ranges"`
	IPArgs     []net.IP       `json:"-"` // Requested IPs from CNI_ARGS and args
}

type HostLocalIPAM struct {
	CNIVersion string     `json:"cniVersion"`
	Name       string     `json:"name"`
	IPAM       IPAMConfig `json:"ipam"`
}

type IPAMConfigGenerator struct{}

func (IPAMConfigGenerator) GenerateConfig(subnet, network, dataDirPath string) (*HostLocalIPAM, error) {
	subnetAsIPNet, err := types.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %s", err)
	}

	return &HostLocalIPAM{
		CNIVersion: "0.3.1",
		Name:       network,
		IPAM: IPAMConfig{
			Type: "host-local",
			Ranges: []RangeSet{
				[]Range{{
					Subnet:     types.IPNet(*subnetAsIPNet),
					RangeStart: ip.NextIP(subnetAsIPNet.IP),
					RangeEnd:   lastIP(types.IPNet(*subnetAsIPNet)),
					Gateway:    lastIP(types.IPNet(*subnetAsIPNet)),
				}},
			},
			Routes:  []*types.Route{},
			DataDir: filepath.Join(dataDirPath, "ipam"),
		},
	}, nil
}

// canonicalizeIPv4 makes sure a provided ip is in standard form
func canonicalizeIPv4(ip *net.IP) error {
	if ip.To4() != nil {
		*ip = ip.To4()
		return nil
	}
	// TODO testdrive
	return fmt.Errorf("IP %s not v4", *ip)
}

// Determine the last IP of a subnet, excluding the broadcast if IPv4
func lastIP(subnet types.IPNet) net.IP {
	if err := canonicalizeIPv4(&subnet.IP); err != nil {
		// TODO testdrive
		panic("err")
	}

	var end net.IP
	for i := 0; i < len(subnet.IP); i++ {
		end = append(end, subnet.IP[i]|^subnet.Mask[i])
	}

	return end
}
