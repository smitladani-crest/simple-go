package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var ipAddress string
var hostname string

const indexPage = `
<h1 style="color: %s">Hello Learner</h1>
<h2>Your IP - %s</h2>
<p><b>Request Header</b> - %s</p>
<h3>Served from %s</h3>
<h4>Hostname: %s</h4>
`

func indexPageHandler(response http.ResponseWriter, request *http.Request) {
	log.Printf("%s %s %s\n", request.RemoteAddr, request.Method, request.RequestURI)

	color := mux.Vars(request)["color"]
	if color == "" {
		// if no color is provided in the URL then black is used
		color = "black"
	}

	var headerString string

	for name, value := range request.Header {
		headerString += name + ": " + strings.Join(value, ", ") + "<br/>"
	}

	fmt.Fprintf(response, indexPage, color, request.RemoteAddr, headerString, ipAddress, hostname)
}

var router = mux.NewRouter()

func getIPAddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func main() {
	var err error

	hostname, _ = os.Hostname()

	port := os.Getenv("GoPort")
	if port == "" {
		port = "8080"
	}

	ipAddress, err = getIPAddress()
	if err != nil {
		log.Println(err)
	}

	router.HandleFunc("/", indexPageHandler)

	router.HandleFunc("/health", func(response http.ResponseWriter, request *http.Request) {
		log.Printf("%s %s %s\n", request.RemoteAddr, request.Method, request.RequestURI)

		fmt.Fprintf(response, "healthy")
	})

	router.HandleFunc("/{color}", indexPageHandler)

	http.Handle("/", router)

	log.Println("Server is listening at 0.0.0.0:" + port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
