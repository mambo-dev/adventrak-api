package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mambo-dev/adventrak-backend/internal/database"
)

type apiConfig struct {
	db             *database.Queries
	jwtSecret      string
	sendGridApiKey string
	frontEndURL    string
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

	router := chi.NewRouter()
	allowedOrigins := []string{"http://*"}

	if workEnv != "dev" {
		allowedOrigins = []string{"https://*"}
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()

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
	}

	if workEnv == "dev" {
		v1Router.Delete("/admin/reset", apiCfg.resetDatabase)
	}

	v1Router.Get("/healthz", handlerReadiness)

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 300,
	}

	log.Printf("Serving on: http://localhost:%s\n", port)
	log.Fatal(srv.ListenAndServe())

}
