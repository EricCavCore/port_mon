package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Output struct {
		File       *string
		Web        *string
		MaxHistory int `yaml:"max_history"`
	}
	Client struct {
		Interval time.Duration
		Timeout  time.Duration
	}
	Targets map[string]string
}

type Target struct {
	Name      string
	Address   string
	Protocol  string
	Alive     bool
	EventID   uint64
	LastUp    *time.Time
	LastDown  *time.Time
	TimesDown []time.Time
}

var (
	NextEventID atomic.Uint64
)

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

func test_all_ports(config Config, targets []Target) {
	wg := sync.WaitGroup{}

	for i, t := range targets {
		wg.Add(1)
		go func() {
			err := test_port(t.Address, t.Protocol, config.Client.Timeout)
			if err != nil && t.Alive {
				targets[i].Alive = false
				if len(targets[i].TimesDown) < config.Output.MaxHistory {
					targets[i].TimesDown = append(targets[i].TimesDown, time.Now())
				}
				targets[i].EventID = NextEventID.Add(1)
				if targets[i].LastDown == nil {
					targets[i].LastDown = new(time.Time)
				}
				*targets[i].LastDown = time.Now()
				log.Printf(" [DOWN] (eventID: %d) : %s (%s) is down! [error: %s]\n", targets[i].EventID, t.Name, t.Address, err.Error())
			} else if err == nil {
				if targets[i].LastUp == nil {
					targets[i].LastUp = new(time.Time)
				}
				*targets[i].LastUp = time.Now()
				if !targets[i].Alive && targets[i].LastUp != nil {
					targets[i].Alive = true
					log.Printf(" [UP] (eventID: %d) : %s (%s) is up! Down for %v, been down %d times\n", t.EventID, t.Name, t.Address, (time.Since(*targets[i].LastUp)), len(t.TimesDown))
				}
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

func (c *Config) make_targets() []Target {
	var targets []Target
	for name, addr_proto := range c.Targets {
		sections := strings.Split(addr_proto, "/")

		targets = append(targets, Target{
			Name:      name,
			Address:   sections[0],
			Protocol:  sections[1],
			Alive:     false,
			LastUp:    nil,
			TimesDown: []time.Time{},
		})
	}

	return targets
}

func main() {
	NextEventID.Store(0)

	var config Config
	config.read_config()
	targets := config.make_targets()

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

	j, err := s.NewJob(gocron.DurationJob(config.Client.Interval), gocron.NewTask(test_all_ports, config, targets))
	if err != nil {
		log.Fatalf("failed to allocate a job for testing ports")
	}

	log.Printf("created job (%d) which runs every %v\n", j.ID(), config.Client.Interval)
	for name, target := range config.Targets {
		fmt.Printf("%s - [%s]\n", name, target)
	}
	fmt.Printf("\n")

	s.Start()

	go func() {
		http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
			t_data, err := json.MarshalIndent(targets, "", "    ")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to marshal data"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(t_data)
		})

		http.Handle("/", http.FileServer(http.Dir("www")))

		err := http.ListenAndServe(*config.Output.Web, nil)
		if err != nil {
			log.Fatalf("http server crashed: %s\n", err.Error())
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
