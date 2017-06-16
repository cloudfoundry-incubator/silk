package lib

import (
	"fmt"
	"net"

	"code.cloudfoundry.org/silk/cni/config"
	"github.com/vishvananda/netlink"
)

type IFBCreator struct {
	NetlinkAdapter netlinkAdapter
}

func (ifbCreator *IFBCreator) Create(cfg *config.Config) error {
	err := ifbCreator.NetlinkAdapter.LinkAdd(&netlink.Ifb{
		LinkAttrs: netlink.LinkAttrs{
			Name:  cfg.IFB.DeviceName,
			Flags: net.FlagUp,
			MTU:   cfg.Container.MTU,
		},
	})

	if err != nil {
		return fmt.Errorf("add a link failed: %s", err)
	}

	return nil
}
