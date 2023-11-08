package slack

import (
	"fmt"
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

		if auth := r.Header.Get("Authorization"); auth == "" {
			_, _ = w.Write([]byte(`{"error": "not_authed"}`))
		} else {
			_, _ = w.Write([]byte(`{"ok": true}`))
		}

	}))
	t.Cleanup(server.Close)
	client := NewClient(ClientOptions{BaseURL: server.URL, SlackToken: "faketoken"})

	t.Run("successful", func(t *testing.T) {
		err := client.PostMessage(ctx, `{"text": "Hello, world!"}`, "test_channel")
		fmt.Print(err)
		assert.NilError(t, err)
		fmt.Print("\nHEADER: ", recorder.LastRequest().Header["Authorization"])
		assert.Check(t, cmp.Contains(recorder.LastRequest().Header["Authorization"], "Bearer faketoken"))
	})

	t.Run("not_authed", func(t *testing.T) {
		client := NewClient(ClientOptions{BaseURL: server.URL, SlackToken: ""})

		err := client.PostMessage(ctx, `{"text": "Hello, world!"}`, "test_channel")
		fmt.Printf("\nERROR: %v", err)
		assert.ErrorContains(t, err, "not_authed")
	})
}
