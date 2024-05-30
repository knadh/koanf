package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/hashicorp/nomad/api"
	"github.com/knadh/koanf/providers/nomad/v2"
	"github.com/knadh/koanf/v2"
)

// Example and test
// Requirements:
//
// - Docker installed
// - Docker daemon running
// - Nomad installed
// - Nomad running
// (sudo nomad agent -dev)
// - Nomad echo project running
// (nomad run httpecho.hcl)
// - curl
func main() {
	fmt.Printf("Checking...\n")
	var k = koanf.New(".")
	var k2 = koanf.New(".")
	var k3 = koanf.New(".")

	// allocs
	out, err := exec.Command("curl", "http://localhost:4646/v1/allocations").Output()
	if err != nil {
		log.Fatalf("error retreiving allocs from curl: %v\n", err)
	}

	var aListStub = []api.AllocationListStub{}
	if err = json.Unmarshal(out, &aListStub); err != nil {
		log.Fatalf("error unmarshalling alloc list: %s\n", err)
	}

	nAllocs := len(aListStub)
	allocsHTTPData := make([]api.Allocation, nAllocs)

	for i := 0; i < nAllocs; i++ {
		cmdPath := fmt.Sprintf("http://localhost:4646/v1/allocation/%s", aListStub[i].ID)
		out, err := exec.Command("curl", cmdPath).Output()
		if err != nil {
			log.Fatalf("error retrieving alloc with ID = %s: %s\n", aListStub[i].ID, err)
		}

		json.Unmarshal(out, &allocsHTTPData[i])
	}

	nmdAllocs, err := nomad.Provider(nil, "allocs")
	if err != nil {
		log.Fatalf("error creating provider: %v\n", err)
	}

	if err = k.Load(nmdAllocs, nil); err != nil {
		log.Fatalf("error loading allocs: %v\n", err)
	}

	// allocs: checking
	for i1 := 0; i1 < nAllocs; i1++ {
		ID := allocsHTTPData[i1].ID

		var key string
		var flagFound bool = false

		// finding alloc with this ID
		for i2 := 0; i2 < nAllocs; i2++ {
			key = "echo" + strconv.Itoa(i2)
			if k.String(key+".ID") == ID {
				flagFound = true
				break
			}
		}

		if !flagFound {
			log.Fatalf("Alloc ID not found: test failed\n")
		}

		// comparing all data
		if k.String(key+".Namespace") != allocsHTTPData[i1].Namespace {
			fmt.Printf("Alloc Namespace: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".EvalID") != allocsHTTPData[i1].EvalID {
			fmt.Printf("Alloc EvalID: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".Name") != allocsHTTPData[i1].Name {
			fmt.Printf("Alloc Name: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".NodeID") != allocsHTTPData[i1].NodeID {
			fmt.Printf("Alloc NodeID: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".NodeName") != allocsHTTPData[i1].NodeName {
			fmt.Printf("Alloc NodeName: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".JobID") != allocsHTTPData[i1].JobID {
			fmt.Printf("Alloc JobID: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".TaskGroup") != allocsHTTPData[i1].TaskGroup {
			fmt.Printf("Alloc TaskGroup: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".DesiredStatus") != allocsHTTPData[i1].DesiredStatus {
			fmt.Printf("Alloc DesiredStatus: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".ClientStatus") != allocsHTTPData[i1].ClientStatus {
			fmt.Printf("Alloc ClientStatus: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".DeploymentID") != allocsHTTPData[i1].DeploymentID {
			fmt.Printf("Alloc DeploymentID: test failed\n")
			os.Exit(1)
		}

		if k.String(key+".DeploymentID") == allocsHTTPData[i1].ID {
			fmt.Printf("Alloc DeploymentID = ID: test failed\n")
			os.Exit(1)
		}

		// network comparison
		for k1 := 0; k1 < len(allocsHTTPData[i1].Resources.Networks); k1++ {
			ipSearch := allocsHTTPData[i1].Resources.Networks[k1].IP

			// ipSearch in koanf
			// koanf key
			var keyNetwork string
			for k2 := 0; k2 < len(allocsHTTPData[i1].Resources.Networks); k2++ {
				keyNetwork = key + ".Network" + strconv.Itoa(k1) + ".IP"
				if ipSearch == k.String(keyNetwork) {
					break
				}
			}

			// port search
			for k2 := 0; k2 < len(allocsHTTPData[i1].Resources.Networks[k1].DynamicPorts); k2++ {
				flagPort := false
				for k3 := 0; k3 <= k2; k3++ {
					keyPort := key + ".Network" + strconv.Itoa(k1) + ".Port" + strconv.Itoa(k3)
					if k.Int(keyPort) == allocsHTTPData[i1].Resources.Networks[k1].DynamicPorts[k2].Value {
						flagPort = true
						break
					}
				}

				if flagPort {
					flagPort = false
				} else {
					fmt.Printf(
						"Alloc ports: test failed, port %d isn't found\n",
						allocsHTTPData[i1].Resources.Networks[k1].DynamicPorts[k2].Value)
					os.Exit(1)
				}
			}
		}
	}

	fmt.Printf("Allocations: test passed\n")

	// raft
	out, err = exec.Command("curl", "http://localhost:4646/v1/operator/raft/configuration").Output()
	if err != nil {
		log.Fatalf("error retreiving raft configuration from curl: %v\n", err)
	}

	raftCfg := api.RaftConfiguration{}
	if err = json.Unmarshal(out, &raftCfg); err != nil {
		log.Fatalf("error unmarshalling raft configuration: %s\n", err)
	}

	nmdRaft, err := nomad.Provider(nil, "raft")
	if err != nil {
		log.Fatalf("error creating provider: %v\n", err)
	}

	if err = k2.Load(nmdRaft, nil); err != nil {
		log.Fatalf("error loading allocs: %v\n", err)
	}

	nServers := len(raftCfg.Servers)

	// raft: checking
	for i1 := 0; i1 < nServers; i1++ {
		ID := raftCfg.Servers[i1].ID

		var key string
		var flagFound bool = false

		// finding alloc with this ID
		for i2 := 0; i2 < nServers; i2++ {
			key = "server" + strconv.Itoa(i2)
			if k2.String(key+".ID") == ID {
				flagFound = true
				break
			}
		}

		if !flagFound {
			log.Fatalf("Raft server ID not found: test failed")
		}

		if k2.String(key+".Node") != raftCfg.Servers[i1].Node {
			log.Fatalf("Raft server node: test failed")
		}

		if k2.String(key+".Address") != raftCfg.Servers[i1].Address {
			log.Fatalf("Raft server address: test failed")
		}

		if k2.Bool(key+".Leader") != raftCfg.Servers[i1].Leader {
			log.Fatalf("Raft server leader: test failed")
		}

		if k2.Bool(key+".Voter") != raftCfg.Servers[i1].Voter {
			log.Fatalf("Raft server voter: test failed")
		}

		if k2.String(key+".RaftProtocol") != raftCfg.Servers[i1].RaftProtocol {
			log.Fatalf("Raft server protocol: test failed")
		}
	}

	fmt.Printf("Raft configuration: test passed\n")

	// vars
	_, err = exec.Command(
		"curl",
		"--header",
		"Content-Type: application/json",
		"--request",
		"POST",
		"--data",
		`{"Path": "databases/sql", "Items": {"mysql": "75cp21", "postgresql": "52pg24"}}`,
		"http://localhost:4646/v1/var/databases/sql",
	).Output()
	if err != nil {
		log.Fatalf("Couldn't create variables: %v\n", err)
	}

	nmdVars, err := nomad.Provider(nil, "vars")
	if err != nil {
		log.Fatalf("error creating provider: %v\n", err)
	}

	if err = k3.Load(nmdVars, nil); err != nil {
		log.Fatalf("error loading vars: %v\n", err)
	}

	// vars: checking
	if k3.String("databases/sql.mysql") != "75cp21" {
		log.Fatalf("Vars: test failed\n")
	}

	if k3.String("databases/sql.postgresql") != "52pg24" {
		log.Fatalf("Vars: test failed\n")
	}

	fmt.Printf("Vars: test passed\n")

	fmt.Printf("ALL TESTS PASSED\n")
}
