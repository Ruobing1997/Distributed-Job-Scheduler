package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http/httputil"
	"net/url"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of http requests",
		},
		[]string{"method", "endpoint", "status"},
	)
)

func setupReverseProxy(target string) gin.HandlerFunc {
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
		status := c.Writer.Status()
		endpoint := c.Request.URL.Path
		method := c.Request.Method
		HttpRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprintf("%d", status)).Inc()
	}
}

func main() {
	r := gin.Default()
	prometheus.MustRegister(HttpRequestsTotal)
	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.Static("/frontend", "./app/frontend")
	r.Static("/generate", "./app/frontend/generate")
	r.Static("/update", "./app/frontend/update")
	r.Static("/imgs", "./app/frontend/imgs")

	r.Any("/api/*any", setupReverseProxy("http://manager:9090"))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.Run() // listen and serve on 0.0.0.0:8080
}
