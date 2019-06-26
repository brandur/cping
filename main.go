package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/cloudflare/cloudflare-go"
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

type Options struct {
	Verbose bool
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
	os.Exit(1)
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

func updateRecord(options *Options, ip string, confSection *ConfSection) error {
	api, err := cloudflare.New(
		confSection.Token,
		confSection.Email,
	)
	if err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(confSection.Zone)
	if err != nil {
		return err
	}

	filter := cloudflare.DNSRecord{Name: confSection.Name, Type: "A"}
	records, err := api.DNSRecords(zoneID, filter)
	if err != nil {
		return err
	}

	if len(records) < 1 {
		return fmt.Errorf("Record not found: %s [zone: %s]",
			confSection.Name, confSection.Zone)
	}

	if len(records) > 1 {
		return fmt.Errorf("Too many records found: %s [zone: %s]",
			confSection.Name, confSection.Zone)
	}

	record := records[0]

	if options.Verbose {
		fmt.Printf("%s: record ID [zone: %s]: %s (%s)\n",
			confSection.Name, confSection.Zone,
			record.ID, record.Content)
	}

	if record.Content == ip {
		if options.Verbose {
			fmt.Printf("%s: no update required\n", confSection.Name)
		}

		return nil
	}

	err = api.UpdateDNSRecord(zoneID, record.ID, cloudflare.DNSRecord{
		Content: ip,
		Name:    record.Name,
		Type:    record.Type,
	})
	if err != nil {
		fail(err)
	}

	if options.Verbose {
		fmt.Printf("%s: updated successfully\n", confSection.Name)
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
		err := updateRecord(&options, ip, confSection)
		if err != nil {
			fail(err)
		}
	}
}
