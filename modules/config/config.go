package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.oease.dev/goe/utils"
	"os"
	"strconv"
	"strings"
)

const (
	defaultFileName         = "/.env"
	defaultOverrideFileName = "/.local.env"
)

type Config struct {
}

func New(folder string) *Config {
	c := &Config{}
	c.read(folder)
	return c
}

func (c *Config) read(folder string) {
	var (
		defaultFile  = folder + defaultFileName
		overrideFile = folder + defaultOverrideFileName
		env          = c.Get("APP_ENV")
	)

	err := godotenv.Load(defaultFile)
	if err != nil {
		fmt.Printf("Failed to load config from file: %v, Err: %v \n", defaultFile, err)
	} else {
		fmt.Printf("Loaded config from file: %v \n", defaultFile)
	}

	switch env {
	case "":
		// If 'APP_ENV' is not set, then Goe will read '.env' from configs directory, and then it will be overwritten
		// by configs present in file '.local.env'
		err = godotenv.Overload(overrideFile)
		if err != nil {
			fmt.Printf("Failed to load config from file: %v, Err: %v \n", overrideFile, err)
		} else {
			fmt.Printf("Loaded config from file: %v \n", overrideFile)
		}

	default:
		// If 'APP_ENV' is set to x, then GoFr will read '.env' from configs directory, and then it will be overwritten
		// by configs present in file '.x.env'
		overrideFile = fmt.Sprintf("%s/.%s.env \n", folder, env)

		err = godotenv.Overload(overrideFile)
		if err != nil {
			fmt.Printf("Failed to load config from file: %v, Err: %v \n", overrideFile, err)
		} else {
			fmt.Printf("Loaded config from file: %v \n", overrideFile)
		}
	}
}

func (c *Config) Get(key string) string {
	return os.Getenv(key)
}

func (c *Config) GetString(key string) string {
	return c.Get(key)
}

func (c *Config) GetInt(key string) int {
	val, err := utils.Convert(c.GetString(key), strconv.Atoi, 0)
	if err != nil {
		return 0
	}
	return *val
}

func (c *Config) GetBool(key string) bool {
	val, err := utils.Convert(c.GetString(key), strconv.ParseBool, false)
	if err != nil {
		return false
	}
	return *val
}

func (c *Config) GetStringSlice(key string) []string {
	str := c.GetString(key)
	if str == "" {
		return nil
	}
	strArr := strings.Split(str, ",")
	for i, s := range strArr {
		strArr[i] = strings.TrimSpace(s)
	}
	return strArr
}

func (c *Config) GetIntSlice(key string) []int {
	str := c.GetString(key)
	if str == "" {
		return nil
	}
	strArr := strings.Split(str, ",")
	intArr := make([]int, len(strArr))
	for i, s := range strArr {
		val, err := utils.Convert(s, strconv.Atoi, 0)
		if err != nil {
			intArr[i] = 0
		} else {
			intArr[i] = *val
		}
	}
	return intArr
}

func (c *Config) GetBoolSlice(key string) []bool {
	str := c.GetString(key)
	if str == "" {
		return nil
	}
	strArr := strings.Split(str, ",")
	boolArr := make([]bool, len(strArr))
	for i, s := range strArr {
		val, err := utils.Convert(s, strconv.ParseBool, false)
		if err != nil {
			boolArr[i] = false
		} else {
			boolArr[i] = *val
		}
	}
	return boolArr
}

func (c *Config) GetOrDefaultString(key string, defaultValue string) string {
	if val := c.GetString(key); val != "" {
		return val
	}
	return defaultValue
}

func (c *Config) GetOrDefaultInt(key string, defaultValue int) int {
	val, err := utils.Convert(c.GetString(key), strconv.Atoi, defaultValue)
	if err != nil {
		return defaultValue
	}
	if val != nil || *val != 0 {
		return *val
	}
	return defaultValue
}

func (c *Config) GetOrDefaultBool(key string, defaultValue bool) bool {
	val, err := utils.Convert(c.GetString(key), strconv.ParseBool, defaultValue)
	if err != nil {
		return defaultValue
	}
	if val != nil {
		return *val
	}
	return defaultValue
}
