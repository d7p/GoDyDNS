package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const defaultOptionsFile = "options.json"

type DNSPutRequet struct {
	RecordType string `json:"type"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
	Proxied    bool   `json:"proxied"`
}

type Options struct {
	BaseURL,
	APIKey,
	ZoneID,
	RecordID,
	AuthEmail string
}

func main() {
	ip := getIP()

	if !areIPsEqual(ip) {
		setURL(ip, loadOptions())
		ioutil.WriteFile("oldip", []byte(ip), os.FileMode(int(0664)))
		os.Exit(0)
	}

	log.Print("IP is the same")
}

func areIPsEqual(ip string) bool {
	oldIP, err := ioutil.ReadFile("oldip")
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Fatal(err)
		}
	}

	return ip == string(oldIP)
}

func getIP() string {
	res, err := http.Get("http://bot.whatismyipaddress.com")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	ip := string(content)
	return ip
}

func setURL(ip string, opt *Options) {
	body := &DNSPutRequet{
		RecordType: "A",
		Name:       "dydns.codefission.co.uk",
		TTL:        120,
		Proxied:    false,
	}
	body.Content = ip

	url := fmt.Sprintf("%szones/%s/dns_records/%s", opt.BaseURL, opt.ZoneID, opt.RecordID)
	bodybytes, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodybytes))

	req.Header.Set("X-Auth-Email", opt.AuthEmail)
	req.Header.Set("X-Auth-Key", opt.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == 400 || res.StatusCode == 401 || res.StatusCode == 500 {
		log.Print(string(contents))
		log.Fatalf("CloudFlare returned: %s", res.Status)
	}
}

func loadOptions() *Options {
	optionsbyte, err := ioutil.ReadFile(defaultOptionsFile)
	if err != nil {
		log.Fatal(err)
	}
	var options Options

	err2 := json.Unmarshal(optionsbyte, &options)
	if err2 != nil {
		log.Fatal(err2)
	}
	return &options
}
