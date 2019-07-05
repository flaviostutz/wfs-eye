package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/paulsmith/gogeos/geos"
)

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func completeDate(dateStr string, midnight bool) string {
	if len(dateStr) == 10 {
		if midnight {
			return dateStr + "T00:00:00Z"
		} else {
			return dateStr + "T23:59:59Z"
		}
	}
	return dateStr
}

func getDateStartEndFromString(timestr string) (dStart *time.Time, dEnd *time.Time, err error) {
	var dateStart *time.Time
	var dateEnd *time.Time
	if timestr != "" {
		ts := strings.Split(timestr, "/")
		if len(ts) == 1 {
			ts0 := completeDate(ts[0], true)
			d, err := time.Parse(time.RFC3339, completeDate(ts0, true))
			if err != nil {
				return nil, nil, fmt.Errorf("Invalid time parameter. date0=%s", ts[0])
			}
			dateStart = &d
			dateEnd = &d
		} else {
			if ts[0] != "" {
				ts0 := completeDate(ts[0], true)
				d, err := time.Parse(time.RFC3339, completeDate(ts0, true))
				if err != nil {
					return nil, nil, fmt.Errorf("Invalid time parameter. date1=%s", ts0)
				}
				dateStart = &d
			}
			if ts[1] != "" {
				ts1 := completeDate(ts[1], false)
				d, err := time.Parse(time.RFC3339, completeDate(ts1, false))
				if err != nil {
					return nil, nil, fmt.Errorf("Invalid time parameter. date2=%s", ts1)
				}
				dateEnd = &d
			}
		}
	}
	return dateStart, dateEnd, nil
}

func intersectionBBoxStr(bboxstr1 string, bbox2 []float64) (bbox string, err error) {
	bbox1, err := bboxFromString(bboxstr1)
	if err != nil {
		return "", err
	}

	coords1 := make([]geos.Coord, 0)
	coords1 = append(coords1, geos.Coord{X: bbox1[0], Y: bbox1[1]})
	coords1 = append(coords1, geos.Coord{X: bbox1[2], Y: bbox1[1]})
	coords1 = append(coords1, geos.Coord{X: bbox1[2], Y: bbox1[3]})
	coords1 = append(coords1, geos.Coord{X: bbox1[0], Y: bbox1[3]})
	bp1, err := geos.NewPolygon(coords1)

	coords2 := make([]geos.Coord, 0)
	coords2 = append(coords2, geos.Coord{X: bbox2[0], Y: bbox2[1]})
	coords2 = append(coords2, geos.Coord{X: bbox2[2], Y: bbox2[1]})
	coords2 = append(coords2, geos.Coord{X: bbox2[2], Y: bbox2[3]})
	coords2 = append(coords2, geos.Coord{X: bbox2[0], Y: bbox2[3]})
	bp2, err := geos.NewPolygon(coords2)

	rg, err := bp1.Intersection(bp2)
	if err != nil {
		return "", err
	}
	rb, err := rg.Boundary()
	if err != nil {
		return "", err
	}
	coords, err := rb.Coords()
	if err != nil {
		return "", err
	}

	bboxstr3 := fmt.Sprintf("%f,%f,%f,%f", coords[0].X, coords[0].Y, coords[2].X, coords[2].Y)

	return bboxstr3, nil
}

func bboxFromString(bboxstr string) ([]float64, error) {
	bbstr := strings.Split(bboxstr, ",")
	a, erra := strconv.ParseFloat(bbstr[0], 64)
	b, errb := strconv.ParseFloat(bbstr[1], 64)
	c, errc := strconv.ParseFloat(bbstr[2], 64)
	d, errd := strconv.ParseFloat(bbstr[3], 64)
	if erra != nil || errb != nil || errc != nil || errd != nil {
		return []float64{}, fmt.Errorf("Invalid numbers in bounding box")
	}
	return []float64{a, b, c, d}, nil
}
