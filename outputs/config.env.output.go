package sdk_opt

type (
	Application struct {
		ENV          string
		PORT         string
		INBOUND_SIZE int
	}

	Redis struct {
		URL      string
		HOST     string
		PORT     int
		USER     string
		PASSWORD string
		DB       string
	}

	Postgres struct {
		URL      string
		HOST     string
		PORT     int
		USER     string
		PASSWORD string
		DB       string
	}

	Mysql struct {
		URL      string
		HOST     string
		PORT     int
		USER     string
		PASSWORD string
		DB       string
	}

	Jwt struct {
		SECRET  string
		EXPIRED int
	}

	RabbitMQ struct {
		URL      string
		VSN      string
		HOST     string
		PORT     int
		USER     string
		PASSWORD string
		SECRET   string
	}

	Smtp struct {
		HOST     string
		PORT     int
		USERNAME string
		PASSWORD string
	}

	Environtment struct {
		APP      Application
		REDIS    Redis
		POSTGRES Postgres
		MYSQL    Mysql
		JWT      Jwt
		RABBITMQ RabbitMQ
		SMTP     Smtp
		BIND     any
	}
)
