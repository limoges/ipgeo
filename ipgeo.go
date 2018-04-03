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

// A NetworkLocator maps an IP address to a known network.
type NetworkLocator interface {
	FindNetworkLocation(addr net.IP) (LocationID, error)
}

type UnknownLocationError struct {
	Addr net.IP
}

func (e UnknownLocationError) Error() string {
	return fmt.Sprintf("unknown location for %s", e.Addr)
}

// A Geolocator is a data structure holding the necessary components to allow
// ip-based geolocation. This Locator operates by first finding a network match
// for an ip, then mapping the location associated with this network.
type Geolocator struct {
	NetLoc     NetworkLocator
	Repository LocationRepository
}

// LocateIP takes an IPv4 address and attempts geolocation, providing an
// approximate geographic location associated with the IP address.
func (l Geolocator) LocateIP(addr net.IP) (Location, error) {
	var loc Location
	locationID, err := l.NetLoc.FindNetworkLocation(addr)
	if err != nil {
		return loc, UnknownLocationError{Addr: addr}
	}
	location, err := l.Repository.FindByID(locationID)
	if err != nil {
		return loc, fmt.Errorf("location unavailable: %d", locationID)
	}
	return location, nil
}
