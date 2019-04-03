package main

import (
	"os"
	"fmt"
	"flag"
	"os/signal"
	"runtime/trace"
	"runtime/pprof"
	log "github.com/sirupsen/logrus"
)

type Args struct {
	ConfigPath		string
	ShowVersion		bool
	LogLevel		string
	Trace			string
	CPUProfile		string
	MemProfile		string
}

func (a *Args) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(a.LogLevel)
}

func setupTrace(filePath string) func() {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	err = trace.Start(file)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Trace will be written to: %s", filePath)
	return func() {
		trace.Stop()
	}
}

func setupCPUProfile(filePath string) func() {
	cpuprof_file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = pprof.StartCPUProfile(cpuprof_file)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("CPU profile will be written to file: %s", filePath)
	return func() {
		pprof.StopCPUProfile()
		cpuprof_file.Close()
	}
}

func setupMemProfile(filePath string) func() {
	memprof_file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = pprof.WriteHeapProfile(memprof_file)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Mem profile will be written to file: %s", filePath)
	return func() {
		memprof_file.Close()
	}
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{DisableLevelTruncation: true, FullTimestamp: true})
}

func main() {

	args := new(Args)
	flag.StringVar(&args.ConfigPath, "config", "/etc/mbpio.conf", "Path to config file (YAML syntax)")
	flag.BoolVar(&args.ShowVersion, "version", false, "Display current version and exit")
	flag.StringVar(&args.LogLevel, "loglevel", "info", "Set the logging level, choices: trace, debug, info, warn, error")
	flag.StringVar(&args.Trace, "trace", "", "Write execution trace to given file")
	flag.StringVar(&args.CPUProfile, "cpuprofile", "", "Write cpu profile to given file")
	flag.StringVar(&args.MemProfile, "memprofile", "", "Write mem profile to given file")
	flag.Parse()

	if args.ShowVersion {
		fmt.Printf("mbpio v%s\n", Version)
		os.Exit(0)
	}

	if args.CPUProfile != "" {
		onExitFunc := setupCPUProfile(args.CPUProfile)
		defer onExitFunc()
	}

	if args.MemProfile != "" {
		onExitFunc := setupMemProfile(args.MemProfile)
		defer onExitFunc()
	}

	if args.Trace != "" {
		onExitFunc := setupTrace(args.Trace)
		defer onExitFunc()
	}

	logLevel, err := args.GetLogLevel()
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	mbpio, err := NewServer(args.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Infof("received %s signal, pending termination...", sig)
			mbpio.Stop()
		}
	}()

	err = mbpio.Start()
	if err != nil {
		log.Fatalf("runtime error: %s", err)
	}
}