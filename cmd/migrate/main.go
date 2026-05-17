package main

import (
	"flag"
	"log"
	"strings"

	"github.com/haiwo-ci/haiwo/internal/conf"
	"github.com/haiwo-ci/haiwo/internal/database"
)

func main() {
	configName := flag.String("c", "config", "Name of the config file, without extension")
	configDirs := flag.String("cPath", "./,./configs/", "Directories to search for config file, separated by ','")
	flag.Parse()

	cfg, err := conf.Load(*configName, strings.Split(*configDirs, ","))
	if err != nil {
		log.Fatal(err)
	}
	if !cfg.PostgresConfiguration.Enabled {
		log.Fatal("PostgresConfiguration.Enabled must be true to run migrations")
	}

	db, err := database.OpenPostgres(cfg.PostgresConfiguration)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	log.Println("database migrated")
}
