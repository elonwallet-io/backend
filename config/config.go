package config

type Config struct {
	MoralisApiKey      string `env:"MORALIS_API_KEY" validate:"required"`
	EnclaveURL         string `env:"ENCLAVE_URL" validate:"required"`
	Host               string `env:"HOST" validate:"required"`
	CorsAllowedUrl     string `env:"CORS_ALLOWED_URL" validate:"required"`
	DBConnectionString string `env:"DB_CONNECTION_STRING" validate:"required"`
	Email              EmailConfig
}

type EmailConfig struct {
	User     string `env:"EMAIL_USER" validate:"required"`
	Password string `env:"EMAIL_PASSWORD" validate:"required"`
	AuthHost string `env:"EMAIL_AUTH_HOST" validate:"required"`
	SmtpHost string `env:"EMAIL_SMTP_HOST" validate:"required"`
}
