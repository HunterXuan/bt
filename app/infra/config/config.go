package config

import (
	"github.com/spf13/viper"
	"log"
)

var Config *viper.Viper

func InitConfig() {
	log.Println("Config Initializing...")

	Config = viper.New()
	Config.AutomaticEnv()

	log.Println("Config Initialized!")
}
