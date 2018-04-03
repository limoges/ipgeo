package ipgeo_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/limoges/ipgeo"
	"github.com/stretchr/testify/assert"
)

// TestGeolocator aims to validate that a failure is properly propagated from
// the depencencies to the Geolocator.
func TestGeolocator(t *testing.T) {
	addr := net.IPv4(213, 123, 101, 56)
	loc := ipgeo.Location{Country: "Canada"}
	success := successMock{
		Network:    &net.IPNet{IP: addr, Mask: addr.DefaultMask()},
		LocationID: ipgeo.LocationID(2),
		Location:   loc,
	}
	failure := failureMock{}
	testcases := []struct {
		addr   net.IP
		mapper ipgeo.NetworkMapper
		netloc ipgeo.NetworkLocator
		repo   ipgeo.LocationRepository
		result ipgeo.Location
		err    bool
	}{
		{addr: addr, mapper: success, netloc: success, repo: success, result: loc, err: false},
		{addr: net.IPv4(111, 23, 46, 22),
			mapper: success, netloc: success, repo: success, err: true},
		{addr: addr, mapper: failure, netloc: success, repo: success, err: true},
		{addr: addr, mapper: success, netloc: failure, repo: success, err: true},
		{addr: addr, mapper: success, netloc: success, repo: failure, err: true},
	}

	for _, testcase := range testcases {
		locator := ipgeo.Geolocator{
			Mapper:     testcase.mapper,
			NetLoc:     testcase.netloc,
			Repository: testcase.repo,
		}
		result, err := locator.LocateIP(testcase.addr)
		if testcase.err {
			assert.NotNil(t, err)
		} else {
			assert.Equal(t, testcase.result, result)
			assert.Nil(t, err)
		}
	}
}

type successMock struct {
	Network    *net.IPNet
	Location   ipgeo.Location
	LocationID ipgeo.LocationID
}

func (m successMock) FindByID(id ipgeo.LocationID) (ipgeo.Location, error) {
	if id == m.LocationID {
		return m.Location, nil
	}
	return ipgeo.Location{}, fmt.Errorf("chaining problem")
}

func (m successMock) FindNetwork(addr net.IP) (*net.IPNet, error) {
	if m.Network.Contains(addr) {
		return m.Network, nil
	}
	return nil, fmt.Errorf("testcase problem")
}

func (m successMock) Map(network *net.IPNet) (ipgeo.LocationID, error) {
	if m.Network == network {
		return m.LocationID, nil
	}
	return ipgeo.LocationID(0), fmt.Errorf("chaining problem")
}

type failureMock struct{}

func (m failureMock) FindByID(id ipgeo.LocationID) (ipgeo.Location, error) {
	return ipgeo.Location{}, fmt.Errorf("unknown location")
}

func (m failureMock) FindNetwork(addr net.IP) (*net.IPNet, error) {
	return nil, fmt.Errorf("unknown network")
}

func (m failureMock) Map(network *net.IPNet) (ipgeo.LocationID, error) {
	return ipgeo.LocationID(0), fmt.Errorf("unknown location")
}
