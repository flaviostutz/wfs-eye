package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

var (
	opt       Options
	viewCache map[string]View
)

type View struct {
	Name       *string            `json:"name,omitempty" bson:"name,omitempty"`
	Collection string             `json:"collection,omitempty" bson:"collection,omitempty"`
	StartTime  *time.Time         `json:"startTime,omitempty" bson:"startTime,omitempty"`
	EndTime    *time.Time         `json:"endTime,omitempty" bson:"endTime,omitempty"`
	Limit      *int               `json:"limit,omitempty" bson:"limit,omitempty"`
	BBox       *[]float64         `json:"bbox,omitempty" bson:"bbox,omitempty"`
	FilterAttr *map[string]string `json:"filterAttr,omitempty" bson:"filterAttr,omitempty"`
	LastUpdate time.Time          `json:"lastUpdate,omitempty" bson:"lastUpdate,omitempty"`
}

func (h *HTTPServer) setupViewHandlers(opt0 Options) {
	opt = opt0
	h.router.POST("/views", createView())
	h.router.PUT("/views/:vname", updateView())
	h.router.GET("/views", listViews())
	h.router.GET("/views/:vname", getView())
	h.router.DELETE("/views/:vname", deleteView())
	viewCache = make(map[string]View)
}

func createView() func(*gin.Context) {
	return func(c *gin.Context) {

		var view View
		data, _ := ioutil.ReadAll(c.Request.Body)
		err := json.Unmarshal(data, &view)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Invalid post data. err=%s", err)})
			return
		}
		if view.Name == nil || *view.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "'name' is required"})
			return
		}

		if view.Collection == nil || *view.Collection == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "'collection' is required"})
			return
		}

		if name == view.Collection {
			c.JSON(http.StatusBadRequest, gin.H{"message":fmt.Sprintf("View collection name cannot be the same as the view name")})
			return
		}
		view.LastUpdate = time.Now()

		sc := opt.MongoSession.Copy()
		defer sc.Close()
		st := sc.DB(opt.MongoDBName).C("views")

		//check duplicate
		count, err1 := st.Find(bson.M{"name": view.Name}).Count()
		if err1 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Error checking for existing view name"})
			logrus.Errorf("Error checking for existing view name. err=%s", err1)
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Duplicate view name"})
			return
		}

		logrus.Debugf("Creating view %s", view.Name)
		err0 := st.Insert(view)
		if err0 != nil {
			c.JSON(http.StatusInternalServerError, "Error storing view")
			logrus.Errorf("Error storing view to Mongo. err=%s", err0)
			return
		}
		delete(viewCache, *view.Name)
		c.JSON(http.StatusCreated, gin.H{"message": "View created successfuly"})
	}
}

func updateView() func(*gin.Context) {
	return func(c *gin.Context) {
		logrus.Debugf("updateView")
		name := c.Param("vname")

		var view View
		data, _ := ioutil.ReadAll(c.Request.Body)
		err := json.Unmarshal(data, &view)
		if err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Error updating view. err=%s", err.Error()))
			return
		}

		if view.Collection == nil || *view.Collection == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "'collection' is required"})
			return
		}

		sc := opt.MongoSession.Copy()
		defer sc.Close()
		st := sc.DB(opt.MongoDBName).C("views")

		view.Name = nil

		if name == view.Collection {
			c.JSON(http.StatusBadRequest, gin.H{"message":fmt.Sprintf("View collection name cannot be the same as the view name")})
			return
		}

		logrus.Debugf("Updating view with %v", view)
		count, err1 := st.Find(bson.M{"name": name}).Count()
		if err1 != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Error updating view. err=%s", err.Error()))
			return
		}
		if count == 0 {
			c.JSON(http.StatusNotFound, fmt.Sprintf("Couldn't find view %s", name))
			return
		}
		view.LastUpdate = time.Now()
		err = st.Update(bson.M{"name": name}, bson.M{"$set": view})
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Error updating view")
			logrus.Errorf("Error updating view %s. err=%s", name, err)
			return
		}
		delete(viewCache, name)
		c.JSON(http.StatusOK, gin.H{"message": "View updated successfully"})
	}
}

func listViews() func(*gin.Context) {
	return func(c *gin.Context) {
		sc := opt.MongoSession.Copy()
		defer sc.Close()
		st := sc.DB(opt.MongoDBName).C("views")

		views := make([]View, 0)
		err := st.Find(nil).All(&views)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error listing schedules. err=%s", err.Error()))
			return
		}
		c.JSON(http.StatusOK, views)
	}
}

func getView() func(*gin.Context) {
	return func(c *gin.Context) {
		logrus.Debugf("getView")
		name := c.Param("vname")

		sc := opt.MongoSession.Copy()
		defer sc.Close()
		st := sc.DB(opt.MongoDBName).C("views")

		var view View
		err := st.Find(bson.M{"name": name}).One(&view)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error getting view. err=%s", err.Error()))
			return
		}

		c.JSON(http.StatusOK, view)
	}
}

func findView(name string) (View, error) {
	sc := opt.MongoSession.Copy()
	defer sc.Close()
	st := sc.DB(opt.MongoDBName).C("views")

	//get view from cache
	view, ok := viewCache[name]
	if ok {
		return view, nil
	}

	//not found in cache. fetch from Mongo
	err := st.Find(bson.M{"name": name}).One(&view)
	if err != nil {
		return View{}, fmt.Errorf("View not found")
	}
	viewCache[name] = view
	return view, nil
}

func deleteView() func(*gin.Context) {
	return func(c *gin.Context) {
		logrus.Debugf("deleteView")
		name := c.Param("vname")

		sc := opt.MongoSession.Copy()
		defer sc.Close()
		st := sc.DB(opt.MongoDBName).C("views")

		err := st.Remove(bson.M{"name": name})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error deleting view. err=%s", err.Error()))
			return
		}
		delete(viewCache, name)
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Deleted view successfully. name=%s", name)})
	}
}
