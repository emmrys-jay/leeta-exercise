package config

type DatabaseConfiguration struct {
	Protocol string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ServerConfiguration struct {
	HttpUrl            string
	HttpPort           string
	HttpAllowedOrigins string
}

type AppConfiguration struct {
	Name string
	Env  string
}

type Configuration struct {
	App      AppConfiguration
	Server   ServerConfiguration
	Database DatabaseConfiguration
}
