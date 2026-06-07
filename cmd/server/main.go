package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/haiwo-ci/haiwo/internal/conf"
	"github.com/haiwo-ci/haiwo/internal/database"
	"github.com/haiwo-ci/haiwo/internal/server"
	"gorm.io/gorm"
)

func main() {
	configName := flag.String("c", "config", "Name of the config file, without extension")
	configDirs := flag.String("cPath", "./,./configs/", "Directories to search for config file, separated by ','")
	flag.Parse()

	appConfig, err := conf.Load(*configName, strings.Split(*configDirs, ","))
	if err != nil {
		log.Fatal(err)
	}

	var db *gorm.DB
	if appConfig.PostgresConfiguration.Enabled {
		var err error
		db, err = database.OpenPostgres(appConfig.PostgresConfiguration)
		if err != nil {
			log.Fatal(err)
		}
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}
		log.Println("database migrated")
	}

	cfg := server.Config{
		Addr:        appConfig.ServiceConfiguration.Addr,
		AgentToken:  appConfig.AgentConfiguration.Token,
		WebPassword: appConfig.WebConfiguration.Password,
		DB:          db,
		Now:         time.Now,
	}

	app := server.New(cfg)
	log.Printf("haiwo server listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, app.Routes()); err != nil {
		log.Fatal(err)
	}
}
