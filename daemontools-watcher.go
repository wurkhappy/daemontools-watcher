package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	to      = flag.String("to", "", "destination Internet mail address")
	from    = flag.String("from", "", "source Internet mail address")
	address = flag.String("address", "localhost:25", "smpt server address:port")
)

func main() {
	flag.Parse()
	if *to == "" || *from == "" {
		log.Fatal("to and from flags are required")
	}
	c, err := smtp.Dial(*address)
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(4 * time.Second)
		servicesOut, err := exec.Command("ls", "/service/").Output()
		if err != nil {
			fmt.Println("error occured")
			fmt.Printf("%s", err)
		}
		var servicesDown []string
		for _, line := range strings.Split(string(servicesOut), "\n") {
			serviceStatus, err := exec.Command("svstat", "/service/"+line).Output()
			if err != nil {
				fmt.Println("error occured")
				fmt.Printf("%s", err)
			}
			serviceData := strings.Split(string(serviceStatus), " ")
			if serviceData[1] == "up" {
				seconds, _ := strconv.Atoi(serviceData[4])
				if seconds < 5 {
					servicesDown = append(servicesDown, fmt.Sprintf("%s has restarted", serviceData[0]))
				}
			}
		}
		if len(servicesDown) == 0 {
			continue
		}

		c.Mail(*from)
		c.Rcpt(*to)
		wc, err := c.Data()
		if err != nil {
			log.Fatal(err)
		}
		buf := bytes.NewBufferString("To: " + *to + "\r\nSubject: Daemontools Update\r\n\r\n" + strings.Join(servicesDown, "\n"))
		if _, err = buf.WriteTo(wc); err != nil {
			log.Fatal(err)
		}
		wc.Close()
	}
}
