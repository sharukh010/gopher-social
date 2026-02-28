package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/sharukh010/social/internal/db"
	"github.com/sharukh010/social/internal/env"
	"github.com/sharukh010/social/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GopherSocial API
//	@description	This is a Social API.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// configurations
	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("API_URL", "localhost"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	//database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	store := store.NewStorage(db)

	api := &application{
		config: cfg,
		store:  store,
		logger: logger,
	}

	mux := api.mount()
	if err := api.run(mux); err != nil {
		logger.Fatal("Error occured: %v\n", err)
	}
}
