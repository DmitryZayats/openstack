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
)

// var auth_url string = "https://192.168.122.157:5000/v3/auth/tokens"
var auth_url string = fmt.Sprintf("%sauth/tokens", os.Getenv("OS_AUTH_URL"))

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
    }
  }
}`, os.Getenv("OS_USERNAME"), os.Getenv("OS_PROJECT_DOMAIN_ID"), os.Getenv("OS_PASSWORD")))

var tr *http.Transport
var client *http.Client
var token string
var endpoints map[string]string
var security_groups map[string][]string

// func main() {
// 	token = Get_token()
// 	fmt.Printf("Received token : %s\n", token)
// 	List_security_groups()
// 	Create_security_group("Internal_Cluster", "Group that allows communication within cluster")
// 	List_security_groups()
// 	Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv4", "tcp")
// 	Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv4", "udp")
// 	Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv4", "icmp")
// 	Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv6", "tcp")
// 	Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv6", "udp")
// 	Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv6", "icmp")
// }

func init() {
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
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
	fmt.Println(endpoints)
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
	return response.Header["X-Subject-Token"][0]
}

func List_security_groups() {
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
	fmt.Println("List security groups: response Status:", response.Status)
	fmt.Println("List security groups: response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	// fmt.Printf("Type of response.Body is %T\n", response.Body)
	// fmt.Println("List security groups: response Body:", string(body))
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	// fmt.Println(dat)
	// return response.Header["X-Subject-Token"][0]
	for _, sec_group_item := range dat["security_groups"].([]interface{}) {
		fmt.Printf("Security group name : %s ID : %s\n", sec_group_item.(map[string]interface{})["name"],
			sec_group_item.(map[string]interface{})["id"])
		sec_group_name := sec_group_item.(map[string]interface{})["name"].(string)
		sec_group_id := sec_group_item.(map[string]interface{})["id"].(string)
		security_groups[sec_group_name] = append(security_groups[sec_group_name], sec_group_id)
		// fmt.Printf("Length of security group list is %d\n", len(security_groups[sec_group_name]))
	}
	dups := 0
	for group_name, group_ids := range security_groups {
		if len(group_ids) > 1 {
			fmt.Printf("Security group with name %s is found %d times\n", group_name, len(group_ids))
			dups++
		}
	}
	if dups != 0 {
		fmt.Printf("Fix security groups with duplicate names to avoid ambiquity, exiting...")
		os.Exit(1)
	}
	fmt.Println("###############################")
	fmt.Println(security_groups)
}

func Create_security_group(groupname string, description string) {
	payload := []byte(fmt.Sprintf(`
	{
		"security_group": {
			"name": "%s",
			"description": "%s",
			"stateful": true
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
	fmt.Printf("Security group create status is %d\n", response.StatusCode)
	text, _ := ioutil.ReadAll(response.Body)
	texts := string(text)
	fmt.Printf("Response : %s\n", texts)
}

func Get_security_group_id(group_name string) string {
	return security_groups[group_name][0]
}

func Add_remote_group_security_rule(group_name string, remote_group_name string, ipversion string, protocol string) {
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
	fmt.Printf("Security group create status is %d\n", response.StatusCode)
	text, _ := ioutil.ReadAll(response.Body)
	texts := string(text)
	fmt.Printf("Response : %s\n", texts)
}
