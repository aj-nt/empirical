package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	if len(os.Args) < 10 {
		fmt.Fprintf(os.Stderr, "Usage: rawdump NAME Y M D H MIN TZ LAT LNG\n")
		os.Exit(1)
	}

	cacheDir, err := empirical.EnsureEpheCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ephe error: %v\n", err)
		os.Exit(1)
	}
	swe.SetEphePath(cacheDir)
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)

	name := os.Args[1]
	year, _ := strconv.Atoi(os.Args[2])
	month, _ := strconv.Atoi(os.Args[3])
	day, _ := strconv.Atoi(os.Args[4])
	hour, _ := strconv.Atoi(os.Args[5])
	minute, _ := strconv.Atoi(os.Args[6])
	tzOff, _ := strconv.ParseFloat(os.Args[7], 64)
	lat, _ := strconv.ParseFloat(os.Args[8], 64)
	lng, _ := strconv.ParseFloat(os.Args[9], 64)

	bd := dignity.BirthData{
		Name: name, Year: year, Month: month, Day: day,
		Hour: hour, Minute: minute, TZOffset: tzOff, Lat: lat, Lng: lng,
	}
	bc, err := dignity.ComputeBaseChart(bd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compute error: %v\n", err)
		os.Exit(1)
	}

	type PlanetDump struct {
		Name      string  `json:"name"`
		Lon       float64 `json:"lon"`
		Lat       float64 `json:"lat"`
		Speed     float64 `json:"speed_deg_per_day"`
		Dist      float64 `json:"dist_au"`
		Sign      string  `json:"sign"`
		SignLon   float64 `json:"sign_lon"`
		House     int     `json:"house"`
		Retrograde bool   `json:"retrograde"`
		Declination float64 `json:"declination"`
	}

	type Dump struct {
		Name      string  `json:"name"`
		Date      string  `json:"date_utc"`
		JD        float64 `json:"julian_day"`
		Ayanamsa  float64 `json:"ayanamsa_lahiri"`
		ASC       float64 `json:"asc"`
		MC        float64 `json:"mc"`
		DSC       float64 `json:"dsc"`
		IC        float64 `json:"ic"`
		NorthNode float64 `json:"north_node"`
		SouthNode float64 `json:"south_node"`
		IsDay     bool    `json:"is_day"`
		Planets   []PlanetDump `json:"planets"`
	}

	dump := Dump{
		Name:      name,
		Date:      fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d%+03d:00", year, month, day, hour, minute, 0, int(tzOff)),
		JD:        bc.JD,
		Ayanamsa:  bc.Ayanamsa,
		ASC:       bc.ASC,
		MC:        bc.MC,
		DSC:       bc.DSC,
		IC:        bc.IC,
		NorthNode: bc.NorthNode,
		SouthNode: bc.SouthNode,
		IsDay:     bc.ASC < 180, // rough: Sun above horizon
	}

	// Determine if it's a day chart (Sun above horizon)
	if sun, ok := bc.Tropical["Sun"]; ok {
		dump.IsDay = sun.Lon > bc.ASC && sun.Lon < bc.DSC
		if bc.DSC < bc.ASC {
			dump.IsDay = sun.Lon > bc.ASC || sun.Lon < bc.DSC
		}
	}

	// Planet order
	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn",
		"Uranus", "Neptune", "Pluto", "Chiron", "Ceres", "Pallas", "Juno", "Vesta",
		"Lilith", "Node", "TrueNode", "Eris", "Makemake", "Gonggong"}

	for _, pname := range order {
		pos, ok := bc.Tropical[pname]
		if !ok {
			continue
		}
		sign := dignity.SignForLongitude(pos.Lon)
		signLon := pos.Lon - float64(int(pos.Lon/30))*30
		house := ((int(pos.Lon/30) - int(bc.ASC/30) + 12) % 12) + 1
		decl := 0.0
		if d, ok := bc.Declinations[pname]; ok {
			decl = d
		}
		dump.Planets = append(dump.Planets, PlanetDump{
			Name:        pname,
			Lon:         pos.Lon,
			Lat:         pos.Lat,
			Speed:       pos.Speed,
			Dist:        pos.Dist,
			Sign:        sign,
			SignLon:     signLon,
			House:       house,
			Retrograde:  pos.Speed < 0,
			Declination: decl,
		})
	}

	js, _ := json.MarshalIndent(dump, "", "  ")
	fmt.Println(string(js))
}
