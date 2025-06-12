package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Output struct {
		File  *string
		Level string
	}
	Client struct {
		Interval time.Duration
		Timeout  time.Duration
	}
	Targets map[string]string
}

func test_port(addr string, proto string, timeout time.Duration) error {
	conn, err := net.DialTimeout(proto, addr, timeout)
	if err != nil {
		return err
	} else {
		if conn != nil {
			return nil
		} else {
			return fmt.Errorf("connection was not established")
		}
	}
}

func test_all_ports(config Config) {
	wg := sync.WaitGroup{}

	for name, target := range config.Targets {
		wg.Add(1)
		go func() {
			sections := strings.Split(target, "/")
			err := test_port(sections[0], sections[1], config.Client.Timeout)
			if err != nil {
				log.Printf(" [DOWN] : %s (%s) is down! [error: %s]\n", name, target, err.Error())
			} else if config.Output.Level == "info" {
				log.Printf(" [UP] : %s (%s) is up!\n", name, target)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (c *Config) read_config() {

	f, err := os.ReadFile("./monitor.yml")
	if err != nil {
		log.Fatalf("could not read config file: %s\n", err.Error())
	}

	if err = yaml.Unmarshal(f, c); err != nil {
		log.Fatalf("invalid config file: %s\n", err.Error())
	}

	if c.Client.Timeout >= c.Client.Interval {
		log.Fatalf("invalid config: Timeout cannot be longer than Interval")
	}
}

func main() {
	var config Config
	config.read_config()

	if config.Output.File != nil {
		fmt.Printf("logging to file: %s\n", *config.Output.File)
		log_file, err := os.OpenFile(*config.Output.File, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("could not open log file for writing: %s\n", err.Error())
		}
		log.SetOutput(log_file)
		defer log_file.Close()
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("could not get scheduler: %s\n", err.Error())
	}

	j, err := s.NewJob(gocron.DurationJob(config.Client.Interval), gocron.NewTask(test_all_ports, config))
	if err != nil {
		log.Fatalf("failed to allocate a job for testing ports")
	}

	log.Printf("created job (%d) which runs every %v\n", j.ID(), config.Client.Interval)
	for name, target := range config.Targets {
		fmt.Printf("%s - [%s]\n", name, target)
	}
	fmt.Printf("\n")

	s.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
