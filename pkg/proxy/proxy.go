package proxy

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Config struct {
	Port int
}

type Proxy interface {
	Run()
	Shutdown(ctx context.Context) error
}

type proxy struct {
	config		Config
	doneChannel chan error
	server		*http.Server
}

func NewProxy(config Config, doneChannel chan error) Proxy {
	return &proxy{config: config, doneChannel: doneChannel}
}

func (p *proxy) Run() {
	if p.server != nil {
		panic(fmt.Errorf("server has already been started"))
	}

	router := gin.Default()

	router.GET("/health", p.health)

	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%v", p.config.Port),
		Handler: router,
	}

	go func() {
		log.Printf("Listening and serving HTTP on %s\n", p.server.Addr)
		if err := p.server.ListenAndServe(); err != nil {
			// always completes with error
			p.doneChannel <- err
		}
	}()
}

func (p *proxy) Shutdown(ctx context.Context) error {
	if p.server != nil {
		return p.server.Shutdown(ctx)
	}
	return nil
}

func (p *proxy) health(c *gin.Context) {
	c.JSON(http.StatusOK, struct{Ok bool}{Ok: true})
}
