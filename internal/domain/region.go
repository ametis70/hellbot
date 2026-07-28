package domain

import "fmt"

// Region holds the name and capital of a galactic campaign region.
type Region struct {
	Name    string
	Capital string
}

// HomeWorldRegion is the region number used for faction homeworlds.
const HomeWorldRegion = 11

// SuperEarthRegion is the region number for Super Earth (Sol System).
const SuperEarthRegion = 0

// regionNames maps Enemy → region number → Region.
// Region 0 is always Super Earth (Sol System).
// Region 11 is the faction homeworld.
// Source: https://helldivers.wiki.gg/wiki/Helldivers_1:Galactic_Campaign
var regionNames = map[Enemy]map[int]Region{
	EnemyCyborg: {
		0:  {Name: "Super Earth", Capital: "Super Earth"},
		1:  {Name: "Sirius Region", Capital: "Stockholm City"},
		2:  {Name: "Polaris Region", Capital: "Thunder Head"},
		3:  {Name: "Pictor Sector", Capital: "New Moscow"},
		4:  {Name: "Sagan Region", Capital: "Highwind"},
		5:  {Name: "Horolium System", Capital: "Providence"},
		6:  {Name: "Gellert Region", Capital: "Gellert City"},
		7:  {Name: "Lacaille Region", Capital: "Bahia Democracia"},
		8:  {Name: "Indi System", Capital: "Winter Hold"},
		9:  {Name: "Ceti System", Capital: "Doral Creek"},
		10: {Name: "Cygni Region", Capital: "New Berlin"},
		11: {Name: "Cyberstan Region", Capital: "Cyberstan"},
	},
	EnemyBug: {
		0:  {Name: "Super Earth", Capital: "Super Earth"},
		1:  {Name: "Wise Region", Capital: "New New York"},
		2:  {Name: "Kruger System", Capital: "Liberty City"},
		3:  {Name: "Ross System", Capital: "Tiberia"},
		4:  {Name: "Struve Region", Capital: "Northman's Creek"},
		5:  {Name: "Xi Tauri Region", Capital: "New Haven"},
		6:  {Name: "Cancri System", Capital: "Freedom Fortress"},
		7:  {Name: "Higgs Region", Capital: "Martyr's Bay"},
		8:  {Name: "Hawking Region", Capital: "Segma Prime"},
		9:  {Name: "Rigel System", Capital: "Freedom Peak"},
		10: {Name: "Aurigae Region", Capital: "Final Frontier"},
		11: {Name: "Kepler System", Capital: "Kepler Prime"},
	},
	EnemyIlluminate: {
		0:  {Name: "Super Earth", Capital: "Super Earth"},
		1:  {Name: "Centaury Region", Capital: "New Hanover"},
		2:  {Name: "Barnard Region", Capital: "Iron Tower"},
		3:  {Name: "Procyon Region", Capital: "White Landing"},
		4:  {Name: "Castor System", Capital: "Justice Bay"},
		5:  {Name: "Orionis Region", Capital: "New Alexandria"},
		6:  {Name: "Prometheus System", Capital: "Ribatishiti"},
		7:  {Name: "Cassiopaiae Region", Capital: "Dal Rage"},
		8:  {Name: "Ursa Region", Capital: "Ultima"},
		9:  {Name: "Canes Region", Capital: "Jiyu Toshi"},
		10: {Name: "Arcturus Region", Capital: "Hawk Nest"},
		11: {Name: "Squ'bai System", Capital: "Squ'bai Shrine"},
	},
}

// GetRegion returns the Region for a given enemy and region number.
// Falls back to a generic name if the region is not found.
func GetRegion(enemy Enemy, region int) Region {
	if factionRegions, ok := regionNames[enemy]; ok {
		if r, ok := factionRegions[region]; ok {
			return r
		}
	}
	return Region{Name: fmt.Sprintf("Region %d", region)}
}

// IsSuperEarth returns true if the region number is Super Earth.
func IsSuperEarth(region int) bool {
	return region == SuperEarthRegion
}

// IsHomeworld returns true if the region number is the faction homeworld.
func IsHomeworld(region int) bool {
	return region == HomeWorldRegion
}

// TotalRegions is the number of non-homeworld, non-Super-Earth regions per faction.
const TotalRegions = 10
