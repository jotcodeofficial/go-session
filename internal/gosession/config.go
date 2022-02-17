package gosession

import (
	"fmt"

	"github.com/spf13/viper"
)

// ConfigApplication - Here is where get the ENV values for the application
type ConfigApplication struct {
	AppName    string `mapstructure:"APP_NAME"`
	AppEnv     string `mapstructure:"APP_ENV"`
	Port       string `mapstructure:"APP_PORT"`
	LogLevel   string `mapstructure:"APP_LOG_LEVEL"`
	DBPort     string `mapstructure:"DATABASE_PORT"`
	DBHost     string `mapstructure:"DATABASE_HOST"`
	DBLogLevel string `mapstructure:"DATABASE_LOG_LEVEL"`
	DBUsername string `mapstructure:"DATABASE_USERNAME"`
	DBPassword string `mapstructure:"DATABASE_PASSWORD"`
	DBName     string `mapstructure:"DATABASE_NAME"`
}

var config ConfigApplication

func startViperConfiguration() {
	// set defaults
	setApplicationDefaults()

	// define configurations
	defineApplicationConfiguration()

	// bind struct values
	bindApplicationEnvStruct()
}

func setApplicationDefaults() {
	viper.SetDefault("APP_NAME", "email-service")
	viper.SetDefault("APP_ENV", "LOCALHOST")
	viper.SetDefault("APP_PORT", "8081")
	viper.SetDefault("APP_LOG_LEVEl", "ERROR")
	viper.SetDefault("DATABASE_PORT", "27017")
	viper.SetDefault("DATABASE_HOST", "localhost")
	viper.SetDefault("DATABASE_LOG_LEVEL", "ERROR")
	viper.SetDefault("DATABASE_USERNAME", "domain")
	viper.SetDefault("DATABASE_PASSWORD", "password")
	viper.SetDefault("DATABASE_NAME", "domain")
}

func defineApplicationConfiguration() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("env")    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./")     // path to look for the config file in. can have multiple lines here to search
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			panic(fmt.Errorf("Fatal error application config file: %s", err))
		} else {
			// Config file was found but another error was produced
			fmt.Println("Config file was found but another error was produced")
			fmt.Println(err)
		}
	}
}

func bindApplicationEnvStruct() {

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Println("unable to decode into struct")
		fmt.Println(err)
	}

}
