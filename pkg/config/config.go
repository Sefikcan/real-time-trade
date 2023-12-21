package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	defaultConfigPath     = "pkg/config/"
	tagName               = "mapstructure"
	configFileType        = "yaml"
	defaultConfigFileName = "config-dev"
	environmentKey        = "environment"
)

type Config struct {
	Server  ServerConfig `mapstructure:"server"`
	Metric  MetricConfig `mapstructure:"metric"`
	Logger  LoggerConfig `mapstructure:"logger"`
	Jaeger  JaegerConfig `mapstructure:"jaeger"`
	Kafka   KafkaConfig  `mapstructure:"kafka"`
	Tickers TickerConfig `mapstructure:"tickers"`
}

type ServerConfig struct {
	AppVersion     string        `mapstructure:"appVersion"`
	Host           string        `mapstructure:"host"`
	Port           string        `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	SSL            bool          `mapstructure:"ssl"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
	CtxTimeout     time.Duration `mapstructure:"ctxTimeout"`
}

type MetricConfig struct {
	Url         string `mapstructure:"url"`
	ServiceName string `mapstructure:"serviceName"`
}

type LoggerConfig struct {
	Development bool   `mapstructure:"development"`
	Encoding    string `mapstructure:"encoding"`
	Level       string `mapstructure:"level"`
}

type JaegerConfig struct {
	Host        string `mapstructure:"host"`
	ServiceName string `mapstructure:"serviceName"`
	LogSpans    bool   `mapstructure:"logSpans"`
}

type TickerConfig struct {
	Tickers string `mapstructure:"tickers"`
}

type KafkaConfig struct {
	Brokers           []string `mapstructure:"brokers"`
	GroupID           string   `mapstructure:"groupID"`
	InitTopics        bool     `mapstructure:"initTopics"`
	TopicName         string   `mapstructure:"topicName"`
	Partitions        int      `mapstructure:"partitions"`
	ReplicationFactor int      `mapstructure:"replicationFactor"`
}

func NewConfig() *Config {
	env, _ := os.LookupEnv(environmentKey)
	fmt.Println("Environment: [" + env + "] read from runtime arguments [" + environmentKey + "].")

	return ReadConfig(&Config{}, strings.ToUpper(env))
}

var ReadConfig = func(c *Config, env string) *Config {
	fmt.Println("Configuration read initiated...")
	v := viper.New()
	switch {
	case env == "DEV":
		v = readFromEnv(v)
	case env == "REMOTE":
		v = readFromConfigServer(v)
	default:
		v = readFromAppYaml(v)
	}
	if err := v.Unmarshal(&c); err != nil {
		panic("Configuration unmarshalling occurred an error, terminating: " + err.Error())
	}

	return c
}

var readFromEnv = func(v *viper.Viper) *viper.Viper {
	fmt.Println("Reading environment configuration")
	addKeysToViper(v)
	v.AutomaticEnv()
	return v
}

var readFromConfigServer = func(v *viper.Viper) *viper.Viper {
	return v
}

var readFromAppYaml = func(v *viper.Viper) *viper.Viper {
	fmt.Println("Reading application yml configuration")
	v.SetConfigName(defaultConfigFileName)
	v.SetTypeByDefaultValue(true)
	v.SetConfigType(configFileType)
	v.AddConfigPath("./" + defaultConfigPath)
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Viper read config has an error: %e\n", err)
	}

	return v
}

func getAllKeys(t reflect.Type) []string {
	var result []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n := strings.ToUpper(f.Tag.Get(tagName))
		if reflect.Struct == f.Type.Kind() {
			subKeys := getAllKeys(f.Type)
			for _, k := range subKeys {
				result = append(result, n+"."+k)
			}
		} else {
			result = append(result, n)
		}
	}

	return result
}

func addKeysToViper(v *viper.Viper) {
	var reply interface{} = Config{}
	t := reflect.TypeOf(reply)
	keys := getAllKeys(t)
	for _, key := range keys {
		v.SetDefault(key, "")
	}
}
