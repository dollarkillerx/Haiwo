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
)

func main() {
	configName := flag.String("c", "config", "Name of the config file, without extension")
	configDirs := flag.String("cPath", "./,./configs/", "Directories to search for config file, separated by ','")
	flag.Parse()

	appConfig, err := conf.Load(*configName, strings.Split(*configDirs, ","))
	if err != nil {
		log.Fatal(err)
	}

	var storage server.StorageOptions
	switch appConfig.StorageConfiguration.Backend {
	case "postgres":
		db, err := database.OpenPostgres(appConfig.PostgresConfiguration)
		if err != nil {
			log.Fatal(err)
		}
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}
		log.Println("storage backend: postgres (migrated)")
		storage.DB = db
	case "file":
		storage.DataFile = appConfig.StorageConfiguration.DataFile
		storage.LogFile = appConfig.StorageConfiguration.LogFile
		storage.LogMaxBytes = int64(appConfig.StorageConfiguration.LogMaxMB) * 1024 * 1024
		log.Printf("storage backend: file (state=%s, logs=%s, logMaxMB=%d)",
			storage.DataFile, storage.LogFile, appConfig.StorageConfiguration.LogMaxMB)
	default:
		log.Fatalf("unknown storage backend %q (expected \"file\" or \"postgres\")", appConfig.StorageConfiguration.Backend)
	}

	cfg := server.Config{
		Addr:        appConfig.ServiceConfiguration.Addr,
		AgentToken:  appConfig.AgentConfiguration.Token,
		WebPassword: appConfig.WebConfiguration.Password,
		Storage:     storage,
		Now:         time.Now,
	}

	app := server.New(cfg)
	log.Printf("haiwo server listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, app.Routes()); err != nil {
		log.Fatal(err)
	}
}
