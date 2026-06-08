package conf

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceConfiguration  ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	StorageConfiguration  StorageConfiguration  `mapstructure:"StorageConfiguration"`
	PostgresConfiguration PostgresConfiguration `mapstructure:"PostgresConfiguration"`
	AgentConfiguration    AgentConfiguration    `mapstructure:"AgentConfiguration"`
	WebConfiguration      WebConfiguration      `mapstructure:"WebConfiguration"`
}

// StorageConfiguration selects the durable backend for the server.
// Backend is "file" (default) or "postgres". For the file backend, control-plane
// state is written to DataFile and run logs stream to LogFile, which is rotated
// once it would exceed LogMaxMB megabytes (the previous segment is dropped).
type StorageConfiguration struct {
	Backend  string `mapstructure:"Backend"`
	DataFile string `mapstructure:"DataFile"`
	LogFile  string `mapstructure:"LogFile"`
	LogMaxMB int    `mapstructure:"LogMaxMB"`
}

type ServiceConfiguration struct {
	Addr  string `mapstructure:"Addr"`
	Debug bool   `mapstructure:"Debug"`
}

type AgentConfiguration struct {
	Token string `mapstructure:"Token"`
}

type WebConfiguration struct {
	Password string `mapstructure:"Password"`
}

type PostgresConfiguration struct {
	Enabled  bool   `mapstructure:"Enabled"`
	Host     string `mapstructure:"Host"`
	Port     int    `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	DBName   string `mapstructure:"DBName"`
	SSLMode  string `mapstructure:"SSLMode"`
	TimeZone string `mapstructure:"TimeZone"`
	LogMode  string `mapstructure:"LogMode"`
}

func Load(configName string, configPaths []string) (Config, error) {
	var cfg Config
	setDefaults(&cfg)

	vp := viper.New()
	vp.SetConfigName(configName)
	vp.SetConfigType("toml")
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.AutomaticEnv()

	for _, path := range configPaths {
		if path != "" {
			vp.AddConfigPath(path)
		}
	}

	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, err
		}
	}
	if err := vp.Unmarshal(&cfg); err != nil {
		return cfg, err
	}
	setDefaults(&cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.ServiceConfiguration.Addr == "" {
		cfg.ServiceConfiguration.Addr = ":8080"
	}
	if cfg.AgentConfiguration.Token == "" {
		cfg.AgentConfiguration.Token = "dev-agent-token"
	}
	if cfg.WebConfiguration.Password == "" {
		cfg.WebConfiguration.Password = "haiwo"
	}
	if cfg.StorageConfiguration.Backend == "" {
		// Backward compatibility: if the legacy Postgres toggle is on and no
		// backend was chosen explicitly, keep using Postgres; otherwise default
		// to the file backend.
		if cfg.PostgresConfiguration.Enabled {
			cfg.StorageConfiguration.Backend = "postgres"
		} else {
			cfg.StorageConfiguration.Backend = "file"
		}
	}
	if cfg.StorageConfiguration.DataFile == "" {
		cfg.StorageConfiguration.DataFile = "./data/haiwo.json"
	}
	if cfg.StorageConfiguration.LogFile == "" {
		cfg.StorageConfiguration.LogFile = "./data/haiwo.log"
	}
	if cfg.StorageConfiguration.LogMaxMB == 0 {
		cfg.StorageConfiguration.LogMaxMB = 50
	}
	if cfg.PostgresConfiguration.Port == 0 {
		cfg.PostgresConfiguration.Port = 5432
	}
	if cfg.PostgresConfiguration.SSLMode == "" {
		cfg.PostgresConfiguration.SSLMode = "disable"
	}
	if cfg.PostgresConfiguration.TimeZone == "" {
		cfg.PostgresConfiguration.TimeZone = "UTC"
	}
	if cfg.PostgresConfiguration.LogMode == "" {
		cfg.PostgresConfiguration.LogMode = "none"
	}
}
