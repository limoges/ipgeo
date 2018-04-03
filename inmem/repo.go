package inmem

import (
	"fmt"
	"strconv"

	"github.com/limoges/ipgeo"
	"github.com/spf13/afero"
)

// LocationRepo is an read-only in-memory implementation of a ipgeo.LocationRepository.
// It supports reification from a simple csv file.
type LocationRepo struct {
	locations map[ipgeo.LocationID]locationEntry
}

// NewLocationRepo instantiates a LocationRepo from the provided CSV-formatted
// file.
func NewLocationRepo(name string) (*LocationRepo, error) {
	return NewLocationRepoFromFs(afero.NewOsFs(), name)
}

// NewLocationRepoFromFs instantiates a LocationRepo from the provided CSV-formatted
// file, appearing on the provided filesystem.
func NewLocationRepoFromFs(fs afero.Fs, name string) (*LocationRepo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	locations := make(map[ipgeo.LocationID]locationEntry)
	parseCSV(f, func(rec []string) error {
		entry, err := parseLocationEntry(rec)
		if err != nil {
			return err
		}
		if cur, ok := locations[entry.ID]; ok {
			return fmt.Errorf("duplicate entry: old %s new %s", cur, entry)
		}
		locations[entry.ID] = entry
		return nil
	})
	return &LocationRepo{locations: locations}, nil
}

// FindByID looks for the Location associated with the provided LocationID in
// essentially O(1) time.
func (r LocationRepo) FindByID(id ipgeo.LocationID) (ipgeo.Location, error) {
	entry, ok := r.locations[id]
	if !ok {
		return ipgeo.Location{}, fmt.Errorf("not found")
	}
	return entry.Location(), nil
}

// locationEntry represents an entry/row in a locations CSV file
type locationEntry struct {
	ID           ipgeo.LocationID
	Country      string
	Subdivision1 string
	Subdivision2 string
	City         string
}

// Location is a shortcut function to extract a ipgeo.Location from an entry.
func (e locationEntry) Location() ipgeo.Location {
	return ipgeo.Location{
		Country:      e.Country,
		Subdivision1: e.Subdivision1,
		Subdivision2: e.Subdivision2,
		City:         e.City,
	}
}

// parseLocationEntry transforms a record into a valid locationEntry. To be valid,
// a record should contain at least a locationID and a country.
func parseLocationEntry(rec []string) (locationEntry, error) {
	var e locationEntry
	r := defaultRecord(rec)

	i, err := strconv.ParseInt(r.Get(0), 10, 32)
	if err != nil {
		return e, err
	}
	if r.Get(1) == "" {
		return e, fmt.Errorf("missing required country")
	}
	e.ID = ipgeo.LocationID(i)
	e.Country = r.Get(1)
	e.Subdivision1 = r.Get(2)
	e.Subdivision2 = r.Get(3)
	e.City = r.Get(4)
	return e, nil
}
