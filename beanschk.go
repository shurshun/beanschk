package main

import (
	"flag"
	"fmt"
	"github.com/kr/beanstalk"
	"os"
	"strconv"
	"strings"
)

var (
	LOCAL_SRV  = flag.String("h", "127.0.0.1:11300", "ip:port of the local beanstalk server")
	CRIT_LIMIT = flag.Int("c", 1000, "Critical limit")
	WARN_LIMIT = flag.Int("w", 500, "Warning limit")

	errors    []string
	exit_code int = 0
)

func init() {
	flag.Parse()
}

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

func main() {
	c, err := beanstalk.Dial("tcp", *LOCAL_SRV)

	check(err)

	defer c.Close()

	localTubes, err := c.ListTubes()

	check(err)

	for _, t := range localTubes {
		tube := &beanstalk.Tube{c, t}

		stat, err := tube.Stats()

		if err != nil {
			continue
		}

		current_jobs, _ := strconv.Atoi(stat["current-jobs-ready"])

		switch {
		case current_jobs >= *WARN_LIMIT && current_jobs < *CRIT_LIMIT:
			errors = append(errors, fmt.Sprintf("%s/%d", t, current_jobs))
			setExitCode(1)
		case current_jobs >= *CRIT_LIMIT:
			errors = append(errors, fmt.Sprintf("%s/%d", t, current_jobs))
			setExitCode(2)
		}
	}

	if len(errors) > 0 {
		exit(exit_code, "tube/jobs: "+strings.Join(errors, ", "))
	}

	exit(exit_code, "ok")

}
