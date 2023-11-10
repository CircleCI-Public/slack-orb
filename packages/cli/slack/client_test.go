package slack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/circleci/ex/testing/httprecorder"
	"github.com/circleci/ex/testing/testcontext"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func Test_Post_Message(t *testing.T) {
	ctx := testcontext.Background()
	recorder := httprecorder.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := recorder.Record(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		request := &struct {
			Channel string `json:"channel"`
		}{}
		bodyBytes, _ := io.ReadAll(r.Body)

		err = json.Unmarshal(bodyBytes, &request)
		if err != nil || request.Channel == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if auth := r.Header.Get("Authorization"); auth == "" {
			_, _ = w.Write([]byte(`{"error": "not_authed"}`))
		} else {
			_, _ = w.Write([]byte(`{"ok": true}`))
		}

	}))
	t.Cleanup(server.Close)

	t.Run("successful", func(t *testing.T) {
		client := NewClient(ClientOptions{BaseURL: server.URL, SlackToken: "faketoken"})
		err := client.PostMessage(ctx, `{"text": "Hello, world!"}`, "test_channel")
		assert.NilError(t, err)
		assert.Check(t, cmp.Contains(recorder.LastRequest().Header["Authorization"], "Bearer faketoken"))
	})

	t.Run("not_authed", func(t *testing.T) {
		client := NewClient(ClientOptions{BaseURL: server.URL, SlackToken: ""})
		err := client.PostMessage(ctx, `{"text": "Hello, world!"}`, "test_channel")
		assert.ErrorContains(t, err, "not_authed")
	})
}
