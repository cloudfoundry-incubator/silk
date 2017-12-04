package config

import (
	"net"
	"path/filepath"

	"github.com/containernetworking/cni/pkg/types"
)

// TODO use IPAMConfig struct in ipam plugin
type IPAM struct {
	Type    string         `json:"type"`
	Subnet  string         `json:"subnet"`
	Gateway string         `json:"gateway,omitempty"`
	Routes  []*types.Route `json:"routes"`
	DataDir string         `json:"dataDir"`
}

type HostLocalIPAM struct {
	CNIVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	IPAM       IPAM   `json:"ipam"`
}

type IPAMConfigGenerator struct{}

func (IPAMConfigGenerator) GenerateConfig(subnet, network, dataDirPath string) *HostLocalIPAM {
	return &HostLocalIPAM{
		CNIVersion: "0.3.0",
		Name:       network,
		IPAM: IPAM{
			Type:   "host-local",
			Subnet: subnet,
			Routes: []*types.Route{
				&types.Route{
					Dst: net.IPNet{
						IP:   net.IPv4zero,
						Mask: net.CIDRMask(0, 32),
					},
				},
			},
			DataDir: filepath.Join(dataDirPath, "ipam"),
		},
	}
}
