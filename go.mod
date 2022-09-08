module github.com/dtm-labs/dtm

go 1.16

require (
	bou.ke/monkey v1.0.2
	github.com/dtm-labs/dtmdriver v0.0.6
	github.com/dtm-labs/dtmdriver-dapr v0.0.1
	github.com/dtm-labs/dtmdriver-gozero v0.0.7
	github.com/dtm-labs/dtmdriver-kratos v0.0.9
	github.com/dtm-labs/dtmdriver-polaris v0.0.5
	github.com/dtm-labs/dtmdriver-springcloud v1.2.3
	github.com/dtm-labs/logger v0.0.1
	github.com/gin-gonic/gin v1.7.7
	github.com/go-errors/errors v1.4.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/lib/pq v1.10.6
	github.com/lithammer/shortuuid/v3 v3.0.7
	github.com/onsi/gomega v1.18.1
	github.com/prometheus/client_golang v1.12.2
	github.com/stretchr/testify v1.8.0
	github.com/ugorji/go v1.2.7 // indirect
	go.etcd.io/bbolt v1.3.6
	go.mongodb.org/mongo-driver v1.9.1
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.0.3
	gorm.io/driver/postgres v1.2.1
	gorm.io/gorm v1.22.2
// gotest.tools v2.2.0+incompatible
)

// replace github.com/dtm-labs/dtmdriver v0.0.2 => /Users/wangxi/dtm/dtmdriver

// replace github.com/dtm-labs/dtmdriver-http => /Users/wangxi/dtm/dtmdriver-http-nacos
