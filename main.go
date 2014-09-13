package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	CloudFlareApiUrl = "https://www.cloudflare.com/api_json.html"
	ICanHazIpUrl     = "https://icanhazip.com"

	Email = ""
	Name  = ""
	Zone  = ""
	Token = ""
)

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

type Record struct {
	Id   string `json:"rec_id"`
	Name string `json:"name"`
	Ttl  string `json:"ttl"`
	Type string `json:"type"`
}

type RecordSet struct {
	Records []Record `json:"objs"`
}

func getDnsRecord() (*Record, error) {
	query := url.Values{}
	query.Set("a", "rec_load_all")
	query.Set("email", Email)
	query.Set("tkn", Token)
	query.Set("z", Zone)

	resp, err := http.Get(CloudFlareApiUrl + "?" + query.Encode())
	if err != nil {
		return nil, err
	}

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
		if record.Name == Name {
			target = &record
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("Record not found: %s [zone: %s]", Name, Zone)
	}

	return target, nil
}

func getIp() (string, error) {
	resp, err := http.Get(ICanHazIpUrl)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func updateDnsRecord(record *Record, ip string) error {
	form := url.Values{}
	form.Set("a", "rec_edit")
	form.Set("content", ip)
	form.Set("email", Email)
	form.Set("id", record.Id)
	form.Set("name", Name)
	form.Set("tkn", Token)
	form.Set("ttl", record.Ttl)
	form.Set("type", record.Type)
	form.Set("z", Zone)
	resp, err := http.PostForm(CloudFlareApiUrl, form)
	if err != nil {
		return err
	}

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
	record, err := getDnsRecord()
	if err != nil {
		panic(err)
	}

	ip, err := getIp()
	if err != nil {
		panic(err)
	}

	err = updateDnsRecord(record, ip)
	if err != nil {
		panic(err)
	}
}
