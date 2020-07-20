package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/mingcheng/pidfile"
	"github.com/BrandYourself/galera-healthcheck/healthcheck"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		healthchecker         *healthcheck.Healthchecker
		serverPort             int
		mysqlUser              string
		mysqlPassword          string
		availableWhenDonor     bool
		availableWhenReadOnly  bool
		pidfilePath            string
	)

	app := &cli.App{
		Name: "galera-healthcheck",
		Usage: "A lightweight web server to report the health of a node in a Galera cluster",
		Authors: []*cli.Author{
			&cli.Author{ Name: "Stefan Schimanski" },
			&cli.Author{ Name: "Brendan Clougherty", Email: "bclougherty@brand-yourself.com" },
		},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name: "port, P",
				Usage: "Specifies the port of the healthcheck server",
				Value: 8080,
				Destination: &serverPort,
			},
			&cli.StringFlag{
				Name: "user, u",
				Usage: "Specifies the MySQL user to connect as",
				EnvVars: []string{"MYSQL_USER"},
				Destination: &mysqlUser,
				Required: true,
			},
			&cli.StringFlag{
				Name: "password, p",
				Usage: "Specifies the MySQL password to connect with",
				EnvVars: []string{"MYSQL_PASSWORD"},
				Destination: &mysqlPassword,
				Required: true,
			},
			&cli.BoolFlag{
				Name: "availWhenDonor, d",
				Usage: "Specifies if the healthcheck allows availability when in donor state",
				Value: true,
				Destination: &availableWhenDonor,
			},
			&cli.BoolFlag{
				Name: "availWhenReadOnly, r",
				Usage: "Specifies if the healthcheck allows availability when in read only mode",
				Value: false,
				Destination: &availableWhenReadOnly,
			},
			&cli.StringFlag{
				Name: "pidfile, i",
				Usage: "Location for the pidfile",
				Value: "/var/run/galera-healthcheck.pid",
				Destination: &pidfilePath,
			},
		},
		Action: func(c *cli.Context) error {
			pid, err := pidfile.New(pidfilePath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error creating pid file: %+v", err), 1)
			}

			// Create a channel to listen for signals
			sigs := make(chan os.Signal, 1)

			// Notify the channel on receiving a SIGINT or SIGTERM
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			// Run a background handler that will remove the pidfile on getting a signal
			go func() {
				<-sigs
				pid.Remove()
				os.Exit(1)
			}()

			db, _ := sql.Open("mysql", fmt.Sprintf("%s:%s@/", mysqlUser, mysqlPassword))
			config := healthcheck.HealthcheckerConfig{
				availableWhenDonor,
				availableWhenReadOnly,
			}

			healthchecker = healthcheck.New(db, config)

			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				result := healthchecker.Check()
				if result.Healthy {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusServiceUnavailable)
				}

				jsonContent, err := json.Marshal(result)
				if err != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					fmt.Fprintf(w, "Error while encoding JSON response: %s", err.Error())
				}

				fmt.Fprintf(w, string(jsonContent))
			})

			http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)

			pid.Remove()

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
