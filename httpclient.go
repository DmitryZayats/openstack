package openstack

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// var auth_url string = "https://192.168.122.157:5000/v3/auth/tokens"
var auth_url string = fmt.Sprintf("%s/auth/tokens", os.Getenv("OS_AUTH_URL"))

// fmt.Printf("OS_AUTH_URL is %s\n",os.Getenv("OS_AUTH_URL"))

// var neutron_url string = "https://192.168.122.157:9696"
var auth = []byte(fmt.Sprintf(`
{ "auth": {
    "identity": {
      "methods": ["password"],
      "password": {
        "user": {
          "name": "%s",
          "domain": { "id": "%s" },
          "password": "%s"
        }
      }
    },
	"scope": {
		"project": {
		  "id": "%s"
		}
	  }
  }
}`, os.Getenv("OS_USERNAME"), os.Getenv("OS_PROJECT_DOMAIN_ID"), os.Getenv("OS_PASSWORD"), os.Getenv("OS_PROJECT_ID")))

var tr *http.Transport
var client *http.Client
var token string
var endpoints map[string]string
var security_groups map[string][]string

func init() {
	// proxyUrl, _ := url.Parse("http://localhost:8080")
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// Proxy:           http.ProxyURL(proxyUrl),
	}
	client = &http.Client{Transport: tr}
	// token = Get_token()
}

func Get_API_Endpoints(responsebody []byte) map[string]string {
	var dat map[string]interface{}
	endpoints = make(map[string]string)
	if err := json.Unmarshal(responsebody, &dat); err != nil {
		panic(err)
	}
	// fmt.Println(dat)
	for _, item := range dat["token"].(map[string]interface{})["catalog"].([]interface{}) {
		// fmt.Println("##################################")
		// fmt.Println(item.(map[string]interface{})["name"])
		for _, endpoint := range item.(map[string]interface{})["endpoints"].([]interface{}) {
			if endpoint.(map[string]interface{})["interface"] == "public" {
				// fmt.Printf("Found API end point url for %s : %s\n", item.(map[string]interface{})["name"].(string), endpoint.(map[string]interface{})["url"])
				endpoints[item.(map[string]interface{})["name"].(string)] = endpoint.(map[string]interface{})["url"].(string)
				// name := item.(map[string]interface{})["name"].(string)
				// url := endpoint.(map[string]interface{})["url"].(string)
				// endpoints[name] = append(endpoints[name], url)
			}
		}
		// fmt.Println("##################################")
	}
	// fmt.Println(endpoints)
	return endpoints
}

func Get_token() string {
	request, error := http.NewRequest("POST", auth_url, bytes.NewBuffer(auth))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if error != nil {
		fmt.Println(error)
	}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	// fmt.Println("response Status:", response.Status)
	// fmt.Println("response Headers:", response.Header)
	// body, _ := ioutil.ReadAll(response.Body)
	// fmt.Println("response Body:", string(body))
	responseBody, _ := io.ReadAll(response.Body)
	Get_API_Endpoints(responseBody)
	// fmt.Println(endpoints)
	// fmt.Printf("Format of responseBody is %T\n", responseBody)
	// fmt.Printf("responseBody is %s\n", responseBody)
	token = response.Header["X-Subject-Token"][0]
	return response.Header["X-Subject-Token"][0]
}

func Delete_security_rule(rule_id string) int {
	request, error := http.NewRequest("DELETE", fmt.Sprintf("%s/v2.0/security-group-rules/%s", endpoints["neutron"], rule_id), nil)
	request.Header.Set("X-Auth-Token", token)
	if error != nil {
		fmt.Println(error)
	}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	// fmt.Printf("Rule:%s Status:%d", rule_id, response.StatusCode)
	return response.StatusCode
}

func List_security_rule_ids(group_name string) []string {
	security_group_ids := List_security_groups()
	var rule_ids []string
	group_id := security_group_ids[group_name][0]
	request, error := http.NewRequest("GET", fmt.Sprintf("%s/v2.0/security-group-rules?security_group_id=%s", endpoints["neutron"], group_id), nil)
	request.Header.Set("X-Auth-Token", token)
	if error != nil {
		fmt.Println(error)
	}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	for _, item := range dat["security_group_rules"].([]interface{}) {
		if item.(map[string]interface{})["direction"].(string) == "ingress" {
			rule_ids = append(rule_ids, item.(map[string]interface{})["id"].(string))
		}
	}
	return rule_ids
}

func List_security_groups() map[string][]string {
	security_groups = make(map[string][]string)
	// request, error := http.NewRequest("GET", neutron_url+"/v2.0/security-groups", nil)
	request, error := http.NewRequest("GET", fmt.Sprintf("%s/v2.0/security-groups", endpoints["neutron"]), nil)
	request.Header.Set("X-Auth-Token", token)
	if error != nil {
		fmt.Println(error)
	}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	// fmt.Println("List security groups: response Status:", response.Status)
	// fmt.Println("List security groups: response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	// fmt.Printf("Type of response.Body is %T\n", response.Body)
	// fmt.Println("List security groups: response Body:", string(body))
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	// fmt.Println(dat)
	// return response.Header["X-Subject-Token"][0]defer response.Body.Close()
	for _, sec_group_item := range dat["security_groups"].([]interface{}) {
		// fmt.Printf("Security group name : %s ID : %s\n", sec_group_item.(map[string]interface{})["name"],
		// 	sec_group_item.(map[string]interface{})["id"])
		sec_group_name := sec_group_item.(map[string]interface{})["name"].(string)
		sec_group_id := sec_group_item.(map[string]interface{})["id"].(string)
		security_groups[sec_group_name] = append(security_groups[sec_group_name], sec_group_id)
		// fmt.Printf("Length of security group list is %d\n", len(security_groups[sec_group_name]))
	}
	return security_groups
	// dups := 0
	// for group_name, group_ids := range security_groups {
	// 	if len(group_ids) > 1 {
	// 		fmt.Printf("Security group with name %s is found %d times\n", group_name, len(group_ids))
	// 		dups++
	// 	}
	// }
	// if dups != 0 {
	// 	fmt.Printf("Fix security groups with duplicate names to avoid ambiquity, exiting...")
	// 	os.Exit(1)
	// }
	// fmt.Println("###############################")
	// fmt.Println(security_groups)
}

func Check_security_group_duplicates(security_group_name string, security_groups map[string][]string) int {
	return len(security_groups[security_group_name])
}

func Create_security_group(groupname string, description string) {
	if Check_security_group_duplicates(groupname, List_security_groups()) == 1 {
		fmt.Printf("Security group %s already exists, not creating it\n", groupname)
		return
	}
	payload := []byte(fmt.Sprintf(`
	{
		"security_group": {
			"name": "%s",
			"description": "%s"
		}
	}`, groupname, description))
	request, error := http.NewRequest("POST", fmt.Sprintf("%s/v2.0/security-groups", endpoints["neutron"]), bytes.NewBuffer(payload))
	request.Header.Set("X-Auth-Token", token)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if error != nil {
		fmt.Println(error)
	}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	fmt.Printf("Security group %s create status is %d\n", groupname, response.StatusCode)
	// text, _ := ioutil.ReadAll(response.Body)
	// texts := string(text)
	// fmt.Printf("Response : %s\n", texts)
}

func Get_security_group_id(group_name string) string {
	return security_groups[group_name][0]
}

func Add_remote_group_security_rule(group_name string, remote_group_name string, ipversion string, protocol string) {
	// Flush security_groups map
	for key := range security_groups {
		delete(security_groups, key)
	}
	// List_security_groups()
	if Check_security_group_duplicates(group_name, List_security_groups()) > 1 {
		fmt.Printf("Security group with the name %s is found more than once, exiting to aboid ambiquity\n", group_name)
		os.Exit(1)
	}
	payload := []byte(fmt.Sprintf(`
	{
		"security_group_rule": {
		  "protocol": "%s",
		  "remote_group_id": "%s",
		  "ethertype": "%s",
		  "direction": "ingress",
		  "security_group_id": "%s"
		}
	  }
	`, protocol, security_groups[remote_group_name][0], ipversion, security_groups[group_name][0]))
	request, error := http.NewRequest("POST", fmt.Sprintf("%s/v2.0/security-group-rules", endpoints["neutron"]), bytes.NewBuffer(payload))
	request.Header.Set("X-Auth-Token", token)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if error != nil {
		fmt.Println(error)
	}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	fmt.Printf("Security group rule src:%s dst:%s ipversion:%s protocol:%s create status is %d\n", group_name, remote_group_name, ipversion, protocol, response.StatusCode)
	// text, _ := ioutil.ReadAll(response.Body)
	// texts := string(text)
	// fmt.Printf("Response : %s\n", texts)
}

func Create_security_rule(group_name string, protocol string, remote_ip string, ethertype string, port_range_min string, port_range_max string) (string, string) {
	// Flush security_groups map
	for key := range security_groups {
		delete(security_groups, key)
	}
	List_security_groups()
	if Check_security_group_duplicates(group_name, List_security_groups()) > 1 {
		fmt.Printf("Security group %s found more than once\nExiting to avoid ambiquity...\n", group_name)
		os.Exit(1)
	}
	payload_s := `
	{
		"security_group_rule": { `
	if protocol == "" {
		payload_s += `
		"protocol": null,`
	} else {
		payload_s += fmt.Sprintf(`
		"protocol": "%s",`, protocol)
	}
	if remote_ip != "" {
		payload_s += fmt.Sprintf(`
		  "remote_ip_prefix": "%s",`, remote_ip)
	}
	if port_range_min != "" {
		payload_s += fmt.Sprintf(`
		"port_range_min": %s,`, port_range_min)
	}
	if port_range_max != "" {
		payload_s += fmt.Sprintf(`
		"port_range_max": %s,`, port_range_max)
	}
	payload_s += fmt.Sprintf(`
		  "ethertype": "%s",
		  "direction": "ingress",
		  "security_group_id": "%s"`, ethertype, security_groups[group_name][0])
	payload_s += `
	}
    }
	`

	payload := []byte(payload_s)

	request, error := http.NewRequest("POST", fmt.Sprintf("%s/v2.0/security-group-rules", endpoints["neutron"]), bytes.NewBuffer(payload))
	request.Header.Set("X-Auth-Token", token)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if error != nil {
		fmt.Println(error)
	}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	// fmt.Printf("Security group rule src:%s ipversion:%s protocol:%s create status is %d\n", group_name, ethertype, protocol, response.StatusCode)
	new_security_rule_id := ""
	result := ""
	if response.StatusCode == 409 {
		// fmt.Printf("%s\n", string(body))
		// fmt.Println(strings.Split(strings.Split(string(body), ".")[1], " ")[4])
		new_security_rule_id = strings.Split(strings.Split(string(body), ".")[1], " ")[4]
		result = "exists"
	} else if response.StatusCode == 201 {
		// fmt.Printf("%s\n", string(body))
		var dat map[string]interface{}
		if err := json.Unmarshal(body, &dat); err != nil {
			panic(err)
		}
		new_security_rule_id = dat["security_group_rule"].(map[string]interface{})["id"].(string)
		result = "created"
	}
	fmt.Printf("%s;%5s;%20s;%5s;%5s:%5s;%7s;%40s;%d;\n", group_name, protocol, remote_ip, ethertype, port_range_min, port_range_max, result, new_security_rule_id, response.StatusCode)
	return result, new_security_rule_id
}
