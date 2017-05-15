package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

// parseCSV parses yahoo finance historical csv export
func parseCSV(file string) (pp pricePoints, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	// load
	cr := csv.NewReader(f)
	for {
		r, err := cr.Read()
		if err != nil {
			break
		}

		// skip the first line
		if r[0] == "Date" {
			continue
		}

		// parse date
		date, err := time.Parse("2006-01-02", r[0])
		if err != nil {
			panic(err)
		}

		// parse price
		price, err := strconv.ParseFloat(r[5], 64)
		if err != nil {
			panic(err)
		}

		pp.points = append(pp.points, &pricePoint{date, price})
	}

	// sort
	pp.ensureSorted()

	return
}

