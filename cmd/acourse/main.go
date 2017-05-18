package main

import (
	"log"
	"net/http"
	"time"

	"github.com/acoshift/acourse/pkg/app"
	"github.com/acoshift/configfile"
	_ "github.com/lib/pq"
)

var config = configfile.NewReader("config")

func main() {
	time.Local = time.UTC

	// redisPool := &redis.Pool{
	// 	IdleTimeout: 30 * time.Minute,
	// 	MaxIdle:     10,
	// 	MaxActive:   100,
	// 	Wait:        true,
	// 	Dial: func() (redis.Conn, error) {
	// 		return redis.Dial("tcp", config.String("redis_addr"),
	// 			redis.DialDatabase(config.Int("redis_db")),
	// 			redis.DialPassword(config.String("redis_pass")),
	// 		)
	// 	},
	// }

	// init email
	err := app.Init(app.Config{
		ServiceAccount: config.Bytes("service_account"),
		BucketName:     config.String("bucket"),
		EmailServer:    config.String("email_server"),
		EmailPort:      config.Int("email_port"),
		EmailUser:      config.String("email_user"),
		EmailPassword:  config.String("email_password"),
		EmailFrom:      config.String("email_from"),
		BaseURL:        config.String("base_url"),
		XSRFSecret:     config.String("xsrf_secret"),
		SQLURL:         config.String("sql_url"),
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	app.Mount(mux)
	h := app.Middleware(mux)

	// lets reverse proxy handle other settings
	srv := &http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	log.Println("Start server at :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
