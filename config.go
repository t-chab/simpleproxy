package main

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"os/user"
	"path"
)

type ProxyConfig struct {
	proxyLogin      string
	proxyPassword   string
	listenAddress   string
	listenPort      int
	targetProxyHost string
	targetProxyPort int
	logVerbose      bool
}

const (
	AppName        string = "simple-proxy"
	ConfigFileName        = AppName + ".yml"
	LOGIN                 = "proxyLogin"
	PASSWORD              = "proxyPassword"
	ListenAddr            = "listenAddress"
	ListenPort            = "listenPort"
	ProxyHost             = "targetProxyHost"
	ProxyPort             = "targetProxyPort"
	VerboseLog            = "logVerbose"
)

func getHomeDirectory() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Error getting current user :", err)
	}
	return usr.HomeDir
}

func getConfigPath() string {
	val, ok := os.LookupEnv("XDG_CONFIG_DIR")
	if ok {
		return path.Join(val, AppName)
	}
	return path.Join(getHomeDirectory(), ".config", AppName)
}

func getConfigFilePath() string {
	return path.Join(getConfigPath(), ConfigFileName)
}

func getProxyCredentials() (string, string) {
	return viper.GetString("proxyLogin"), viper.GetString("proxyPassword")
}

func getProxyConfig() ProxyConfig {
	loadConfiguration()
	return ProxyConfig{
		proxyLogin:      viper.GetString(LOGIN),
		proxyPassword:   viper.GetString(PASSWORD),
		listenAddress:   viper.GetString(ListenAddr),
		listenPort:      viper.GetInt(ListenPort),
		targetProxyHost: viper.GetString(ProxyHost),
		targetProxyPort: viper.GetInt(ProxyPort),
		logVerbose:      viper.GetBool(VerboseLog),
	}
}

func loadConfiguration() {
	viper.SetConfigType("yml")
	viper.SetDefault(ListenAddr, "127.0.0.1")
	viper.SetDefault(ListenPort, 8118)
	viper.SetDefault(ProxyHost, "")
	viper.SetDefault(ProxyPort, 8000)
	viper.SetDefault(LOGIN, "")
	viper.SetDefault(PASSWORD, "")
	viper.SetDefault(VerboseLog, false)
	viper.AddConfigPath(getConfigPath())
	viper.SetConfigName(AppName)
	if err := viper.ReadInConfig(); err != nil { // Find and read the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Printf("Configuration file not found in %s, a new file will be created.", getConfigFilePath())
			errDir := os.MkdirAll(getConfigPath(), os.ModePerm)
			if errDir != nil {
				log.Fatalf("Failed to create %s", getConfigPath())
			}
			_, errFile := os.Create(getConfigFilePath())
			if errFile != nil {
				log.Fatalf("Failed to create %s", getConfigFilePath())
			}
			confError := viper.WriteConfigAs(getConfigFilePath())
			if confError != nil {
				log.Fatalf("Can't write configuration file : %s", confError)
			}
		} else {
			// Config file was found but another error was produced
			log.Fatal("Fatal error when reading configuration file", err)
		}
	}
}
