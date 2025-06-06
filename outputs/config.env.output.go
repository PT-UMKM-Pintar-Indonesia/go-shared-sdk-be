package sdk_opt

type (
	Application struct {
		ENV          string
		PORT         string
		INBOUND_SIZE int
	}

	Redis struct {
		URL string
	}

	Postgres struct {
		URL string
	}

	Jwt struct {
		SECRET  string
		EXPIRED int
	}

	RabbitMQ struct {
		URL    string
		VSN    string
		SECRET string
	}

	Smtp struct {
		HOST     string
		PORT     int
		USERNAME string
		PASSWORD string
	}

	Environtment[T any] struct {
		APP      Application
		REDIS    Redis
		POSTGRES Postgres
		JWT      Jwt
		RABBITMQ RabbitMQ
		SMTP     Smtp
		BIND     T
	}
)
