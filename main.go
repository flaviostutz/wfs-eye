package main

import (
	"flag"
	"os"

	"github.com/flaviostutz/wfs-eye/handlers"
	"github.com/sirupsen/logrus"
)

func main() {
	logLevel := flag.String("loglevel", "debug", "debug, info, warning, error")
	wfsURL := flag.String("wfs-url", "", "WFS 3.0 server API URL from which to get features")
	mongoDBName0 := flag.String("mongo-dbname", "", "Mongo db name")
	mongoAddress0 := flag.String("mongo-address", "", "MongoDB address. Example: 'mongo', or 'mongdb://mongo1:1234/db1,mongo2:1234/db1")
	mongoUsername0 := flag.String("mongo-username", "root", "MongoDB username")
	mongoPassword0 := flag.String("mongo-password", "root", "MongoDB password")
	flag.Parse()

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("====Starting WFS-EYE====")

	opt := handlers.Options{
		WFSURL:        *wfsURL,
		MongoDBName:   *mongoDBName0,
		MongoAddress:  *mongoAddress0,
		MongoUsername: *mongoUsername0,
		MongoPassword: *mongoPassword0,
	}

	if opt.MongoAddress == "" {
		logrus.Errorf("'mongo-address' parameter is required")
		os.Exit(1)
	}

	if opt.WFSURL == "" {
		logrus.Errorf("'--wfs-url' is required")
		os.Exit(1)
	}

	logrus.Infof("WFS3 provider URL is %s", opt.WFSURL)
	logrus.Infof("Listening on port 4000...")

	h := handlers.NewHTTPServer(opt)
	err := h.Start()
	if err != nil {
		logrus.Errorf("Error starting server. err=%s", err)
		os.Exit(1)
	}

}
