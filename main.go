package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
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
	fmt.Fprintf(streams.ErrOut(), "%s version: %s %s %s %s", binaryName, version, commit, date, builtBy)
	if err := env.Parse(&cfg); err != nil {
		fmt.Fprintf(streams.ErrOut(), "%+v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	r := setupRouter(cfg)
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port),
		Handler: r,
	}

	var group run.Group

	group.Add(run.SignalHandler(ctx, syscall.SIGINT, syscall.SIGTERM))
	group.Add(func() error {
		return srv.ListenAndServe()
	}, func(err error) {
		// https://github.com/gin-gonic/gin#graceful-shutdown-or-restart

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		fmt.Fprintf(streams.ErrOut(), "rungroup received error: %s", err)

		if err = srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(streams.ErrOut(), "Server forced to shutdown: %s", err)
		}
	})

	err := group.Run()
	if err != nil {
		// TODO cleanup, these error printouts are probably very repetitive
		_, err2 := fmt.Fprintf(streams.ErrOut(), "final error: %s", err)
		if err2 != nil {
			panic(err2)
		}
		os.Exit(1)
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
	// Although the original requirement doc specifies that any path should be supported from the root
	// We'll implement a prefix method in order to support other required paths
	// Additionally, it would make sense to try to contribute back to Gin, as these issues have been open for a while

	// Additionally, create a path group so that can capture metrics for only this group
	paths := r.Group("/p")
	paths.Use(responseDuration(reg))

	paths.GET("/*path", func(c *gin.Context) {
		//client, err := aws.NewClient(c)
		//if err != nil {
		//	c.JSON(http.StatusInternalServerError, aws.Response{
		//		Error: err.Error(),
		//	})
		//}
		//c.JSON(response.Code, response.Response)
	})

	return r
}
