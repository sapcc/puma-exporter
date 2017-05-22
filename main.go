package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	appName = "Puma exporter"
)

var (
	landingPage = []byte(fmt.Sprintf(`<html>
<head><title>Puma exporter</title></head>
<body>
<h1>Puma exporter</h1>
<p>%s</p>
<p><a href='/metrics'>Metrics</a></p>
</body>
</html>
`, versionString()))
	requestBacklog = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_request_backlog",
		Help: "Number of requests waiting to be processed by a thread",
	})
	runningThreads = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_thread_count",
		Help: "Number of threads currently running",
	})

	//We use our own registry instead of the default to avoid the standard metrics
	registry = prometheus.NewRegistry()
)

type pumaStats struct {
	RunningThreads int `json:"running"`
	RequestBacklog int `json:"backlog"`
}

func init() {
	// Metrics have to be registered to be exposed:
	registry.MustRegister(requestBacklog)
	registry.MustRegister(runningThreads)
}

func main() {
	app := cli.NewApp()

	app.Name = appName
	app.Version = versionString()
	app.Authors = []cli.Author{
		{
			Name:  "Fabian Ruff",
			Email: "fabian.ruff@sap.com",
		},
	}
	app.Usage = "Prometheus exporter for the puma webserver"
	app.Action = runServer
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "bind-address,b",
			Usage:  "Listen address for metrics HTTP endpoint",
			Value:  "0.0.0.0:9235",
			EnvVar: "BIND_ADDRESS",
		},
		cli.StringFlag{
			Name:   "control-url,u",
			Usage:  "url for the puma control socket",
			EnvVar: "CONTROL_URL",
			Value:  "http://127.0.0.1:7353",
		},
		cli.StringFlag{
			Name:   "auth-token,a",
			Usage:  "authentication token for the control server",
			EnvVar: "AUTH_TOKEN",
		},
	}

	app.Run(os.Args)
}

func runServer(c *cli.Context) {

	go func() {
		for {
			updateMetrics(fmt.Sprintf("%s/stats?token=%s", c.GlobalString("control-url"), c.GlobalString("auth-token")))
			time.Sleep(5 * time.Second)
		}
	}()

	// init the router and server
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", serveVersion)
	log.Printf("Listening on %s...", c.GlobalString("bind-address"))
	err := http.ListenAndServe(c.GlobalString("bind-address"), nil)
	fatalfOnError(err, "Failed to bind on %s: ", c.GlobalString("bind-address"))
}

func updateMetrics(url string) {

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch metrics: %s", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("Got error %s from control url", resp.Status)
		return
	}

	var stats pumaStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Printf("Error decoding response from control server: %s", err)
	}
	requestBacklog.Set(float64(stats.RequestBacklog))
	runningThreads.Set(float64(stats.RunningThreads))
}

func serveVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(landingPage)
}

func fatalfOnError(err error, msg string, args ...interface{}) {
	if err != nil {
		log.Fatalf(msg, args...)
	}
}
