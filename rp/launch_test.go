package rp

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLaunch(t *testing.T) {
	l := NewLaunch(nil, "test", "test", "test", nil)
	assert.NotNil(t, l)
}

func TestStartLaunch(t *testing.T) {
	t.Run("Correctly created", func(t *testing.T) {
		okResponse := `{"id": "testid"}`
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "/test_project/launch", r.URL.Path)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(okResponse))
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
			Project:  "test_project",
		}
		l := &Launch{
			client: c,
		}
		err := l.Start()

		assert.Equal(t, "testid", l.Id)
		assert.NoError(t, err)
	})

	t.Run("Differ status code", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
		}
		l := &Launch{
			client: c,
		}
		err := l.Start()

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "failed with status 200 OK")
	})
}

func TestFinalizeLaunch(t *testing.T) {
	t.Run("Successful finalize", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			d, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			rx, _ := regexp.Compile(`\{\"status\"\:\"test status\"\,\"end\_time\"\:\d+\}`)
			assert.Regexp(t, rx, string(d))
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
		}
		l := &Launch{
			client: c,
		}
		err := l.Stop("test status")

		assert.NoError(t, err)
	})

	t.Run("Wrong status code fail", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
		}
		l := &Launch{
			client: c,
		}
		err := l.Stop("")

		assert.EqualError(t, err, "failed with status 500 Internal Server Error")
	})
}

func TestStopLaunch(t *testing.T) {
	t.Run("Successful run", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test_project/launch/id123/stop", r.URL.Path)
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
			Project:  "test_project",
		}
		l := &Launch{
			client: c,
			Id:     "id123",
		}
		err := l.Stop("")

		assert.NoError(t, err)
	})

	t.Run("Wrong status code", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		s := httptest.NewServer(h)
		defer s.Close()

		l := &Launch{
			client: &Client{
				Endpoint: s.URL,
			},
		}
		err := l.Stop("")
		assert.EqualError(t, err, "failed with status 500 Internal Server Error")
	})
}

func TestFinishLaunch(t *testing.T) {
	t.Run("Successful finish", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test_project/launch/id123/finish", r.URL.Path)
		})
		s := httptest.NewServer(h)
		defer s.Close()

		l := &Launch{
			client: &Client{
				Endpoint: s.URL,
				Project:  "test_project",
			},
			Id: "id123",
		}
		err := l.Finish("")
		assert.NoError(t, err)
	})

	t.Run("Wrong status code", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		s := httptest.NewServer(h)
		defer s.Close()

		l := &Launch{
			client: &Client{
				Endpoint: s.URL,
			},
		}
		err := l.Finish("")
		assert.EqualError(t, err, "failed with status 500 Internal Server Error")
	})
}

func TestDeleteLaunch(t *testing.T) {
	t.Run("Successful delete", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test_project/launch/id123", r.URL.Path)
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
			Project:  "test_project",
		}
		l := &Launch{
			client: c,
			Id:     "id123",
		}
		err := l.Delete()
		assert.NoError(t, err)
	})

	t.Run("Wrong status code fail", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
		}
		l := &Launch{
			client: c,
		}
		err := l.Delete()
		assert.EqualError(t, err, "failed with status 500 Internal Server Error")
	})
}

func TestUpdateLaunch(t *testing.T) {
	t.Run("Successful update", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test_project/launch/id123/update", r.URL.Path)
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			d, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, `{"description":"new description","mode":"new mode","tags":["new","tags"]}`, string(d))
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
			Project:  "test_project",
		}
		l := &Launch{
			client: c,
			Id:     "id123",
		}

		err := l.Update("new description", "new mode", []string{"new", "tags"})
		assert.NoError(t, err)
	})

	t.Run("Wrong status code fail", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		s := httptest.NewServer(h)
		defer s.Close()

		c := &Client{
			Endpoint: s.URL,
		}
		l := &Launch{
			client: c,
		}
		err := l.Update("", "", nil)
		assert.EqualError(t, err, "failed with status 500 Internal Server Error")
	})
}
