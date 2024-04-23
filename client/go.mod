module client-demo

go 1.19

// replace github.com/sigmavirus24/circuitry => ./lib/circuitry/

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/bsm/redislock v0.9.4
	github.com/go-resty/resty/v2 v2.12.0
	github.com/redis/go-redis/v9 v9.5.1
	github.com/sigmavirus24/circuitry v0.1.1
	github.com/sirupsen/logrus v1.9.3
	github.com/sony/gobreaker v0.5.0
	golang.org/x/net v0.22.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/smartystreets/goconvey v1.8.1 // indirect
	golang.org/x/sys v0.18.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
