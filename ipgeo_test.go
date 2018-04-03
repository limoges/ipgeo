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
		Addr:       addr,
		LocationID: ipgeo.LocationID(2),
		Location:   loc,
	}
	failure := failureMock{}
	testcases := []struct {
		addr   net.IP
		netloc ipgeo.NetworkLocator
		repo   ipgeo.LocationRepository
		result ipgeo.Location
		err    bool
		id     string
	}{
		{id: "1", addr: addr, netloc: success, repo: success, result: loc, err: false},
		{id: "2", addr: addr, netloc: failure, repo: success, err: true},
		{id: "3", addr: addr, netloc: success, repo: failure, err: true},
		{id: "4", addr: net.IPv4(111, 23, 46, 22),
			netloc: success, repo: success, err: true},
	}

	for _, testcase := range testcases {
		locator := ipgeo.Geolocator{
			NetLoc:     testcase.netloc,
			Repository: testcase.repo,
		}
		result, err := locator.LocateIP(testcase.addr)
		if testcase.err {
			assert.NotNil(t, err, testcase.id)
		} else {
			assert.Equal(t, testcase.result, result, testcase.id)
			assert.Nil(t, err, testcase.id)
		}
	}
}

type successMock struct {
	Addr       net.IP
	Location   ipgeo.Location
	LocationID ipgeo.LocationID
}

func (m successMock) FindByID(id ipgeo.LocationID) (ipgeo.Location, error) {
	if id == m.LocationID {
		return m.Location, nil
	}
	return ipgeo.Location{}, fmt.Errorf("chaining problem")
}

func (m successMock) FindNetworkLocation(addr net.IP) (ipgeo.LocationID, error) {
	if m.Addr.String() == addr.String() {
		return m.LocationID, nil
	}
	return ipgeo.LocationID(0), fmt.Errorf("chaining problem")
}

type failureMock struct{}

func (m failureMock) FindByID(id ipgeo.LocationID) (ipgeo.Location, error) {
	return ipgeo.Location{}, fmt.Errorf("unknown location")
}

func (m failureMock) FindNetworkLocation(addr net.IP) (ipgeo.LocationID, error) {
	return ipgeo.LocationID(0), fmt.Errorf("unknown location")
}
