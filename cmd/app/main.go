package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ghostsquad/currency-converter-practice/internal/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

const binaryName = "currency-converter-practice"

func main() {
	streams := config.NewStdIOStreams()
	cfg := config.Config{
		IOStreams: streams,
	}

	// TODO make the output stream configurable
	// TODO an error here indicates the inability to log
	// Depending on requirements, logging may be absolutely required,
	// and so we should handle the error in some way, such as a panic.
	// Logging can also cause applications to lag, due to latency in writing to the desired stream
	// It may be helpful to use a specific logger with various behaviors, such as
	// - best-effort
	// - retries
	// - buffering
	// TODO use an actual logger to avoid the mistakes like forgetting to suffix with \n
	fmt.Fprintf(streams.ErrOut(), "%s version: %s %s %s %s\n", binaryName, version, commit, date, builtBy)
	if err := env.Parse(&cfg); err != nil {
		fmt.Fprintf(streams.ErrOut(), "%+v\n", err)
		os.Exit(1)
		return
	}

	ctx := context.Background()
	r := setupRouter(cfg)

	listenAddr := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: r,
	}

	fmt.Fprintf(streams.ErrOut(), "listening on %s\n", listenAddr)

	var group run.Group

	group.Add(run.SignalHandler(ctx, syscall.SIGINT, syscall.SIGTERM))
	group.Add(func() error {
		return srv.ListenAndServe()
	}, func(err error) {
		// https://github.com/gin-gonic/gin#graceful-shutdown-or-restart

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		fmt.Fprintf(streams.ErrOut(), "rungroup received error: %s\n", err)

		if err = srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(streams.ErrOut(), "Server forced to shutdown: %s\n", err)
		}
	})

	err := group.Run()
	if err != nil {
		// TODO cleanup, these error printouts are probably very repetitive
		fmt.Fprintf(streams.ErrOut(), "final error: %s\n", err)
		os.Exit(1)
		return
	}
}

// TODO this is basically a composition root
// Treat it that way, and keep branching logic to an absolutely minimum
func setupRouter(cfg config.Config) *gin.Engine {
	logger := gin.LoggerWithWriter(cfg.IOStreams.Out())
	r := gin.New()
	r.Use(logger, gin.Recovery())

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
	))

	// TODO metrics endpoints and other liveness endpoints should likely be part of a different listener
	// so that they can be monitored internally but not exposed to the internet
	// https://github.com/gin-gonic/gin#run-multiple-service-using-gin
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	)))

	// Gin currently doesn't support both wild-card routes that overlap/conflict with static routes
	// https://github.com/gin-gonic/gin/issues/2920
	// https://github.com/gin-gonic/gin/issues/2930
	// panic: catch-all wildcard '*path' in new path '/*path' conflicts with existing path segment 'ping' in existing prefix '/ping'

	// A work-around for this would be to suffix *path paths to avoid the conflict
	// We'll implement a prefix method in order to support other required paths
	// Additionally, it would make sense to try to contribute back to Gin, as these issues have been open for a while

	// Additionally, create a convert group so that can capture metrics for only this group
	convertGroup := r.Group("/convert")
	convertGroup.Use(responseDuration(reg))

	const (
		fromParam       = "from"
		fromAmountParam = "fromAmount"
		toParam         = "to"
	)

	convertService, err := NewConvertService(http.Get)
	if err != nil {
		panic(err)
	}

	// TODO better errors/messages, some of these error messages the user should not see
	// TODO log errors as errors that indicate code problems and/or unrecoverable situations
	// TODO log errors as warnings that are recoverable, but non-nominal
	// TODO log interesting events that cannot be captured as metrics (metrics are cheaper and easier than logs!)
	convertGroup.GET(fmt.Sprintf("/:%s/:%s/:%s", fromParam, fromAmountParam, toParam), func(c *gin.Context) {
		currencyFrom := c.Params.ByName(fromParam)
		currencyTo := c.Params.ByName(toParam)
		fromAmountStr := c.Params.ByName(fromAmountParam)

		if currencyFrom == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": `no "from" currency provided`})
			return
		}

		if currencyTo == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": `no "to" currency provided`})
			return
		}

		fromAmount, err := strconv.ParseFloat(fromAmountStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse amount"})
			return
		}

		resp, err := convertService.Convert(currencyFrom, currencyTo, fromAmount)
		if err != nil {
			// TODO need to understand Gin a bit more and the various ways that errors can be presented
			// specifics: can I return a rich object with a stack trace?
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
		return
	})

	return r
}
