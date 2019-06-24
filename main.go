package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"

	flag "github.com/ogier/pflag"
	"gopkg.in/gcfg.v1"
)

const (
	ConfFile         = ".cping"
	CloudFlareApiUrl = "https://www.cloudflare.com/api_json.html"
	ICanHazIpUrl     = "https://icanhazip.com"
)

type Conf struct {
	CloudFlare map[string]*ConfSection
}

type ConfSection struct {
	Email string
	Name  string
	Token string
	Zone  string
}

type EditRecordWrapper struct {
	Result string `json:"result"`
}

type LoadRecordResponse struct {
	Records RecordSet `json:"recs"`
}

type LoadRecordWrapper struct {
	Response LoadRecordResponse `json:"response"`
	Result   string             `json:"result"`
}

type Options struct {
	Verbose bool
}

type Record struct {
	Content string `json:"content"`
	Id      string `json:"rec_id"`
	Name    string `json:"name"`
	Ttl     string `json:"ttl"`
	Type    string `json:"type"`
}

type RecordSet struct {
	Records []Record `json:"objs"`
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}

func getDnsRecord(conf *ConfSection) (*Record, error) {
	query := url.Values{
		"a":     {"rec_load_all"},
		"email": {conf.Email},
		"tkn":   {conf.Token},
		"z":     {conf.Zone},
	}
	resp, err := http.Get(CloudFlareApiUrl + "?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LoadRecordWrapper
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("Failed to query API for target name")
	}

	var target *Record
	for _, record := range result.Response.Records.Records {
		if record.Name == conf.Name {
			target = &record
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("Record not found: %s [zone: %s]",
			conf.Name, conf.Zone)
	}

	return target, nil
}

func getIp() (string, error) {
	resp, err := http.Get(ICanHazIpUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func loadConf() (*Conf, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	var conf Conf
	err = gcfg.ReadFileInto(&conf, user.HomeDir+"/"+ConfFile)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func updateDnsRecord(conf *ConfSection, record *Record, ip string) error {
	form := url.Values{
		"a":       {"rec_edit"},
		"content": {ip},
		"email":   {conf.Email},
		"id":      {record.Id},
		"name":    {conf.Name},
		"tkn":     {conf.Token},
		"ttl":     {record.Ttl},
		"type":    {record.Type},
		"z":       {conf.Zone},
	}
	resp, err := http.PostForm(CloudFlareApiUrl, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result EditRecordWrapper
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	if result.Result != "success" {
		return fmt.Errorf("Failed to update API with new content: %s", ip)
	}

	return nil
}

func main() {
	options := Options{}
	flag.BoolVarP(&options.Verbose, "verbose", "v", false, "Verbose mode")
	flag.Parse()

	conf, err := loadConf()
	if err != nil {
		fail(err)
	}

	ip, err := getIp()
	if err != nil {
		fail(err)
	}
	if options.Verbose {
		fmt.Printf("Current IP: %s\n", ip)
	}

	for _, confSection := range conf.CloudFlare {
		record, err := getDnsRecord(confSection)
		if err != nil {
			fail(err)
		}
		if options.Verbose {
			fmt.Printf("%s: record ID [zone: %s]: %s (%s)\n",
				confSection.Name, confSection.Zone,
				record.Id, record.Content)
		}

		if record.Content != ip {
			err = updateDnsRecord(confSection, record, ip)
			if err != nil {
				fail(err)
			}
			if options.Verbose {
				fmt.Printf("%s: updated successfully\n", confSection.Name)
			}
		} else {
			if options.Verbose {
				fmt.Printf("%s: no update required\n", confSection.Name)
			}
		}
	}
}
