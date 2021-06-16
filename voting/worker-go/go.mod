module atix.de/voting/worker-go

go 1.16

require (
	github.com/go-redis/redis/v8 v8.10.0
	github.com/jackc/pgx/v4 v4.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.7.1
	github.com/valyala/fasthttp v1.26.0
	golang.org/x/sys v0.0.0-20210608053332-aa57babbf139 // indirect
)

replace (
	atix.de/voting/worker-go/postgres => ./postgres
	atix.de/voting/worker-go/redis => ./redis
)
