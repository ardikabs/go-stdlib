package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func NewConfig(configPath, configName, envPrefix string) *viper.Viper {
	fang := viper.New()

	if envPrefix != "" {
		fang.SetEnvPrefix(envPrefix)
	}

	fang.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	fang.AutomaticEnv()

	fang.SetConfigName(configName)
	fang.AddConfigPath(".")
	fang.AddConfigPath(configPath)

	if err := fang.ReadInConfig(); err != nil { // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	return fang
}

// GetString get integer value in the viper and environment variable with default value
func GetString(viperkey string, env string, defaultVal string) string {
	if value := viper.GetString(viperkey); value != "" {
		return value
	}

	if value := os.Getenv(env); value != "" {
		return value
	}

	return defaultVal
}

// GetInt get integer value in the viper and environment variable with default value
func GetInt(viperkey string, env string, defaultVal int) int {
	if value := viper.Get(viperkey); value != nil {
		switch value.(type) {
		case string:
			if v, err := strconv.Atoi(value.(string)); err == nil {
				return v
			}
		case int:
			return value.(int)
		}
	}

	if value := os.Getenv(env); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultVal
}

// GetBool get bool value in the viper and environment variable with default value
func GetBool(viperkey string, env string, defaultVal bool) bool {
	if value := viper.GetString(viperkey); value != "" {
		return viper.GetBool(viperkey)
	}

	value := os.Getenv(env)
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return defaultVal
	}

	return boolVal
}

// GetStringFromBase64Encoded get string from base64 encoded value in the viper and environment variable
func GetStringFromBase64Encoded(viperkey string, env string) string {
	value := viper.GetString(viperkey)
	if value == "" {
		value = os.Getenv(env)
	}

	content, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return ""
	}

	return string(content)
}
