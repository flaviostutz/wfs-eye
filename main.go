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
	// cacheControl := flag.String("cache-control", "", "HTTP response Cache-Control header contents for all requests. If empty, no header is set.")
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
		WFSURL: *wfsURL,
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
