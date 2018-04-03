package ipgeo

// A Location represents a geographic region of varying size, from country,
// to administrative division and subdivision, to city.
type Location struct {
	Country      string
	Subdivision1 string
	Subdivision2 string
	City         string
}
