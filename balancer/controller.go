package balancer

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Authentication struct {
	Username string
	Password string
}

type Controller struct {
	Address string
	Auth    *Authentication
}

func (c *Controller) Run(service *Service) {
	mux := http.NewServeMux()
	mux.Handle("/stats", &StatsHandler{service})
	mux.Handle("/virtualserver", &VirtualServerHandler{service})
	http.ListenAndServe(c.Address, AuthMiddleware(c.Auth)(mux))
}

func AuthMiddleware(auth *Authentication) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			username := r.Header.Get("Username")
			password := r.Header.Get("Password")
			if username == auth.Username && password == auth.Password {
				next.ServeHTTP(w, r)
			} else {
				log.Errorf("Unauthorized <%s, %s> from %s", username, password, r.RemoteAddr)
				WriteError(w, ErrUnauthorized)
				return
			}
		}
		return http.HandlerFunc(fn)
	}
}

type StatsHandler struct {
	service *Service
}

func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result := []string{}
	for _, vs := range h.service.VServers {
		s := fmt.Sprintf("pool-%s:\n%s", vs.Name, vs.stats)
		log.Infof(s)
		result = append(result, s)
	}
	result = append(result, "\n")
	io.WriteString(w, strings.Join(result, "\n"))
}

type VirtualServerHandler struct {
	service *Service
}

func (h *VirtualServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, world!\n")
}