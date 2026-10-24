package main

import (
	"net"
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/mdoroguncev/toll-calculator/go-kit-example/aggsvc/aggendpoint"
	"github.com/mdoroguncev/toll-calculator/go-kit-example/aggsvc/aggservice"
	"github.com/mdoroguncev/toll-calculator/go-kit-example/aggsvc/aggtransport"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	service := aggservice.New(logger)
	endpoints := aggendpoint.New(service, logger)
	httpHandler := aggtransport.NewHTTPHandler(endpoints, logger)

	httpListener, err := net.Listen("tcp", ":3000")
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
		os.Exit(1)
	}

	logger.Log("transport", "HTTP", "addr", ":3000")
	err = http.Serve(httpListener, httpHandler)
	if err != nil {
		panic(err)
	}
}
