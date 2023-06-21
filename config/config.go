package config

type Config struct {
	MoralisApiKey      string `env:"MORALIS_API_KEY" validate:"required"`
	DBConnectionString string `env:"DB_CONNECTION_STRING" validate:"required"`
	BackendHost        string `env:"BACKEND_HOST" validate:"required_if=UseInsecureHTTP false"`
	FrontendURL        string `env:"FRONTEND_URL" validate:"required"`
	DeployerURL        string `env:"DEPLOYER_URL" validate:"required"`
	UseInsecureHTTP    bool   `env:"USE_INSECURE_HTTP"`
	Environment        string `env:"ENVIRONMENT"`
	Email              EmailConfig
	Wallet             WalletConfig
}

type EmailConfig struct {
	User     string `env:"EMAIL_USER" validate:"required"`
	Password string `env:"EMAIL_PASSWORD" validate:"required"`
	AuthHost string `env:"EMAIL_AUTH_HOST" validate:"required"`
	SmtpHost string `env:"EMAIL_SMTP_HOST" validate:"required"`
}

type WalletConfig struct {
	PrivateKeyHex string `env:"WALLET_PRIVATE_KEY_HEX" validate:"required"`
	Address       string `env:"WALLET_ADDRESS" validate:"required"`
}
