package config

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ProxyConfig struct {
	ProxyLogin       string
	ProxyPassword    string
	ListenAddress    string
	ListenPort       int
	TargetProxyHost  string
	TargetProxyPort  int
	LogVerbose       bool
	ForwardingStatus bool
}

const (
	AppName          string = "simpleproxy"
	CfgFileName             = AppName + ".yml"
	LOGIN                   = "proxyLogin"
	PASSWORD                = "proxyPassword"
	ListenAddr              = "listenAddress"
	ListenPort              = "listenPort"
	ProxyHost               = "targetProxyHost"
	ProxyPort               = "targetProxyPort"
	VerboseLog              = "logVerbose"
	ForwardingStatus        = "ForwardingStatus"
)

type DefaultValues struct {
	listenAddr string
	listenPort int
	proxyHost  string
	proxyPort  int
	login      string
	password   string
}

func NewDefaultValues() DefaultValues {
	defaultValues := DefaultValues{}
	defaultValues.listenAddr = "127.0.0.1"
	defaultValues.listenPort = 8118
	defaultValues.proxyHost = ""
	defaultValues.proxyPort = 8000
	defaultValues.login = ""
	defaultValues.password = ""

	return defaultValues
}

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

func GetConfigFilePath() string {
	return path.Join(getConfigPath(), CfgFileName)
}

func GetProxyConfig() ProxyConfig {
	loadConfiguration(NewDefaultValues())
	return ProxyConfig{
		ProxyLogin:       viper.GetString(LOGIN),
		ProxyPassword:    viper.GetString(PASSWORD),
		ListenAddress:    viper.GetString(ListenAddr),
		ListenPort:       viper.GetInt(ListenPort),
		TargetProxyHost:  viper.GetString(ProxyHost),
		TargetProxyPort:  viper.GetInt(ProxyPort),
		ForwardingStatus: viper.GetBool(ForwardingStatus),
		LogVerbose:       viper.GetBool(VerboseLog),
	}
}

func loadConfiguration(values DefaultValues) {
	envPrefix := strings.ReplaceAll(AppName, "-", "_")
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	viper.SetDefault(ListenAddr, values.listenAddr)
	viper.SetDefault(ListenPort, values.listenPort)
	viper.SetDefault(ProxyHost, values.proxyHost)
	viper.SetDefault(ProxyPort, values.proxyPort)
	viper.SetDefault(LOGIN, values.login)
	viper.SetDefault(PASSWORD, values.password)
	viper.SetDefault(VerboseLog, false)
	viper.AddConfigPath(getConfigPath())
	viper.SetConfigName(AppName)
	viper.SetEnvPrefix(envPrefix)

	if err := viper.ReadInConfig(); err != nil { // Find and read the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Configuration file not found in %s, a new file will be created.", GetConfigFilePath())

			// Config file not found; ignore error if desired
			errDir := os.MkdirAll(getConfigPath(), os.ModePerm)
			if errDir != nil {
				log.Fatalf("Failed to create %s", getConfigPath())
			}

			_, errFile := os.Create(GetConfigFilePath())
			if errFile != nil {
				log.Fatalf("Failed to create %s", GetConfigFilePath())
			}

			confError := viper.WriteConfigAs(GetConfigFilePath())
			if confError != nil {
				log.Fatalf("Can't write configuration file : %s", confError)
			}
		} else {
			// Config file was found but another error was produced
			log.Fatal("Fatal error when reading configuration file", err)
		}
	}
}

// CmdLineFlags Read command line flags and override current config if needed.
func CmdLineFlags(defaultValues DefaultValues) {
	// Read command line flags
	flag.String(ListenAddr, defaultValues.listenAddr,
		"Hostname or ip address used to listen for incoming connections.")
	flag.Int(ListenPort, defaultValues.listenPort,
		"TCP port used to listen for incoming connections.")
	flag.String(ProxyHost, defaultValues.proxyHost,
		"Hostname or ip address of the target proxy where the queries will be forwarded.")
	flag.Int(ProxyPort, defaultValues.proxyPort,
		"Port number of the target proxy where the queries will be forwarded.")
	flag.String(LOGIN, defaultValues.login,
		"Login to use for proxy auth.")
	flag.String(PASSWORD, defaultValues.password,
		"Login to use for proxy auth.")
	flag.Bool(VerboseLog, false, "Verbose logging. Default to false.")

	// Override config with command line args
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	errBindFlags := viper.BindPFlags(pflag.CommandLine)
	if errBindFlags != nil {
		log.Fatal("Failed to bind command line parameters to configuration", errBindFlags)
	}
}
