package config

type Config struct {
	Repository RepositoryConfig
	Server     ServerConfig
	Api        ApiConfig
}

type ApiConfig struct {
	MoralisApiKey string `env:"MORALIS_API_KEY" validate:"required"`
	EnclaveURL    string `env:"ENCLAVE_URL" validate:"required"`
}

type ServerConfig struct {
	CorsAllowedUrl string `env:"CORS_ALLOWED_URL" validate:"required"`
}

type RepositoryConfig struct {
	ConnectionString string `env:"DATABASE_CONNECTION_STRING" validate:"required"`
}
