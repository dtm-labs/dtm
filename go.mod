module github.com/dtm-labs/dtm

go 1.15

require (
	github.com/dtm-labs/dtmdriver v0.0.1
	github.com/dtm-labs/dtmdriver-gozero v0.0.1
	github.com/dtm-labs/dtmdriver-polaris v0.0.2
	github.com/dtm-labs/dtmdriver-protocol1 v0.0.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/lib/pq v1.10.3
	github.com/lithammer/shortuuid v2.0.3+incompatible
	github.com/lithammer/shortuuid/v3 v3.0.7
	github.com/onsi/gomega v1.16.0
	github.com/prometheus/client_golang v1.11.0
	github.com/stretchr/testify v1.7.0
	go.etcd.io/bbolt v1.3.6
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/automaxprocs v1.4.1-0.20210525221652-0180b04c18a7
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa // indirect
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.0.3
	gorm.io/driver/postgres v1.2.1
	gorm.io/gorm v1.22.2
// gotest.tools v2.2.0+incompatible
)

replace google.golang.org/grpc => google.golang.org/grpc v1.38.0
