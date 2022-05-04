package render_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/lmika/gopkgs/http/middleware/render"
	"github.com/stretchr/testify/assert"
)

func TestInv_UseFrame(t *testing.T) {
	t.Run("should add frame to the list of frames which will be used", func(t *testing.T) {
		rw := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "https://www.example.com/", nil)

		rnd := render.New(fstest.MapFS{
			"index.html": &fstest.MapFile{
				Data: []byte(`Template: {{.alpha}} - {{.bravo}}`),
			},
			"frame.html": &fstest.MapFile{
				Data: []byte(`Frame: [{{.Content}}]`),
			},
		})

		inv := rnd.NewInv()
		inv.UseFrame("frame.html")
		inv.Set("alpha", "Hello")
		inv.Set("bravo", "World")
		inv.HTML(r, rw, http.StatusOK, "index.html")

		assert.Equal(t, http.StatusOK, rw.Result().StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", rw.Header().Get("Content-type"))
		assert.Equal(t, `Frame: [Template: Hello - World]`, rw.Body.String())
	})

	t.Run("should add to any global frames", func(t *testing.T) {
		rw := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "https://www.example.com/", nil)

		rnd := render.New(fstest.MapFS{
			"index.html": &fstest.MapFile{
				Data: []byte(`Template: {{.alpha}} - {{.bravo}}`),
			},
			"frame.html": &fstest.MapFile{
				Data: []byte(`Frame: [{{.Content}}]`),
			},
			"global.html": &fstest.MapFile{
				Data: []byte(`Global: [{{.Content}}]`),
			},
		}, render.WithFrame("global.html"))

		inv := rnd.NewInv()
		inv.UseFrame("frame.html")
		inv.Set("alpha", "Hello")
		inv.Set("bravo", "World")
		inv.HTML(r, rw, http.StatusOK, "index.html")

		assert.Equal(t, http.StatusOK, rw.Result().StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", rw.Header().Get("Content-type"))
		assert.Equal(t, `Global: [Frame: [Template: Hello - World]]`, rw.Body.String())
	})
}

func TestInv_SetFrameArg(t *testing.T) {
	t.Run("should set the argument on all frames", func(t *testing.T) {
		rw := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "https://www.example.com/", nil)

		rnd := render.New(fstest.MapFS{
			"index.html": &fstest.MapFile{
				Data: []byte(`Template: {{.alpha}} - {{.bravo}}`),
			},
			"frame.html": &fstest.MapFile{
				Data: []byte(`{{.frameName}}: [{{.Content}}]`),
			},
			"global.html": &fstest.MapFile{
				Data: []byte(`{{.frameName}}: [{{.Content}}]`),
			},
		}, render.WithFrame("global.html"))

		inv := rnd.NewInv()
		inv.UseFrame("frame.html")
		inv.Set("alpha", "Hello")
		inv.Set("bravo", "World")
		inv.SetFrameArg("frameName", "The Frame")
		inv.HTML(r, rw, http.StatusOK, "index.html")

		assert.Equal(t, http.StatusOK, rw.Result().StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", rw.Header().Get("Content-type"))
		assert.Equal(t, `The Frame: [The Frame: [Template: Hello - World]]`, rw.Body.String())
	})
}