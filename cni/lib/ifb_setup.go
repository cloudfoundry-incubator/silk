package lib

import "code.cloudfoundry.org/silk/cni/config"

type IFB struct {
}

func (ifb *IFB) Setup(cfg *config.Config) error {
	return nil
}
