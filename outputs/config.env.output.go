package sdk_dro

type (
	Application struct {
		ENV           string
		PORT          string
		INBOUND_SIZE  int
		OUTBOUND_SIZE int
	}

	Redis struct {
		URL              string
		URLS             []string
		HOST             string
		PORT             int
		USER             string
		PASSWORD         string
		DB               string
		CLUSTER          bool
		CLUSTER_NAME     string
		CLUSTER_PASSWORD string
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

	DatabaseConfig struct {
		TIMEOUT       int
		DIAL_TIMEOUT  int
		READ_TIMEOUT  int
		WRITE_TIMEOUT int
		MAX_CONN      int
		MAX_IDLE      int
		CON_MAX       int
		CON_IDLE      int
	}

	Jwt struct {
		SECRET  string
		EXPIRED int
	}

	RabbitMQ struct {
		URL         string
		URLS        []string
		VSN         string
		HOST        string
		PORT        int
		USER        string
		PASSWORD    string
		SECRET      string
		CONCURRENCY int
		QOS         int
		CLUSTER     bool
	}

	Smtp struct {
		HOST     string
		PORT     int
		USERNAME string
		PASSWORD string
	}

	Environment struct {
		APP      Application
		REDIS    Redis
		POSTGRES Postgres
		MYSQL    Mysql
		DBCONFIG DatabaseConfig
		JWT      Jwt
		RABBITMQ RabbitMQ
		SMTP     Smtp
		BIND     any
	}
)
