package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/InVisionApp/go-health/v2"
	"github.com/InVisionApp/go-health/v2/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	cli "github.com/jawher/mow.cli"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	Port string `mapstructure:"PORT"`
}

type Application struct {
	Name        string
	Description string
	Cli         *cli.Cli
	Setup       func(ctx context.Context)
	Config      Config
	Router      *chi.Mux
	Health      *health.Health
}

func NewApplication(name string, description string) *Application {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"WWW-Authenticate", "Content-Type", "Accept", "Host", "Date"},
	}))

	r.Use(middleware.Timeout(60 * time.Second))
	r.Mount("/debug", middleware.Profiler())

	var conf Config
	app := &Application{
		Name:        strings.TrimPrefix(name, "./"),
		Description: description,
		Cli:         cli.App(name, description),
		Setup:       func(ctx context.Context) {},
		Config:      conf,
		Router:      r,
		Health:      health.New(),
	}

	app.Cli.Action = func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
		err = viper.Unmarshal(&app.Config)

		// TODO: signal handling and graceful shutdown
		// to nie moje TODO, bylo w oryginale, ale nie wiem jak to naprawic i co tu zrobic...
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		app.Setup(ctx)

		if err := app.Health.Start(); err != nil {
			log.Fatalf("healthcheck: failed to start healthcheck: %v", err)
		}

		app.Router.Get("/healthz", handlers.NewJSONHandlerFunc(app.Health, nil))
		app.Router.Get("/about", func(w http.ResponseWriter, r *http.Request) {
			res := map[string]interface{}{
				"name":        app.Name,
				"description": app.Description,
			}
			w.Header().Set("Content-type", "application/json")
			json.NewEncoder(w).Encode(&res)
		})

		//staticRootDir := "templates/static/"
		//fs := http.FileServer(http.Dir(staticRootDir))
		//
		//app.Router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		//	if _, err := os.Stat(staticRootDir + r.RequestURI); os.IsNotExist(err) {
		//		http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		//	} else {
		//		fs.ServeHTTP(w, r)
		//	}
		//})

		log.Printf("%s http server listen start on %s", app.Name, app.Config.Port)
		log.Fatal(http.ListenAndServe(app.Config.Port, app.Router))
	}

	return app
}

func (app *Application) Run(params []string) error {
	return app.Cli.Run(params)
}
