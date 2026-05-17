package conf

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceConfiguration  ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	PostgresConfiguration PostgresConfiguration `mapstructure:"PostgresConfiguration"`
	AgentConfiguration    AgentConfiguration    `mapstructure:"AgentConfiguration"`
	WebConfiguration      WebConfiguration      `mapstructure:"WebConfiguration"`
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
