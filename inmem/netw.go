package inmem

import (
	"fmt"
	"net"
	"strconv"

	"github.com/limoges/ipgeo"
	"github.com/spf13/afero"
)

// NetworkLocator is a read-only in-memory implementation of a ipgeo.NetworkLocator.
// It supports reification from a simple csv file.
type NetworkLocator struct {
	networks networkEntries
}

// NewNetworkLocator instantiates a NetworkLocator from the provided CSV-formatted
// file.
func NewNetworkLocator(name string) (*NetworkLocator, error) {
	return NewNetworkLocatorFromFs(afero.NewOsFs(), name)
}

// NewNetworkLocatorFromFs instantiates a NetworkLocator from the provided CSV-formatted
// file, appearing on the provided filesystem.
func NewNetworkLocatorFromFs(fs afero.Fs, name string) (*NetworkLocator, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}

	networks := newNetworkTrie()
	parseCSV(f, func(rec []string) error {
		entry, err := parseNetworkEntry(rec)
		if err != nil {
			return err
		}
		return networks.Insert(entry)
	})
	return &NetworkLocator{networks: networks}, nil
}

// NetworkLocator maps an IPv4 address to the highest resolution network known
// to the NetworkLocator. This implementation uses a prefix tree (trie) to provide
// O(k) time access.
func (m NetworkLocator) FindNetworkLocation(addr net.IP) (ipgeo.LocationID, error) {
	matches, err := m.networks.FindNetworks(addr)
	if err != nil {
		return ipgeo.LocationID(0), err
	}
	switch len(matches) {
	case 0:
		return ipgeo.LocationID(0), fmt.Errorf("no network for: %s", addr)
	case 1:
		return matches[0].LocationID, nil
	default:
		return longestEntry(matches).LocationID, nil
	}
}

// networkEntry represents an entry/row in a CSV file
type networkEntry struct {
	Network    *net.IPNet
	LocationID ipgeo.LocationID
}

// Size is a shortcut method to e.Network.Mask.Size()
func (e networkEntry) Size() int {
	if e.Network == nil {
		return 0
	}
	size, _ := e.Network.Mask.Size()
	return size
}

// parseNetworkEntry transforms a csv record into a valid networkEntry. To be
// valid, a record must contain a CIDR formatted network and a LocationID.
func parseNetworkEntry(rec []string) (networkEntry, error) {
	var e networkEntry
	r := defaultRecord(rec)

	_, network, err := net.ParseCIDR(r.Get(0))
	if err != nil {
		return e, err
	}
	s := r.Get(1)
	if s == "" {
		return e, fmt.Errorf("missing locationID")
	}
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return e, err
	}
	e.Network = network
	e.LocationID = ipgeo.LocationID(i)
	return e, nil
}

// longestEntry is a specialist max() function which must find the network with
// the longest mask (in bits) from a list of network entries.
func longestEntry(entries []networkEntry) *networkEntry {
	if entries == nil || len(entries) == 0 {
		return nil
	}

	longest := entries[0]
	for _, entry := range entries {
		if longest.Size() < entry.Size() {
			longest = entry
		}
	}
	return &longest
}
