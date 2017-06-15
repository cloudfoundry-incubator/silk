package lib

import (
	"fmt"

	"code.cloudfoundry.org/silk/cni/config"

	"github.com/vishvananda/netlink"
)

type TokenBucketFilter struct {
	NetlinkAdapter netlinkAdapter
}

func (tbf *TokenBucketFilter) tick2Time(tick uint32) uint32 {
	return uint32(float64(tick) / float64(tbf.NetlinkAdapter.TickInUsec()))
}

func (tbf *TokenBucketFilter) time2Tick(time uint32) uint32 {
	return uint32(float64(time) * float64(tbf.NetlinkAdapter.TickInUsec()))
}

func (tbf *TokenBucketFilter) buffer(rate uint64, burst uint32) uint32 {
	// do reverse of netlink.burst calculation
	return tbf.time2Tick(uint32(float64(burst) * float64(netlink.TIME_UNITS_PER_SEC) / float64(rate)))
}

func (tbf *TokenBucketFilter) limit(rate uint64, latency, buffer uint32) uint32 {
	// do reverse of netlink.latency calculation
	return uint32(float64(rate) / float64(netlink.TIME_UNITS_PER_SEC) * float64(latency+tbf.tick2Time(buffer)))
}

func (tbf *TokenBucketFilter) Setup(rateInBits, burstInBits int, cfg *config.Config) error {
	// Equivalent to
	// tc qdisc add dev cfg.Host.DeviceName root tbf
	//		rate netConf.BandwidthLimits.Rate
	//		burst netConf.BandwidthLimits.Burst
	if rateInBits <= 0 {
		return fmt.Errorf("invalid rate: %d", rateInBits)
	}
	if burstInBits <= 0 {
		return fmt.Errorf("invalid burst: %d", burstInBits)
	}
	link, err := tbf.NetlinkAdapter.LinkByName(cfg.Host.DeviceName)
	if err != nil {
		return fmt.Errorf("get host device: %s", err)
	}
	rateInBytes := rateInBits / 8
	bufferInBytes := tbf.buffer(uint64(rateInBytes), uint32(burstInBits))
	latency := uint32(100000) // 100 msec or 100000 usec
	limitInBytes := tbf.limit(uint64(rateInBytes), latency, uint32(bufferInBytes))

	qdisc := &netlink.Tbf{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(1, 0),
			Parent:    netlink.HANDLE_ROOT,
		},
		Limit:  uint32(limitInBytes),
		Rate:   uint64(rateInBytes),
		Buffer: uint32(bufferInBytes),
	}
	err = tbf.NetlinkAdapter.QdiscAdd(qdisc)
	if err != nil {
		return fmt.Errorf("create qdisc: %s", err)
	}
	return nil
}
