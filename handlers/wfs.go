package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/paulmach/orb/geojson"
)

func (h *HTTPServer) setupWFSHandlers(opt Options) {
	h.router.GET("/collections/:collection/items", getFeatures(opt))
}

func getFeatures(opt Options) func(*gin.Context) {
	return func(c *gin.Context) {
		collection := c.Param("collection")

		bboxstr := c.Query("bbox")
		if bboxstr != "" {
			bboxstr = fmt.Sprintf("%s", bboxstr)
		}
		if bboxstr != "" {
			bb, err := bboxFromString(bboxstr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Invalid 'bbox'. err=%s", err)})
				return
			}
			if !validBBox(bb) {
				c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Invalid 'bbox'. It must be in order (west,north,east,south). err=%s", err)})
				return
			}
		}

		limitstr := c.Query("limit")
		if limitstr != "" {
			limitstr = fmt.Sprintf("%s", limitstr)
		}

		timestr := c.Query("time")
		if timestr != "" {
			timestr = fmt.Sprintf("%s", timestr)
		}

		propertiesFilterStr := ""
		params := c.Request.URL.Query()
		for k, v := range params {
			if k != "time" && k != "bbox" && k != "limit" {
				propertiesFilterStr = fmt.Sprintf("%s&%s=%s", propertiesFilterStr, k, v)
			}
		}

		pc := make([]string, 0)
		fc, err := resolveFeatureCollection(collection, bboxstr, limitstr, timestr, propertiesFilterStr, pc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Error getting collection features. err=%s", err)})
			logrus.Warnf("Error getting collection features. err=%s", err)
			return
		}

		c.JSON(http.StatusOK, fc)
	}
}

func resolveFeatureCollection(collectionName string, bboxstr string, limitstr string, timestr string, propertiesFilterStr string, previousCollectionNames []string) (*geojson.FeatureCollection, error) {
	logrus.Debugf("resolveFeatureCollection. collectionName=%s; bboxstr=%s; limitstr=%s; timestr=%s; propertiesFilterStr=%s; previousCollectionNames=%v", collectionName, bboxstr, limitstr, timestr, propertiesFilterStr, previousCollectionNames)
	if containsString(previousCollectionNames, collectionName) {
		return nil, fmt.Errorf("View %s chain has a circular dependency", collectionName)
	}
	previousCollectionNames = append(previousCollectionNames, collectionName)

	view, err := findView(collectionName)
	if err == nil {
		//ENVELOPE PARAMETERS

		//BBOX
		bboxstr2 := bboxstr
		if bboxstr2 == "" {
			if view.DefaultBBox != nil {
				bb := *view.DefaultBBox
				bboxstr2 = fmt.Sprintf("%f,%f,%f,%f", bb[0], bb[1], bb[2], bb[3])
			}
		}
		if bboxstr2 != "" {
			if view.MaxBBox != nil {
				logrus.Debugf("intersectionBBoxStr %s %v", bboxstr2, *view.MaxBBox)
				bboxstr2, err = intersectionBBoxStr(bboxstr2, *view.MaxBBox)
				if err != nil {
					return nil, err
				}
			}
		}

		//LIMIT
		limitstr2 := limitstr
		if limitstr != "" {
			if view.MaxLimit != nil {
				limit, err := strconv.Atoi(limitstr)
				if err != nil {
					return nil, err
				}
				limit1 := int(math.Min(float64(limit), float64(*view.MaxLimit)))
				limitstr2 = fmt.Sprintf("%d", limit1)
			}
		} else {
			if view.DefaultLimit != nil {
				limitstr2 = fmt.Sprintf("%d", *view.DefaultLimit)
			}
		}

		//TIME
		if timestr == "" {
			if view.DefaultTime != nil {
				timestr = *view.DefaultTime
			}
		}
		dateStart, dateEnd, err := getDateStartEndFromString(timestr)
		if err != nil {
			return nil, fmt.Errorf("Invalid date parameters. err=%s", err)
		}
		sd := ""
		var maxStartDate *time.Time
		var maxEndDate *time.Time
		if view.MaxTimeRange != nil {
			maxStartDate, maxEndDate, err = getDateStartEndFromString(*view.MaxTimeRange)
			if err != nil {
				return nil, err
			}
		}
		if dateStart != nil {
			st1 := *dateStart
			if maxStartDate != nil {
				st2 := *maxStartDate
				if st1.Before(st2) {
					st1 = st2
				}
				sd = st1.Format(time.RFC3339)
			}
		}
		ed := ""
		if dateEnd != nil {
			st1 := *dateEnd
			if maxEndDate != nil {
				st2 := *maxEndDate
				if st1.After(st2) {
					st1 = st2
				}
				ed = st1.Format(time.RFC3339)
			}
		}
		logrus.Debugf("dateStart=%s dateEnd=%s maxStartDate=%s maxEndDate=%s", dateStart, dateEnd, maxStartDate, maxEndDate)

		timestr2 := ""
		if sd != "" || ed != "" {
			timestr2 = fmt.Sprintf("%s/%s", sd, ed)
		}

		//FILTER ATTRIBUTES
		defaultPropertiesFilterStr := ""
		if view.DefaultFilterAttr != nil {
			m := *view.DefaultFilterAttr
			for k, v := range m {
				defaultPropertiesFilterStr = fmt.Sprintf("%s&%s=%s", defaultPropertiesFilterStr, k, v)
			}
		}
		propertiesFilterStr2 := fmt.Sprintf("&%s&%s", propertiesFilterStr, defaultPropertiesFilterStr)

		return resolveFeatureCollection(view.Collection, bboxstr2, limitstr2, timestr2, propertiesFilterStr2, previousCollectionNames)
	}

	logrus.Debugf("Fetching WFS service for collection %s", collectionName)

	if bboxstr != "" {
		bboxstr = fmt.Sprintf("&bbox=%s", bboxstr)
	}
	if limitstr != "" {
		limitstr = fmt.Sprintf("&limit=%s", limitstr)
	}
	if timestr != "" {
		timestr = fmt.Sprintf("&time=%s", timestr)
	}
	q := fmt.Sprintf("%s/collections/%s/items?%s%s%s%s", opt.WFSURL, collectionName, bboxstr, limitstr, timestr, propertiesFilterStr)
	q = strings.ReplaceAll(q, "&&&", "&")
	q = strings.ReplaceAll(q, "&&", "&")
	q = strings.ReplaceAll(q, "?&", "?")
	logrus.Debugf("WFS query: %s", q)
	resp, err := http.Get(q)
	if err != nil {
		return nil, fmt.Errorf("Error requesting WFS service. err=%s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data1, err1 := ioutil.ReadAll(resp.Body)
		if err1 != nil {
			return nil, fmt.Errorf("WFS invocation status != 200. status=%d. body=[failed to get contents]. err=%s", resp.StatusCode, err1)
		}
		return nil, fmt.Errorf("WFS invocation error. status=%d. body=%s", resp.StatusCode, string(data1))
	}

	var fc geojson.FeatureCollection
	data, err0 := ioutil.ReadAll(resp.Body)
	if err0 != nil {
		return nil, fmt.Errorf("Error reading WFS service response. err=%s", err0)
	}

	err = json.Unmarshal(data, &fc)
	if err != nil {
		return nil, fmt.Errorf("Error parsing WFS service response. err=%s", err)
	}
	logrus.Debugf("WFS response OK. feature-count=%d. size-bytes=%d", len(fc.Features), len(data))
	return &fc, nil
}
