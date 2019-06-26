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

func main() {
	opts := runOptions{}
	flag.BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose mode")
	flag.Parse()

	conf, err := loadconf()
	if err != nil {
		fail(err)
	}

	ip, err := getIP()
	if err != nil {
		fail(err)
	}
	if opts.Verbose {
		fmt.Printf("Current IP: %s\n", ip)
	}

	for _, confSection := range conf.CloudFlare {
		err := updateRecord(&opts, ip, confSection)
		if err != nil {
			fail(err)
		}
	}
}

//
// Private
//

const (
	confFile     = ".cping"
	iCanHazIPURL = "https://icanhazip.com"
)

type conf struct {
	CloudFlare map[string]*confSection
}

type confSection struct {
	Email string
	Name  string
	Token string
	Zone  string
}

type runOptions struct {
	Verbose bool
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}

func getIP() (string, error) {
	resp, err := http.Get(iCanHazIPURL)
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

func loadconf() (*conf, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	var conf conf
	err = gcfg.ReadFileInto(&conf, user.HomeDir+"/"+confFile)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func updateRecord(opts *runOptions, ip string, confSection *confSection) error {
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

	if opts.Verbose {
		fmt.Printf("%s: record ID [zone: %s]: %s (%s)\n",
			confSection.Name, confSection.Zone,
			record.ID, record.Content)
	}

	if record.Content == ip {
		if opts.Verbose {
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

	if opts.Verbose {
		fmt.Printf("%s: updated successfully\n", confSection.Name)
	}

	return nil
}
