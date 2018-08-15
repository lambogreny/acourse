package main

import (
	"context"
	"database/sql"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/profiler"
	"cloud.google.com/go/storage"
	"github.com/acoshift/configfile"
	"github.com/acoshift/go-firebase-admin"
	"github.com/acoshift/hime"
	"github.com/acoshift/middleware"
	"github.com/acoshift/probehandler"
	"github.com/acoshift/session"
	redisstore "github.com/acoshift/session/store/goredis"
	"github.com/acoshift/webstatic"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"google.golang.org/api/option"

	"github.com/acoshift/acourse/app"
	"github.com/acoshift/acourse/context/redisctx"
	"github.com/acoshift/acourse/context/sqlctx"
	"github.com/acoshift/acourse/email"
	"github.com/acoshift/acourse/image"
	"github.com/acoshift/acourse/internal"
	"github.com/acoshift/acourse/notify"
)

func main() {
	time.Local = time.UTC

	config := configfile.NewReader("config")

	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// init profiler
	profiler.Start(profiler.Config{Service: "acourse"})

	firApp, err := firebase.InitializeApp(ctx, firebase.AppOptions{
		ProjectID: config.String("project_id"),
	}, option.WithCredentialsFile("config/service_account"))
	if err != nil {
		log.Fatal(err)
	}
	firAuth := firApp.Auth()

	// init google storage
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile("config/service_account"))
	if err != nil {
		log.Fatal(err)
	}
	bucketHandle := storageClient.Bucket(config.String("bucket"))

	// init email sender
	emailSender := email.NewSMTPSender(email.SMTPConfig{
		Server:   config.String("email_server"),
		Port:     config.Int("email_port"),
		User:     config.String("email_user"),
		Password: config.String("email_password"),
		From:     config.String("email_from"),
	})

	adminNotifier := notify.NewOutgoingWebhookAdminNotifier(config.String("slack_url"))

	// init redis pool
	redisClient := redis.NewClient(&redis.Options{
		MaxRetries:  3,
		PoolSize:    5,
		IdleTimeout: 60 * time.Minute,
		Addr:        config.String("redis_addr"),
		Password:    config.String("redis_pass"),
	})

	// init databases
	db, err := sql.Open("postgres", config.String("sql_url"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(4)

	himeApp := hime.New()
	himeApp.ParseConfigFile("settings/server.yaml")
	himeApp.ParseConfigFile("settings/routes.yaml")

	static := configfile.NewYAMLReader("static.yaml")
	himeApp.TemplateFunc("static", func(s string) string {
		return "/-/" + static.StringDefault(s, s)
	})

	baseURL := config.String("base_url")

	himeApp.Template().
		Funcs(app.TemplateFunc()).
		ParseConfigFile("settings/template.yaml")

	mux := http.NewServeMux()

	// health check
	probe := probehandler.New()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ready") == "1" {
			// readiness probe
			probe.ServeHTTP(w, r)
			return
		}

		// liveness probe
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/-/", http.StripPrefix("/-", webstatic.New(webstatic.Config{
		Dir:          "assets",
		CacheControl: "public, max-age=31536000",
	})))

	mux.Handle("/favicon.ico", internal.FileHandler("assets/favicon.ico"))

	mux.Handle("/", middleware.Chain(
		sqlctx.Middleware(db),
		redisctx.Middleware(redisClient, config.String("redis_prefix")),
		session.Middleware(session.Config{
			Secret:   config.Bytes("session_secret"),
			Path:     "/",
			MaxAge:   7 * 24 * time.Hour,
			HTTPOnly: true,
			Secure:   session.PreferSecure,
			SameSite: session.SameSiteLax,
			Rolling:  true,
			Proxy:    true,
			Store: redisstore.New(redisstore.Config{
				Prefix: config.String("redis_prefix"),
				Client: redisClient,
			}),
		}),
	)(app.New(app.Config{
		BaseURL:            baseURL,
		Auth:               firAuth,
		Location:           loc,
		AdminNotifier:      adminNotifier,
		EmailSender:        emailSender,
		BucketHandle:       bucketHandle,
		BucketName:         config.String("bucket"),
		ImageResizeEncoder: image.NewJPEGResizeEncoder(),
	})))

	h := middleware.Chain(
		internal.ErrorRecovery,
		internal.SetHeaders,
		middleware.CSRF(middleware.CSRFConfig{
			Origins:     []string{baseURL},
			IgnoreProto: true,
			ForbiddenHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Cross-site origin detected!", http.StatusForbidden)
			}),
		}),
	)(mux)

	err = himeApp.
		Handler(h).
		GracefulShutdown().
		Notify(probe.Fail).
		ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
