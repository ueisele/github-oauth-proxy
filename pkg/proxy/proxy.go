package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
)

type Config struct {
	Port 			int
	ClientId 		string
	ClientSecret 	string
	AllowOrigin		string
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
	router.POST("/access_token", p.accessToken)

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

type token struct {
	AccessToken string `json:"access_token"`
	Scope     	string `json:"scope"`
	TokenType 	string `json:"token_type"`
}

func (p *proxy) accessToken(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Origin", p.config.AllowOrigin)

	code, hasCode := c.GetQuery("code")
	if !hasCode {
		code, hasCode = c.GetPostForm("code")
	}
	redirectUri, hasRedirectUri := c.GetQuery("redirect_uri")
	if !hasRedirectUri {
		redirectUri, hasRedirectUri = c.GetPostForm("redirect_uri")
	}
	if !hasCode || !hasRedirectUri {
		c.AbortWithStatusJSON(http.StatusBadRequest, struct{Error string}{Error: "Requires 'code' and 'redirect_uri' parameters!"})
		return
	}

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token" , nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			struct{Error string}{Error: fmt.Sprintf("Could not request access token from GitHub: %s", err)})
		return
	}

	header := req.Header
	header.Add("Accept", "application/json")

	query := req.URL.Query()
	query.Add("client_id", p.config.ClientId)
	query.Add("client_secret", p.config.ClientSecret)
	query.Add("code", code)
	query.Add("redirect_uri", redirectUri)
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			struct{Error string}{Error: fmt.Sprintf("Could not request access token from GitHub: %s", err)})
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			struct{Error string}{Error: fmt.Sprintf("Could not request access token from GitHub: %s", err)})
		return
	}

	var tokenResponse map[string]interface{}
	err = json.Unmarshal(responseBody, &tokenResponse)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			struct{Error string}{Error: fmt.Sprintf("Could not request access token from GitHub: %s", err)})
		return
	}

	c.JSON(resp.StatusCode, tokenResponse)
}
