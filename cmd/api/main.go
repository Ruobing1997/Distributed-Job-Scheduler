package main

import (
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
)

func setupReverseProxy(target string) gin.HandlerFunc {
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()
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

	r.POST("/register", func(c *gin.Context) {
		api.RegisterUserHandler(c)
	})
	r.POST("/login", func(c *gin.Context) {
		api.LoginUserHandler(c)
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
