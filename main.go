package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mambo-dev/adventrak-backend/internal/database"
	"github.com/mambo-dev/adventrak-backend/internal/utils"
)

type apiConfig struct {
	db             *database.Queries
	jwtSecret      string
	sendGridApiKey string
	frontEndURL    string
	assetsRoot     string
	baseApiUrl     string
}

func main() {
	go cleanUpOldTimers()

	err := godotenv.Load(".env")

	if err != nil {
		log.Printf("WARNING: assuming default configuration. .env unreadable: %v", err)
	}

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("FATAL: PORT environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set")
	}

	workEnv := os.Getenv("WORKENV")

	if workEnv == "" {
		log.Fatal("FATAL: working environment variable is not set")
	}

	sendGridApiKey := os.Getenv("SENDGRID_API_KEY")

	if sendGridApiKey == "" {
		log.Fatal("FATAL: sendgrid api  not set")
	}

	frontEndURL := os.Getenv("BASE_FRONTEND_URL")

	if frontEndURL == "" {
		log.Fatal("FATAL: frontend url not set")
	}

	assetsRoot := os.Getenv("ASSETS_ROOT")
	if assetsRoot == "" {
		log.Fatal("ASSETS_ROOT environment variable is not set")
	}

	baseApiUrl := os.Getenv("BASE_API_URL")
	if baseApiUrl == "" {
		log.Fatal("ASSETS_ROOT environment variable is not set")
	}

	apiCfg := apiConfig{}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatalf("could not open db:%v\n", err.Error())
	}

	apiCfg.db = database.New(db)
	apiCfg.jwtSecret = jwtSecret
	apiCfg.sendGridApiKey = sendGridApiKey
	apiCfg.frontEndURL = frontEndURL
	apiCfg.assetsRoot = assetsRoot
	apiCfg.baseApiUrl = baseApiUrl

	router := chi.NewRouter()
	allowedOrigins := []string{"http://*"}

	if workEnv != "dev" {
		allowedOrigins = []string{"https://*"}
	}

	if workEnv == "dev" {
		apiCfg.baseApiUrl = fmt.Sprintf("%v%v", baseApiUrl, port)
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	assetsHandler := http.StripPrefix("/assets", http.FileServer(http.Dir(assetsRoot)))
	v1Router := chi.NewRouter()

	err = utils.EnsureAssetsDir(apiCfg.assetsRoot)

	if err != nil {
		log.Fatalf("Couldn't create assets directory: %v", err)
	}

	if apiCfg.db != nil {
		log.Println("Db is active")
		v1Router.Post("/auth/signup", apiCfg.handlerSignup)
		v1Router.Post("/auth/login", apiCfg.handlerLogin)
		v1Router.Post("/auth/refresh", apiCfg.handlerRefresh)
		v1Router.Get("/auth/send-verification",
			apiCfg.UseAuth(apiCfg.handlerSendVerification))
		v1Router.Put("/auth/verify-email", apiCfg.UseAuth(http.HandlerFunc(apiCfg.handlerVerifyEmail)))
		v1Router.Get("/auth/request-password-reset", apiCfg.handlerResetRequest)
		v1Router.Put("/auth/reset-password", apiCfg.handlerResetPassword)
		v1Router.Post("/auth/logout", apiCfg.UseAuth(http.HandlerFunc(apiCfg.handlerLogout)))

		v1Router.Get("/trips", apiCfg.UseAuth(apiCfg.handlerGetTrips))
		v1Router.Get("/trips/{tripID}", apiCfg.UseAuth(apiCfg.handlerGetTrip))
		v1Router.Post("/trips", apiCfg.UseAuth(apiCfg.handlerCreateTrip))
		v1Router.Put("/trips/{tripID}", apiCfg.UseAuth(apiCfg.handlerUpdateTripDetails))
		v1Router.Patch("/trips/{tripID}/end", apiCfg.UseAuth(apiCfg.handlerMarkTripComplete))
		v1Router.Delete("/trips/{tripID}", apiCfg.UseAuth(apiCfg.handlerDeleteTrip))

		v1Router.Get("/stops", apiCfg.UseAuth(apiCfg.handlerGetStops))
		v1Router.Get("/stops/{stopID}", apiCfg.UseAuth(apiCfg.handlerGetStop))
		v1Router.Post("/stops/{tripID}", apiCfg.UseAuth(apiCfg.handlerCreateStop))
		v1Router.Put("/stops/{stopID}", apiCfg.UseAuth(apiCfg.handlerUpdateStop))
		v1Router.Delete("/stops/{stopID}", apiCfg.UseAuth(apiCfg.handlerDeleteStop))

		v1Router.Post("/media/photos", apiCfg.UseAuth(apiCfg.handlerUploadPhotos))

	}

	if workEnv == "dev" {
		v1Router.Delete("/admin/reset", apiCfg.resetDatabase)
	}

	v1Router.Get("/healthz", handlerReadiness)

	router.Handle("/assets/*", assetsHandler)
	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 300,
	}

	log.Printf("Serving on: http://localhost:%s\n", port)
	log.Fatal(srv.ListenAndServe())

}
