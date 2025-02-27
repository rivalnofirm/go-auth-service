package rest

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	usecases "go-auth-service/src/app/usecases"
	"go-auth-service/src/infra/config"

	//healthHandler "auth-user-service/src/interface/rest/handlers"
	userHandler "go-auth-service/src/interface/rest/handlers/user"

	"go-auth-service/src/interface/rest/route"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

// HttpServer holds the dependencies for a HTTP server.
type HttpServer struct {
	*http.Server
	logger *logrus.Logger
}

// New creates and configures a server serving all application routes.
// The server implements a graceful shutdown.
// chi.Mux is used for registering some convenient middlewares and easy configuration of
// routes using different http verbs.
func New(
	conf config.HttpConf,
	isProd bool,
	logger *logrus.Logger,
	useCases usecases.AllUseCases,
) (*HttpServer, error) {
	// wrap all the routes
	routeHandler := makeRoute(conf.XRequestID, conf.Timeout, isProd, logger, useCases)

	// http service
	srv := http.Server{
		Addr:    ":" + conf.Port,
		Handler: routeHandler,
	}

	return &HttpServer{&srv, logger}, nil
}

// makeRoute register routes
func makeRoute(
	xRequestID string,
	timeout int,
	isProd bool,
	logger *logrus.Logger,
	useCases usecases.AllUseCases,
) *chi.Mux {

	r := chi.NewRouter()

	// CORS middleware configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // MaxAge untuk OPTIONS preflight request
	}))

	// apply common middleware here ...
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// logging middleware
	r.Use(middleware.Logger)

	// timeout middleware
	if timeout <= 0 {
		logger.Fatalf("invalid http timeout")
	}

	// instantiate the handlers here ...
	uh := userHandler.NewUserHandler(useCases.UserUC)

	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", route.UserRouter(uh))
	})

	return r
}

// Start runs ListenAndServe on the http.Server with graceful shutdown
func (srv *HttpServer) Start(ctx context.Context) {
	// run HTTP service
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srv.logger.Fatal(err)
		}
	}()

	// ready to serve
	srv.logger.Info("listen on", srv.Addr)

	srv.gracefulShutdown(ctx)
}

func (srv *HttpServer) gracefulShutdown(ctx context.Context) {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// wait forever until 1 of signals above are received
	<-quit
	srv.logger.Warnf("got signal: %v, shutting down server ...", quit)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	srv.SetKeepAlivesEnabled(false)
	if err := srv.Shutdown(ctx); err != nil {
		srv.logger.Error(err)
	}

	srv.logger.Println("server exiting")
}
