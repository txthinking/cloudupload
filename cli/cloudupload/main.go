package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/mux"
	"github.com/pilu/xrequestid"
	"github.com/rs/cors"
	"github.com/txthinking/cloudupload"
	"github.com/unrolled/secure"
	"golang.org/x/crypto/acme/autocert"
)

var debug bool
var debugListen string

func main() {
	app := cli.NewApp()
	app.Name = "Cloud Upload"
	app.Version = "20200411"
	app.Usage = "Upload files to multiple cloud storage in parallel"
	app.Authors = []*cli.Author{
		{
			Name:  "Cloud",
			Email: "cloud@txthinking.com",
		},
	}
	app.Copyright = "https://github.com/txthinking/cloudupload"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			Usage:       "Enable debug, more logs",
			Destination: &debug,
		},
		&cli.StringFlag{
			Name:        "debugListen",
			Aliases:     []string{"l"},
			Usage:       "Listen address for debug",
			Value:       "127.0.0.1:6060",
			Destination: &debugListen,
		},
		&cli.StringFlag{
			Name:  "listen",
			Usage: "Listen address",
		},
		&cli.StringFlag{
			Name:  "domain",
			Usage: "If domain is specified, 80 and 443 ports will be used. Listen address is no longer needed",
		},
		&cli.Int64Flag{
			Name:  "maxBodySize",
			Usage: "Max size of http body, M",
		},
		&cli.Int64Flag{
			Name:  "timeout",
			Usage: "Read timeout, write timeout x2, idle timeout x20, s",
		},
		&cli.StringSliceFlag{
			Name:  "origin",
			Usage: "Allow origins for CORS, can repeat more times. like https://google.com",
		},
		&cli.BoolFlag{
			Name:  "enableLocal",
			Usage: "Enable local store",
		},
		&cli.StringFlag{
			Name:  "localStorage",
			Value: "",
			Usage: "Local directory path",
		},
		&cli.BoolFlag{
			Name:  "enableGoogle",
			Usage: "Enable google store",
		},
		&cli.StringFlag{
			Name:  "googleServiceAccountFile",
			Value: "",
			Usage: "Google service account file",
		},
		&cli.StringFlag{
			Name:  "googleBucket",
			Value: "",
			Usage: "Google bucket name",
		},
		&cli.BoolFlag{
			Name:  "enableAliyun",
			Usage: "Enable aliyun OSS",
		},
		&cli.StringFlag{
			Name:  "aliyunAccessKeyID",
			Value: "",
			Usage: "Aliyun access key id",
		},
		&cli.StringFlag{
			Name:  "aliyunAccessKeySecret",
			Value: "",
			Usage: "Aliyun access key secret",
		},
		&cli.StringFlag{
			Name:  "aliyunEndpoint",
			Value: "",
			Usage: "Aliyun endpoint, like: https://oss-cn-shanghai.aliyuncs.com",
		},
		&cli.StringFlag{
			Name:  "aliyunBucket",
			Value: "",
			Usage: "Aliyun bucket name",
		},
		&cli.BoolFlag{
			Name:  "enableTencent",
			Usage: "Enable Tencent",
		},
		&cli.StringFlag{
			Name:  "tencentSecretId",
			Value: "",
			Usage: "Tencent secret id",
		},
		&cli.StringFlag{
			Name:  "tencentSecretKey",
			Value: "",
			Usage: "Tencent secret key",
		},
		&cli.StringFlag{
			Name:  "tencentHost",
			Value: "",
			Usage: "domain",
		},
	}
	app.Action = func(c *cli.Context) error {
		ss := make([]cloudupload.Storer, 0, 1)
		if c.Bool("enableLocal") {
			local := &cloudupload.Local{
				StoragePath: c.String("localStorage"),
			}
			ss = append(ss, local)
		}
		if c.Bool("enableGoogle") {
			google := &cloudupload.Google{
				ServiceAccountFile: c.String("googleServiceAccountFile"),
				Bucket:             c.String("googleBucket"),
			}
			ss = append(ss, google)
		}
		if c.Bool("enableAliyun") {
			aliyun := &cloudupload.Aliyun{
				AccessKeyID:     c.String("aliyunAccessKeyID"),
				AccessKeySecret: c.String("aliyunAccessKeySecret"),
				Endpoint:        c.String("aliyunEndpoint"),
				Bucket:          c.String("aliyunBucket"),
			}
			ss = append(ss, aliyun)
		}
		if c.Bool("enableTencent") {
			tencent := &cloudupload.Tencent{
				SecretId:  c.String("tencentSecretId"),
				SecretKey: c.String("tencentSecretKey"),
				Host:      c.String("tencentHost"),
			}
			ss = append(ss, tencent)
		}
		if debug {
			go func() {
				log.Println(http.ListenAndServe(debugListen, nil))
			}()
		}
		return runHTTPServer(c.String("listen"), c.String("domain"), c.StringSlice("origin"), ss, c.Int64("maxBodySize")*1024*1024, c.Int64("timeout"))
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runHTTPServer(address, domain string, origins []string, stores []cloudupload.Storer, maxBodySize, timeout int64) error {
	up := &cloudupload.Upload{Stores: stores}
	r := mux.NewRouter()
	r.Methods("POST").Path("/").Handler(up)
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	if debug {
		n.Use(negroni.NewLogger())
	}
	if domain != "" {
		n.Use(negroni.HandlerFunc(secure.New(secure.Options{
			AllowedHosts:            []string{domain},
			SSLRedirect:             false,
			SSLHost:                 domain,
			SSLProxyHeaders:         map[string]string{"X-Forwarded-Proto": "https"},
			STSSeconds:              315360000,
			STSIncludeSubdomains:    true,
			STSPreload:              true,
			FrameDeny:               true,
			CustomFrameOptionsValue: "SAMEORIGIN",
			ContentTypeNosniff:      true,
			BrowserXssFilter:        true,
			ContentSecurityPolicy:   "default-src 'self'",
		}).HandlerFuncWithNext))
	}
	if len(origins) != 0 {
		n.Use(cors.New(cors.Options{
			AllowedOrigins:     origins,
			AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:     []string{"Accept", "Content-Type", "X-File-Name", "X-Request-With"},
			MaxAge:             60 * 60 * 24 * 30,
			ExposedHeaders:     []string{"Content-Type", "X-Request-Id"},
			AllowCredentials:   true,
			OptionsPassthrough: true,
		}))
	}

	n.UseFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		w.Header().Set("Server", "github.com/txthinking/cloudupload")
		next(w, r)
	})

	n.UseFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		next(w, r)
	})

	n.Use(xrequestid.New(16))

	for _, store := range stores {
		if local, ok := store.(*cloudupload.Local); ok {
			n.Use(negroni.NewStatic(http.Dir(local.StoragePath)))
			break
		}
	}

	lmt := tollbooth.NewLimiter(30, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	if domain == "" {
		lmt.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"})
	} else {
		lmt.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})
	}
	n.UseFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if r.Method == "POST" {
			httpError := tollbooth.LimitByRequest(lmt, w, r)
			if httpError != nil {
				w.Header().Add("Content-Type", lmt.GetMessageContentType())
				w.WriteHeader(httpError.StatusCode)
				w.Write([]byte(httpError.Message))
				return
			}
		}
		next(w, r)
	})

	n.UseHandler(r)

	if domain == "" {
		s := &http.Server{
			Addr:           address,
			ReadTimeout:    time.Duration(timeout) * time.Second,
			WriteTimeout:   time.Duration(timeout) * 2 * time.Second,
			IdleTimeout:    time.Duration(timeout) * 20 * time.Second,
			MaxHeaderBytes: 1 << 20,
			Handler:        n,
		}
		go func() {
			log.Println(s.ListenAndServe())
		}()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		return s.Shutdown(context.Background())
	}
	m := autocert.Manager{
		Cache:      autocert.DirCache(".letsencrypt"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Email:      "cloud@txthinking.com",
	}
	go http.ListenAndServe(":80", m.HTTPHandler(nil))
	ss := &http.Server{
		Addr:           ":443",
		ReadTimeout:    time.Duration(timeout) * time.Second,
		WriteTimeout:   time.Duration(timeout) * 2 * time.Second,
		IdleTimeout:    time.Duration(timeout) * 20 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        n,
		TLSConfig:      &tls.Config{GetCertificate: m.GetCertificate},
	}
	go func() {
		log.Println(ss.ListenAndServeTLS("", ""))
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	return ss.Shutdown(context.Background())
}
