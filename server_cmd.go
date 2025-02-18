package server

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/resolvers"
	"github.com/interline-io/transitland-server/rest"
)

type Command struct {
	DisableGraphql   bool
	DisableRest      bool
	EnablePlayground bool
	EnableProfiler   bool
	config.Config
}

func (cmd *Command) Parse(args []string) error {
	fl := flag.NewFlagSet("sync", flag.ExitOnError)
	fl.Usage = func() {
		log.Print("Usage: server")
		fl.PrintDefaults()
	}
	fl.StringVar(&cmd.DBURL, "dburl", "", "Database URL (default: $TL_DATABASE_URL)")
	fl.IntVar(&cmd.Timeout, "timeout", 60, "")
	fl.StringVar(&cmd.Port, "port", "8080", "")
	fl.StringVar(&cmd.JwtAudience, "jwt-audience", "", "JWT Audience")
	fl.StringVar(&cmd.JwtIssuer, "jwt-issuer", "", "JWT Issuer")
	fl.StringVar(&cmd.JwtPublicKeyFile, "jwt-public-key-file", "", "Path to JWT public key file")
	fl.StringVar(&cmd.UseAuth, "auth", "", "")
	fl.StringVar(&cmd.GtfsDir, "gtfsdir", "", "Directory to store GTFS files")
	fl.StringVar(&cmd.GtfsS3Bucket, "s3", "", "S3 bucket for GTFS files")
	fl.StringVar(&cmd.RestPrefix, "rest-prefix", "", "REST prefix for generating pagination links")
	fl.BoolVar(&cmd.ValidateLargeFiles, "validate-large-files", false, "Allow validation of large files")
	fl.BoolVar(&cmd.DisableImage, "disable-image", false, "Disable image generation")
	fl.BoolVar(&cmd.DisableGraphql, "disable-graphql", false, "Disable GraphQL endpoint")
	fl.BoolVar(&cmd.DisableRest, "disable-rest", false, "Disable REST endpoint")
	fl.BoolVar(&cmd.EnablePlayground, "playground", false, "Enable GraphQL playground")
	fl.BoolVar(&cmd.EnableProfiler, "profile", false, "Enable profiling")
	fl.Parse(args)
	if cmd.DBURL == "" {
		cmd.DBURL = os.Getenv("TL_DATABASE_URL")
	}
	return nil
}

func (cmd *Command) Run() error {
	// Open database
	model.DB = model.MustOpenDB(cmd.DBURL)

	// Setup CORS and logging
	root := mux.NewRouter()
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type", "apikey", "authorization"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	root.Use(cors)
	root.Use(loggingMiddleware)

	// Profiling
	if cmd.EnableProfiler {
		root.HandleFunc("/debug/pprof/", pprof.Index)
		root.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		root.HandleFunc("/debug/pprof/profile", pprof.Profile)
		root.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	// Add servers
	graphqlServer, err := resolvers.NewServer(cmd.Config)
	if err != nil {
		return err
	}
	if !cmd.DisableGraphql {
		mount(root, "/query", graphqlServer)
	}
	if !cmd.DisableRest {
		restServer, err := rest.NewServer(cmd.Config, graphqlServer)
		if err != nil {
			return err
		}
		mount(root, "/rest", restServer)
	}
	if cmd.EnablePlayground && !cmd.DisableGraphql {
		root.Handle("/", playground.Handler("GraphQL playground", "/query/"))
	}
	// Start server
	addr := fmt.Sprintf("%s:%s", "0.0.0.0", cmd.Port)
	fmt.Println("listening on:", addr)
	timeOut := time.Duration(cmd.Timeout)
	srv := &http.Server{
		Handler:      root,
		Addr:         addr,
		WriteTimeout: timeOut * time.Second,
		ReadTimeout:  timeOut * time.Second,
	}
	return srv.ListenAndServe()

}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If requesting /query rewrite to /query/ to match subrouter's "/"
		if r.URL.Path == path {
			r.URL.Path = r.URL.Path + "/"
		}
		// Remove path prefix
		r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
		handler.ServeHTTP(w, r)
	}))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
