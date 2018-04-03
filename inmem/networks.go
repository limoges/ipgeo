package inmem

import (
	"net"

	"github.com/yl2chen/cidranger"
)

// networkEntries implements a basic storage that can be used for querying
// a networkEntry that matches an IP address.
type networkEntries interface {
	Insert(e networkEntry) error
	FindNetworks(ip net.IP) ([]networkEntry, error)
}

// networkList implements networkEntries using a list.
// It has access time O(n) where n=len(list). It is a slow but easily verifiable
// implementation.
type networkList struct {
	list []networkEntry
}

func newNetworkList() *networkList {
	var list []networkEntry
	return &networkList{list: list}
}

func (l *networkList) Insert(e networkEntry) error {
	l.list = append(l.list, e)
	return nil
}

func (l networkList) FindNetworks(ip net.IP) ([]networkEntry, error) {
	var matches []networkEntry
	for _, entry := range l.list {
		if entry.Network.Contains(ip) {
			matches = append(matches, entry)
		}
	}
	return matches, nil
}

// networkTrie implements networkEntries using a trie (prefix tree).
// It should have access time corresponding to O(k) where k=len(key).
type networkTrie struct {
	r cidranger.Ranger
}

func newNetworkTrie() *networkTrie {
	return &networkTrie{r: cidranger.NewPCTrieRanger()}
}

func (t networkTrie) Insert(e networkEntry) error {
	return t.r.Insert(storedEntry{in: e})
}

func (t networkTrie) FindNetworks(ip net.IP) ([]networkEntry, error) {
	entries, err := t.r.ContainingNetworks(ip)
	if err != nil {
		return nil, err
	}
	var matches []networkEntry
	for _, entry := range entries {
		matches = append(matches, entry.(storedEntry).in)
	}
	return matches, nil
}

type storedEntry struct {
	in networkEntry
}

func (e storedEntry) Network() net.IPNet {
	return *e.in.Network
}
