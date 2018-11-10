package config

// ConfigData contains app configuration data
type ConfigData struct {
	Database ConfigDatabase
	Redis    ConfigRedis
}

// ConfigDatabase contains database configuration data
type ConfigDatabase struct {
	User    string
	Name    string
	Pass    string
	Address string
	SSL     string
}

// ConfigRedis contains redis configuration data
type ConfigRedis struct {
	Address  string
	Password string
	DataBase int
}

func GetDefaultConfig() ConfigData {
	return ConfigData{
		Database: ConfigDatabase{
			User:    "gochanuser",
			Pass:    "gochanpass",
			Name:    "gochandb",
			SSL:     "disable",
			Address: "localhost",
		},
		Redis: ConfigRedis{
			Address:  "localhost:6379",
			Password: "",
			DataBase: 0,
		},
	}
}
