package config

var GlobalConfig AppConfig

type AppConfig struct {
	CapitalizeName bool
	Language       string
}

func LoadConfig() {
	// Load your configurations into GlobalConfig
	// GlobalConfig.CapitalizeName = strings.ToUpper(os.Getenv("GOCLIAPP_CAPITALIZE_NAME")) == "TRUE"
	// GlobalConfig.Language = os.Getenv("GOCLIAPP_LANGUAGE")
	// if GlobalConfig.Language == "" {
	// 	GlobalConfig.Language = "English" // Default language
	// }
}
