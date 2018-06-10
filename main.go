package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/cli"
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

	count = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_count",
		Help: "Number of all (minor+major) GC runs",
	})
	minorGcCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_minor_gc_count",
		Help: "Number of minor GC runs",
	})
	majorGcCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_major_gc_count",
		Help: "Number of major GC runs",
	})

	heapAllocatedPages = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_allocated_pages",
		Help: "puma_gc_heap_allocated_pages",
	})
	heapSortedLength = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_sorted_length",
		Help: "puma_gc_heap_sorted_length",
	})
	heapAllocatablePages = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_allocatable_pages",
		Help: "puma_gc_heap_allocatable_pages",
	})
	heapAvailableSlots = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_available_slots",
		Help: "puma_gc_heap_available_slots",
	})
	heapLiveSlots = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_live_slots",
		Help: "puma_gc_heap_live_slots",
	})
	heapFreeSlots = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_free_slots",
		Help: "puma_gc_heap_free_slots",
	})
	heapFinalSlots = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_final_slots",
		Help: "puma_gc_heap_final_slots",
	})
	heapMarkedSlots = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_marked_slots",
		Help: "puma_gc_heap_marked_slots",
	})
	heapSweptSlots = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_swept_slots",
		Help: "puma_gc_heap_swept_slots",
	})
	heapEdenPages = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_eden_pages",
		Help: "puma_gc_heap_eden_pages",
	})
	heapTombPages = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_heap_tomb_pages",
		Help: "puma_gc_heap_tomb_pages",
	})
	totalAllocatedPages = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_total_allocated_pages",
		Help: "puma_gc_total_allocated_pages",
	})
	totalFreedPages = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_total_freed_pages",
		Help: "puma_gc_total_freed_pages",
	})
	totalAllocatedObjects = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_total_allocated_objects",
		Help: "puma_gc_total_allocated_objects",
	})
	totalFreedObjects = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_total_freed_objects",
		Help: "puma_gc_total_freed_objects",
	})
	mallocIncreaseBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_malloc_increase_bytes",
		Help: "puma_gc_malloc_increase_bytes",
	})
	mallocIncreaseBytesLimit = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_malloc_increase_bytes_limit",
		Help: "puma_gc_malloc_increase_bytes_limit",
	})
	rememberedWbUnprotectedObjects = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_remembered_wb_unprotected_objects",
		Help: "puma_gc_remembered_wb_unprotected_objects",
	})
	rememberedWbUnprotectedObjectsLimit = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_remembered_wb_unprotected_objects_limit",
		Help: "puma_gc_remembered_wb_unprotected_objects_limit",
	})
	oldObjects = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_old_objects",
		Help: "puma_gc_old_objects",
	})
	oldObjectsLimit = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_old_objects_limit",
		Help: "puma_gc_old_objects_limit",
	})
	oldmallocIncreaseBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_oldmalloc_increase_bytes",
		Help: "puma_gc_oldmalloc_increase_bytes",
	})
	oldmallocIncreaseBytesLimit = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "puma_gc_oldmalloc_increase_bytes_limit",
		Help: "puma_gc_oldmalloc_increase_bytes_limit",
	})

	//We use our own registry instead of the default to avoid the standard metrics
	registry = prometheus.NewRegistry()
)

type pumaStats struct {
	RunningThreads int `json:"running"`
	RequestBacklog int `json:"backlog"`
}

type pumaGcStats struct {
	Count                               int `json:"count"`
	MinorGcCount                        int `json:"minor_gc_count"`
	MajorGcCount                        int `json:"major_gc_count"`
	HeapAllocatedPage                   int `json:"heap_allocated_pages"`
	HeapSortedLength                    int `json:"heap_sorted_length"`
	HeapAllocatablePage                 int `json:"heap_allocatable_pages"`
	HeapAvailableSlot                   int `json:"heap_available_slots"`
	HeapLiveSlot                        int `json:"heap_live_slots"`
	HeapFreeSlot                        int `json:"heap_free_slots"`
	HeapFinalSlot                       int `json:"heap_final_slots"`
	HeapMarkedSlot                      int `json:"heap_marked_slots"`
	HeapSweptSlot                       int `json:"heap_swept_slots"`
	HeapEdenPage                        int `json:"heap_eden_pages"`
	HeapTombPage                        int `json:"heap_tomb_pages"`
	TotalAllocatedPage                  int `json:"total_allocated_pages"`
	TotalFreedPage                      int `json:"total_freed_pages"`
	TotalAllocatedObject                int `json:"total_allocated_objects"`
	TotalFreedObject                    int `json:"total_freed_objects"`
	MallocIncreaseByte                  int `json:"malloc_increase_bytes"`
	MallocIncreaseBytesLimit            int `json:"malloc_increase_bytes_limit"`
	RememberedWbUnprotectedObject       int `json:"remembered_wb_unprotected_objects"`
	RememberedWbUnprotectedObjectsLimit int `json:"remembered_wb_unprotected_objects_limit"`
	OldObject                           int `json:"old_objects"`
	OldObjectsLimit                     int `json:"old_objects_limit"`
	OldmallocIncreaseByte               int `json:"oldmalloc_increase_bytes"`
	OldmallocIncreaseBytesLimit         int `json:"oldmalloc_increase_bytes_limit"`
}

func init() {
	// Metrics have to be registered to be exposed:
	registry.MustRegister(requestBacklog)
	registry.MustRegister(runningThreads)
	registry.MustRegister(count)
	registry.MustRegister(minorGcCount)
	registry.MustRegister(majorGcCount)
	registry.MustRegister(heapAllocatedPages)
	registry.MustRegister(heapSortedLength)
	registry.MustRegister(heapAllocatablePages)
	registry.MustRegister(heapAvailableSlots)
	registry.MustRegister(heapLiveSlots)
	registry.MustRegister(heapFreeSlots)
	registry.MustRegister(heapFinalSlots)
	registry.MustRegister(heapMarkedSlots)
	registry.MustRegister(heapSweptSlots)
	registry.MustRegister(heapEdenPages)
	registry.MustRegister(heapTombPages)
	registry.MustRegister(totalAllocatedPages)
	registry.MustRegister(totalFreedPages)
	registry.MustRegister(totalAllocatedObjects)
	registry.MustRegister(totalFreedObjects)
	registry.MustRegister(mallocIncreaseBytes)
	registry.MustRegister(mallocIncreaseBytesLimit)
	registry.MustRegister(rememberedWbUnprotectedObjects)
	registry.MustRegister(rememberedWbUnprotectedObjectsLimit)
	registry.MustRegister(oldObjects)
	registry.MustRegister(oldObjectsLimit)
	registry.MustRegister(oldmallocIncreaseBytes)
	registry.MustRegister(oldmallocIncreaseBytesLimit)
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
			updateGcMetrics(fmt.Sprintf("%s/gc-stats?token=%s", c.GlobalString("control-url"), c.GlobalString("auth-token")))
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

func updateGcMetrics(url string) {
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

	var gcstats pumaGcStats
	if err := json.NewDecoder(resp.Body).Decode(&gcstats); err != nil {
		log.Printf("Error decoding response from control server: %s", err)
	}

	count.Set(float64(gcstats.Count))
	minorGcCount.Set(float64(gcstats.MinorGcCount))
	majorGcCount.Set(float64(gcstats.MajorGcCount))
	heapAllocatedPages.Set(float64(gcstats.HeapAllocatedPage))
	heapSortedLength.Set(float64(gcstats.HeapSortedLength))
	heapAllocatablePages.Set(float64(gcstats.HeapAllocatablePage))
	heapAvailableSlots.Set(float64(gcstats.HeapAvailableSlot))
	heapLiveSlots.Set(float64(gcstats.HeapLiveSlot))
	heapFreeSlots.Set(float64(gcstats.HeapFreeSlot))
	heapFinalSlots.Set(float64(gcstats.HeapFinalSlot))
	heapMarkedSlots.Set(float64(gcstats.HeapMarkedSlot))
	heapSweptSlots.Set(float64(gcstats.HeapSweptSlot))
	heapEdenPages.Set(float64(gcstats.HeapEdenPage))
	heapTombPages.Set(float64(gcstats.HeapTombPage))
	totalAllocatedPages.Set(float64(gcstats.TotalAllocatedPage))
	totalFreedPages.Set(float64(gcstats.TotalFreedPage))
	totalAllocatedObjects.Set(float64(gcstats.TotalAllocatedObject))
	totalFreedObjects.Set(float64(gcstats.TotalFreedObject))
	mallocIncreaseBytes.Set(float64(gcstats.MallocIncreaseByte))
	mallocIncreaseBytesLimit.Set(float64(gcstats.MallocIncreaseBytesLimit))
	rememberedWbUnprotectedObjects.Set(float64(gcstats.RememberedWbUnprotectedObject))
	rememberedWbUnprotectedObjectsLimit.Set(float64(gcstats.RememberedWbUnprotectedObjectsLimit))
	oldObjects.Set(float64(gcstats.OldObject))
	oldObjectsLimit.Set(float64(gcstats.OldObjectsLimit))
	oldmallocIncreaseBytes.Set(float64(gcstats.OldmallocIncreaseByte))
	oldmallocIncreaseBytesLimit.Set(float64(gcstats.OldmallocIncreaseBytesLimit))
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
