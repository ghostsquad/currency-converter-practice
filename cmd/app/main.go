package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"github.com/pkg/errors"
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

	// TODO extract into a separate unit testable form
	// TODO better error messages, some of these error messages the user should not see
	// TODO log errors as errors that indicate code problems and/or unrecoverable situations
	// TODO log errors as warnings that are recoverable, but non-nominal
	// TODO log interesting events that cannot be captured as metrics (metrics are cheaper and easier than logs!)
	convertGroup.GET(fmt.Sprintf("/:%s/:%s/:%s", fromParam, fromAmountParam, toParam), func(c *gin.Context) {
		currencyFrom := c.Params.ByName(fromParam)
		currencyTo := c.Params.ByName(toParam)
		fromAmountStr := c.Params.ByName(fromAmountParam)

		fromAmount, err := strconv.ParseFloat(fromAmountStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse amount"})
			return
		}

		if currencyFrom == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "no \"from\" currency provided"})
			return
		}

		if currencyFrom == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "no \"to\" currency provided"})
			return
		}

		url := fmt.Sprintf(
			"https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/%s/%s.json",
			currencyFrom,
			currencyTo,
		)

		// TODO this could probably be cached
		// TODO we could examine what the error is and handle it gracefully
		// Possible Scenarios:
		// - throttled -> retry with exponential backoff
		// - bad request -> we did not validate the user input, so maybe we should do that
		// - bad gateway or other networking related errors -> retry with exponential backoff
		resp, err := http.Get(url)
		if err != nil {
			// TODO need to understand Gin a bit more and the various ways that errors can be presented
			// TODO specifics: can I return a rich object with a stack trace?
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrap(err, "upstream api get").Error()})
			return
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		response := map[string]interface{}{}
		err = json.Unmarshal(body, &response)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrap(err, "unmarshalling upstream response").Error()})
			return
		}

		responseDate, ok := response["date"]
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.New("unknown response, no date found").Error()})
		}

		responseDateStr, ok := responseDate.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.New("unknown response, date is not a string").Error()})
		}

		responseConversionRate, ok := response[currencyTo]
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Errorf("unknown response, key [%s] not found", currencyTo).Error()})
		}

		responseConversionRateFloat, ok := responseConversionRate.(float64)
		if !ok {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": errors.Errorf("unknown response, value of [%s], [%s], is not a float", currencyTo, responseConversionRate).Error()},
			)
		}

		c.JSON(http.StatusOK, gin.H{
			"date":       responseDateStr,
			"rate":       responseConversionRateFloat,
			"from":       currencyFrom,
			"fromAmount": fromAmount,
			"to":         currencyTo,
			// TODO this could result in overflow issues, could use some validation
			"toAmount": responseConversionRateFloat * fromAmount,
		})
		return
	})

	return r
}
