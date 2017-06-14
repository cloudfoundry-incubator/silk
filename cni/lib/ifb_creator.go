package lib

import "code.cloudfoundry.org/silk/cni/config"

type IFBCreator struct {
	NetlinkAdapter netlinkAdapter
}

func (ifbCreator *IFBCreator) Create(cfg *config.Config) error {
	return nil
}
