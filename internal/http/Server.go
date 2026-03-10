package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

type HttpServer struct {
	Address   string
	Server    *http.Server
	Database  *gorm.DB
	ExtrasDir string
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var method string
		if strings.Contains(r.URL.Path, "/api/new/") {
			method = "POST"
		} else {
			method = r.Method
		}
		log.Printf("%s %s from %s",
			method,
			r.URL.Path,
			r.RemoteAddr,
		)
		next.ServeHTTP(w, r)
	})
}

func (h *HttpServer) Start() error {

	log.Printf("Starting http listening on %s", h.Address)

	router := httprouter.New()

	// iPXE Client
	router.GET("/api/boot/:mac", h.BootScript)
	router.GET("/api/new/host/:mac/:hostname", h.NewHostiPXE)

	// New Object
	router.POST("/api/new/host", h.NewHost)
	router.POST("/api/new/task", h.NewTask)

	// Update Object
	router.POST("/api/edit/host/:id", h.EditHost)
	router.POST("/api/edit/task/:id", h.EditTask)

	// Delete Object
	router.POST("/api/delete/host/:id", h.DeleteHost)
	router.POST("/api/delete/task/:id", h.DeleteTask)

	// UI
	router.GET("/", h.UI)
	router.GET("/hosts", h.UI)
	router.GET("/hosts/new", h.UI)
	router.GET("/hosts/edit/:id", h.UI)
	router.GET("/tasks", h.UI)
	router.GET("/tasks/new", h.UI)
	router.GET("/tasks/edit/:id", h.UI)

	// UI Login
	router.GET("/login", h.UI)
	router.POST("/login", h.LoginPost)

	// User Extras
	router.ServeFiles("/extras/*filepath", http.Dir(h.ExtrasDir))

	h.Server = &http.Server{
		Addr:    h.Address,
		Handler: loggingMiddleware(router),
	}

	errChan := make(chan error, 1)
	go func() {
		if err := h.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		log.Printf("Started http")
		return nil
	}
}

func (h *HttpServer) Stop() error {
	if h.Server == nil {
		return fmt.Errorf("server not running")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return h.Server.Shutdown(ctx)
}
