package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"

	cors "github.com/itsjamie/gin-cors"
)

type HTTPServer struct {
	server *http.Server
	router *gin.Engine
}

type Options struct {
	WFSURL        string
	MongoDBName   string
	MongoAddress  string
	MongoUsername string
	MongoPassword string
	MongoSession  *mgo.Session
}

func NewHTTPServer(opt Options) *HTTPServer {
	router := gin.Default()

	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET",
		RequestHeaders:  "Origin, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          24 * 3600 * time.Second,
		Credentials:     false,
		ValidateHeaders: false,
	}))

	h := &HTTPServer{server: &http.Server{
		Addr:    ":4000",
		Handler: router,
	}, router: router}

	logrus.Debugf("Connecting to MongoDB")
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    strings.Split(opt.MongoAddress, ","),
		Timeout:  2 * time.Second,
		Database: opt.MongoDBName,
		Username: opt.MongoUsername,
		Password: opt.MongoPassword,
	}

	var mongoSession *mgo.Session
	for i := 0; i < 30; i++ {
		ms, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			logrus.Infof("Couldn't connect to mongdb. err=%s", err)
			time.Sleep(1 * time.Second)
			logrus.Infof("Retrying...")
			continue
		}
		mongoSession = ms
		logrus.Infof("Connected to MongoDB successfully")
		break
	}

	if mongoSession == nil {
		logrus.Errorf("Couldn't connect to MongoDB")
		os.Exit(1)
	}

	opt.MongoSession = mongoSession

	logrus.Infof("Initializing HTTP Handlers...")
	h.setupWFSHandlers(opt)
	h.setupViewHandlers(opt)

	return h
}

//Start the main HTTP Server entry
func (s *HTTPServer) Start() error {
	logrus.Infof("Starting HTTP Server on port 4000")
	return s.server.ListenAndServe()
}
