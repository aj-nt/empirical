package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aj-nt/empirical/internal/dignity"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: bazi NAME YEAR MONTH DAY HOUR")
		os.Exit(1)
	}

	name := os.Args[1]
	year, _ := strconv.Atoi(os.Args[2])
	month, _ := strconv.Atoi(os.Args[3])
	day, _ := strconv.Atoi(os.Args[4])
	hour, _ := strconv.Atoi(os.Args[5])

	pillars := dignity.ComputeBaZiPillars(year, month, day, hour)

	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Birth: %d-%02d-%02d %02d:00\n", year, month, day, hour)
	fmt.Printf("\nFour Pillars:\n")
	fmt.Printf("  Year:  %s %s (%s)\n", pillars.Year.Stem, pillars.Year.Branch, pillars.Year.Element)
	fmt.Printf("  Month: %s %s (%s)\n", pillars.Month.Stem, pillars.Month.Branch, pillars.Month.Element)
	fmt.Printf("  Day:   %s %s (%s)\n", pillars.Day.Stem, pillars.Day.Branch, pillars.Day.Element)
	fmt.Printf("  Hour:  %s %s (%s)\n", pillars.Hour.Stem, pillars.Hour.Branch, pillars.Hour.Element)
	fmt.Printf("\nDay Master: %s %s\n", pillars.DayMaster.YinYang, pillars.DayMaster.Element)
}