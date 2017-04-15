// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"code.cloudfoundry.org/silk/controller"
)

type LeaseReleaser struct {
	ReleaseSubnetLeaseStub        func(lease controller.Lease) error
	releaseSubnetLeaseMutex       sync.RWMutex
	releaseSubnetLeaseArgsForCall []struct {
		lease controller.Lease
	}
	releaseSubnetLeaseReturns struct {
		result1 error
	}
	releaseSubnetLeaseReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *LeaseReleaser) ReleaseSubnetLease(lease controller.Lease) error {
	fake.releaseSubnetLeaseMutex.Lock()
	ret, specificReturn := fake.releaseSubnetLeaseReturnsOnCall[len(fake.releaseSubnetLeaseArgsForCall)]
	fake.releaseSubnetLeaseArgsForCall = append(fake.releaseSubnetLeaseArgsForCall, struct {
		lease controller.Lease
	}{lease})
	fake.recordInvocation("ReleaseSubnetLease", []interface{}{lease})
	fake.releaseSubnetLeaseMutex.Unlock()
	if fake.ReleaseSubnetLeaseStub != nil {
		return fake.ReleaseSubnetLeaseStub(lease)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.releaseSubnetLeaseReturns.result1
}

func (fake *LeaseReleaser) ReleaseSubnetLeaseCallCount() int {
	fake.releaseSubnetLeaseMutex.RLock()
	defer fake.releaseSubnetLeaseMutex.RUnlock()
	return len(fake.releaseSubnetLeaseArgsForCall)
}

func (fake *LeaseReleaser) ReleaseSubnetLeaseArgsForCall(i int) controller.Lease {
	fake.releaseSubnetLeaseMutex.RLock()
	defer fake.releaseSubnetLeaseMutex.RUnlock()
	return fake.releaseSubnetLeaseArgsForCall[i].lease
}

func (fake *LeaseReleaser) ReleaseSubnetLeaseReturns(result1 error) {
	fake.ReleaseSubnetLeaseStub = nil
	fake.releaseSubnetLeaseReturns = struct {
		result1 error
	}{result1}
}

func (fake *LeaseReleaser) ReleaseSubnetLeaseReturnsOnCall(i int, result1 error) {
	fake.ReleaseSubnetLeaseStub = nil
	if fake.releaseSubnetLeaseReturnsOnCall == nil {
		fake.releaseSubnetLeaseReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.releaseSubnetLeaseReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *LeaseReleaser) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.releaseSubnetLeaseMutex.RLock()
	defer fake.releaseSubnetLeaseMutex.RUnlock()
	return fake.invocations
}

func (fake *LeaseReleaser) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}