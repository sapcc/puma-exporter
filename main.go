package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli"
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
	pumaBacklog = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "puma_backlog",
		Help: "Number of established but unaccepted connections in the backlog",
	}, []string{"index"})
	pumaRunning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "puma_running",
		Help: "Number of running worker threads",
	}, []string{"index"})
	pumaPoolCapacity = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "puma_pool_capacity",
		Help: "Number of allocatable worker threads",
	}, []string{"index"})
	pumaMaxThreads = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "puma_max_threads",
		Help: "Maximum number of worker threads",
	}, []string{"index"})
	pumaRequestsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "puma_requests_count",
		Help: "Number of processed requests",
	}, []string{"index"})
	pumaWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_workers",
		Help: "Number of configured workers",
	})
	pumaBootedWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_booted_workers",
		Help: "Number of booted workers",
	})
	pumaOldWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_old_workers",
		Help: "Number of old workers",
	})
	//We use our own registry instead of the default to avoid the standard metrics
	registry = prometheus.NewRegistry()
)

type pumaStats struct {
	StartedAt     time.Time `json:"started_at"`
	Workers       int       `json:"workers"`
	Phase         int       `json:"phase"`
	BootedWorkers int       `json:"booted_workers"`
	OldWorkers    int       `json:"old_workers"`
	WorkerStatus  []struct {
		StartedAt   time.Time `json:"started_at"`
		Pid         int       `json:"pid"`
		Index       int       `json:"index"`
		Phase       int       `json:"phase"`
		Booted      bool      `json:"booted"`
		LastCheckin time.Time `json:"last_checkin"`
		LastStatus  struct {
			Backlog       int `json:"backlog"`
			Running       int `json:"running"`
			PoolCapacity  int `json:"pool_capacity"`
			MaxThreads    int `json:"max_threads"`
			RequestsCount int `json:"requests_count"`
		} `json:"last_status"`
	} `json:"worker_status"`
}

func init() {
	// Metrics have to be registered to be exposed:
	// with workers labels
	registry.MustRegister(pumaBacklog)
	registry.MustRegister(pumaRunning)
	registry.MustRegister(pumaRequestsCount)
	registry.MustRegister(pumaPoolCapacity)
	registry.MustRegister(pumaMaxThreads)
	// without labels
	registry.MustRegister(pumaWorkers)
	registry.MustRegister(pumaBootedWorkers)
	registry.MustRegister(pumaOldWorkers)

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
			Value:  "0.0.0.0:9882",
			EnvVar: "BIND_ADDRESS",
		},
		cli.StringFlag{
			Name:   "control-url,u",
			Usage:  "url for the puma control socket",
			EnvVar: "CONTROL_URL",
			Value:  "http://127.0.0.1:9292",
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
	// without labels
	pumaWorkers.Set(float64(stats.Workers))
	pumaBootedWorkers.Set(float64(stats.BootedWorkers))
	pumaOldWorkers.Set(float64(stats.OldWorkers))
	for i, status := range stats.WorkerStatus {
		index := fmt.Sprintf("%d", i)
		pumaBacklog.With(prometheus.Labels{"index": index}).Set(float64(status.LastStatus.Backlog))
		pumaRunning.With(prometheus.Labels{"index": index}).Set(float64(status.LastStatus.Running))
		pumaRequestsCount.With(prometheus.Labels{"index": index}).Set(float64(status.LastStatus.RequestsCount))
		pumaPoolCapacity.With(prometheus.Labels{"index": index}).Set(float64(status.LastStatus.PoolCapacity))
		pumaMaxThreads.With(prometheus.Labels{"index": index}).Set(float64(status.LastStatus.MaxThreads))
	}
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
