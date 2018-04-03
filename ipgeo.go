package ipgeo

import (
	"fmt"
	"net"
)

// A Location represents a geographic region of varying size, from country,
// to administrative division and subdivision, to city.
type Location struct {
	Country      string
	Subdivision1 string
	Subdivision2 string
	City         string
}

// LocationID uniquely identifies a single geographic location/region.
type LocationID int32

// A LocationRepository provides access to a location store.
type LocationRepository interface {
	FindByID(id LocationID) (Location, error)
}

// A NetworkMapper maps an IP address to a known network.
type NetworkMapper interface {
	FindNetwork(addr net.IP) (*net.IPNet, error)
}

// A NetworkLocator maps a known network to a location.
type NetworkLocator interface {
	Map(network *net.IPNet) (LocationID, error)
}

// A Geolocator is a data structure holding the necessary components to allow
// ip-based geolocation. This Locator operates by first finding a network match
// for an ip, then mapping the location associated with this network.
type Geolocator struct {
	Mapper     NetworkMapper
	NetLoc     NetworkLocator
	Repository LocationRepository
}

// LocateIP takes an IPv4 address and attempts geolocation, providing an
// approximate geographic location associated with the IP address.
func (l Geolocator) LocateIP(addr net.IP) (Location, error) {
	var loc Location
	network, err := l.Mapper.FindNetwork(addr)
	if err != nil {
		return loc, fmt.Errorf("unknown network for: %s", addr)
	}
	locationID, err := l.NetLoc.Map(network)
	if err != nil {
		return loc, fmt.Errorf("unknown location for: %s", network)
	}
	location, err := l.Repository.FindByID(locationID)
	if err != nil {
		return loc, fmt.Errorf("location unavailable: %d", locationID)
	}
	return location, nil
}
