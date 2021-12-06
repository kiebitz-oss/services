module github.com/kiebitz-oss/services

go 1.16

require (
	github.com/bsm/redislock v0.7.1
	github.com/go-redis/redis/v8 v8.11.4
	github.com/kiprotect/go-helpers v0.0.0-20211206172950-5ae593696a59
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
	gopkg.in/yaml.v2 v2.4.0
)

// replace github.com/kiprotect/go-helpers => ../../../../../geordi/kiprotect/go-helpers
