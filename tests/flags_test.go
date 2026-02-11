package tests

import "flag"

var (
	bgLoad        bool
	bgCPU         string
	bgMemory      string
	bgRPS         int
	bgPayloadSize int
)

func init() {
	flag.BoolVar(&bgLoad, "bg-load", true, "Enable background load deployment")
	flag.StringVar(&bgCPU, "bg-cpu", "1", "CPU cores for compute stressor")
	flag.StringVar(&bgMemory, "bg-memory", "2Gi", "Memory for compute stressor")
	flag.IntVar(&bgRPS, "bg-rps", 100, "Requests per second for network load")
	flag.IntVar(&bgPayloadSize, "bg-payload-size", 125000, "Payload size in bytes for network load")
}
