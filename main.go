package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Shaheryarkhalid/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


type ApiConfig struct {
	db *database.Queries
	tokenSecret string
	Platform string
	polkaKey string
	fileserverHits atomic.Int32
}

func main(){
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error: Trying to load environment variables.")
		fmt.Println(err)
		os.Exit(1)
	}
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == ""{
		fmt.Println("Database Url not found in .env")
		os.Exit(1)
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		platform = "prod"
	}
	tokenSecret := os.Getenv("TOKEN_SECRET")
	if tokenSecret == "" {
		fmt.Println("TOKEN_SECRET not found in environment.")
		os.Exit(1)
	}

	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		fmt.Println("POLKA_KEY not found in environment.")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Println("Error: Trying to open db connection.")
		os.Exit(1)
	}
	cnfg := ApiConfig{
		db: database.New(db),
		Platform: platform,
		tokenSecret: tokenSecret,
		polkaKey: polkaKey,
	}
	serverMultiplexer := http.NewServeMux() 
	server := http.Server{
		Handler: serverMultiplexer,
		Addr: ":8080",
		MaxHeaderBytes: 1<<20,
	}
	handler := NewHandler(&cnfg)

	serverMultiplexer.Handle("/app/",  middlewareMetricsInc(&cnfg, http.StripPrefix("/app", http.FileServer(http.Dir(".")) )))

	serverMultiplexer.HandleFunc("GET /api/healthz", handler.handlerHealthz )
	serverMultiplexer.HandleFunc("GET /api/crash", handler.handlerCrash )

	serverMultiplexer.HandleFunc("POST /api/users", handler.handlerCreateUser)
	serverMultiplexer.HandleFunc("POST /api/login", handler.handlerLogin)
	serverMultiplexer.HandleFunc("POST /api/refresh", handler.handlerRefresh)
	serverMultiplexer.HandleFunc("POST /api/revoke", handler.handlerRevoke)

	serverMultiplexer.HandleFunc("PUT /api/users", handler.handlerUpdateUser)
	serverMultiplexer.HandleFunc("POST /api/polka/webhooks", handler.handlerUpgradeUserToRed)


	serverMultiplexer.HandleFunc("POST /api/chirps", handler.handlerCreateChirp)
	serverMultiplexer.HandleFunc("GET /api/chirps", handler.handlerGetChirps)
	serverMultiplexer.HandleFunc("GET /api/chirps/{chirpID}", handler.handlerGetOneChirp)
	serverMultiplexer.HandleFunc("DELETE /api/chirps/{chirpID}", handler.handlerDeleteChirp)


	serverMultiplexer.HandleFunc("GET /admin/metrics", handler.handlerMetrics )
	serverMultiplexer.HandleFunc("POST /admin/reset", handler.handlerReset)
	log.Fatal(server.ListenAndServe())
}
