module github.com/kiebitz-oss/services

go 1.16

require (
	github.com/bsm/redislock v0.7.1
	github.com/go-redis/redis/v8 v8.11.4
	github.com/kiprotect/go-helpers v0.0.0-20211210144244-79ce90e73e79
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	gopkg.in/yaml.v2 v2.4.0
)

// replace github.com/kiprotect/go-helpers => ../../../../../geordi/kiprotect/go-helpers
