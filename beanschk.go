package main

import (
	"fmt"
	"github.com/kr/beanstalk"
	"github.com/urfave/cli"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	errors    []string
	exit_code int = 0
)

func check(e error) {
	if e != nil {
		exit(2, e.Error())
	}
}

func exit(code int, msg string) {
	fmt.Println(" " + msg)
	os.Exit(code)
}

func setExitCode(code int) {
	if code > exit_code {
		exit_code = code
	}
}

func checkTubes(c *cli.Context) {
	beansClient, err := beanstalk.Dial("tcp", net.JoinHostPort(c.String("bean-addr"), c.String("bean-port")))

	check(err)

	defer beansClient.Close()

	localTubes, err := beansClient.ListTubes()

	check(err)

	for _, t := range localTubes {
		tube := &beanstalk.Tube{beansClient, t}

		stat, err := tube.Stats()

		check(err)

		jobsCount, _ := strconv.Atoi(stat["current-jobs-ready"])

		switch {
		case jobsCount >= c.Int("warn-limit") && jobsCount < c.Int("crit-limit"):
			errors = append(errors, fmt.Sprintf("%s/%d", t, jobsCount))
			setExitCode(1)
		case jobsCount >= c.Int("crit-limit"):
			errors = append(errors, fmt.Sprintf("%s/%d", t, jobsCount))
			setExitCode(2)
		}
	}

	if len(errors) > 0 {
		exit(exit_code, "tube/jobs: "+strings.Join(errors, ", "))
	}

	exit(exit_code, "ok")
}

func main() {
	app := cli.NewApp()

	app.Name = "beanschk"
	app.Usage = "Beanstalkd tubes checker"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "bean-addr, a",
			Value:  "localhost",
			Usage:  "Beanstalk server addr",
			EnvVar: "BEAN_ADDR",
		},
		cli.StringFlag{
			Name:   "bean-port, p",
			Value:  "11300",
			Usage:  "Beanstalk server port",
			EnvVar: "BEAN_PORT",
		},
		cli.IntFlag{
			Name:   "warn-limit, w",
			Value:  500,
			Usage:  "Warning limit",
			EnvVar: "WARN_LIMIT",
		},
		cli.IntFlag{
			Name:   "crit-limit, c",
			Value:  1000,
			Usage:  "Critical limit",
			EnvVar: "CRIT_LIMIT",
		},
	}

	app.Action = func(c *cli.Context) {
		checkTubes(c)
	}

	app.Run(os.Args)
}
