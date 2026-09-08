package linode

// RegionDisplay keeps human readable names while Linode API remains the source of truth.
type RegionDisplay struct {
	ID          string
	Name        string
	Continent   string
}

var RegionNames = map[string]RegionDisplay{
	"jp-tyo-3": {
		ID: "jp-tyo-3", Name: "日本东京3", Continent: "亚洲 Asia",
	},
	"jp-osa": {
		ID: "jp-osa", Name: "日本大阪", Continent: "亚洲 Asia",
	},
	"sg-sin-2": {
		ID: "sg-sin-2", Name: "新加坡2", Continent: "亚洲 Asia",
	},
}
