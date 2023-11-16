package fakeslack

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/circleci/ex/httpserver/ginrouter"
	"github.com/circleci/ex/testing/httprecorder"
	"github.com/circleci/ex/testing/httprecorder/ginrecorder"
	"github.com/gin-gonic/gin"
)

type API struct {
	*httprecorder.RequestRecorder
	router *gin.Engine

	mu sync.RWMutex
}

type APIRequest struct {
	Channel string `json:"channel"`
	Message []byte `json:"message"`
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type APIResponse struct {
	Error   string  `json:"error"`
	Ok      bool    `json:"ok"`
	Message Message `json:"message"`
}

func New(ctx context.Context) *API {
	rec := httprecorder.New()
	r := ginrouter.Default(ctx, "fake-slack")
	// record all requests
	r.Use(ginrecorder.Middleware(ctx, rec))

	r.POST("chat.postMessage", func(c *gin.Context) {
		if c.Request.Header.Get("Content-Type") == "" {
			c.JSON(http.StatusBadRequest, struct{ Error string }{
				Error: "POSTs with a body must set a Content-Type header",
			})
		}
		var request APIRequest
		err := json.Unmarshal(rec.LastRequest().Body, &request)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{Error: err.Error()})
		}

		c.JSON(http.StatusOK, APIResponse{
			Message: Message{Type: "message", Text: string(request.Message)},
		})
	})

	return &API{
		RequestRecorder: rec,
		router:          r,
	}
}

func (f *API) Handler() http.Handler {
	return f.router
}
