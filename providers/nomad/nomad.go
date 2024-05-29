package nomad

import (
	"errors"
	"strconv"

	"github.com/hashicorp/nomad/api"
)

var errUnprovided = errors.New("Nomad provider does not support this method")

// Nomad implements Nomad provider.
// datatype can be "allocs", "raft" or "vars"
type Nomad struct {
	client *api.Client
	dtype  string
}

// Provider returns an instance of Nomad provider.
func Provider(cfg *api.Config, datatype string) (*Nomad, error) {
	if cfg == nil {
		cfg = api.DefaultConfig()
	}

	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Nomad{client: c, dtype: datatype}, nil
}

// ReadBytes is not supported by the Nomad provider.
func (n *Nomad) ReadBytes() ([]byte, error) {
	return nil, errUnprovided
}

// TODO: read with configuration

func (n *Nomad) Read() (map[string]interface{}, error) {
	var mp map[string]interface{}
	var err error

	switch n.dtype {
	case "allocs":
		mp, err = n.ReadAllocs()
		if err != nil {
			return nil, err
		}
	case "raft":
		mp, err = n.ReadRaft()
		if err != nil {
			return nil, err
		}
	case "vars":
		mp, err = n.ReadVars()
		if err != nil {
			return nil, err
		}
	}

	return mp, nil
}

func (n *Nomad) ReadAllocs() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	mpKeys := make(map[string]int)

	allocs := n.client.Allocations()

	allocsInfo, _, err := allocs.List(nil)
	if err != nil {
		return nil, err
	}

	nAllocs := len(allocsInfo)
	allocsData := make([]*api.Allocation, nAllocs)

	for i := 0; i < nAllocs; i++ {
		allocsData[i], _, err = allocs.Info(allocsInfo[i].ID, nil)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < nAllocs; i++ {
		mpAlloc := make(map[string]interface{})

		mpAlloc["ID"] = allocsData[i].ID
		mpAlloc["Namespace"] = allocsData[i].Namespace
		mpAlloc["EvalID"] = allocsData[i].EvalID
		mpAlloc["Name"] = allocsData[i].Name
		mpAlloc["NodeID"] = allocsData[i].NodeID
		mpAlloc["NodeName"] = allocsData[i].NodeName
		mpAlloc["JobID"] = allocsData[i].JobID
		mpAlloc["TaskGroup"] = allocsData[i].TaskGroup
		mpAlloc["DesiredStatus"] = allocsData[i].DesiredStatus
		mpAlloc["ClientStatus"] = allocsData[i].ClientStatus
		mpAlloc["DeploymentID"] = allocsData[i].DeploymentID

		for n := 0; n < len(allocsData[i].Resources.Networks); n++ {
			mpNetwork := make(map[string]interface{})

			mpNetwork["IP"] = allocsData[i].Resources.Networks[n].IP

			for p := 0; p < len(allocsData[i].Resources.Networks[n].DynamicPorts); p++ {
				keyPort := "Port" + strconv.Itoa(p)
				mpNetwork[keyPort] = allocsData[i].Resources.Networks[n].DynamicPorts[p].Value
			}

			keyNetwork := "Network" + strconv.Itoa(n)
			mpAlloc[keyNetwork] = mpNetwork
		}

		keyAlloc := allocsData[i].TaskGroup
		if _, ok := mpKeys[keyAlloc]; ok {
			mpKeys[keyAlloc]++
		} else {
			mpKeys[keyAlloc] = 0
		}

		keyAlloc += strconv.Itoa(mpKeys[keyAlloc])
		mp[keyAlloc] = mpAlloc
	}

	return mp, nil
}

func (n *Nomad) ReadRaft() (map[string]interface{}, error) {
	mp := make(map[string]interface{})

	operator := n.client.Operator()
	raftCfg, err := operator.RaftGetConfiguration(nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(raftCfg.Servers); i++ {
		mpServer := make(map[string]interface{})

		mpServer["ID"] = raftCfg.Servers[i].ID
		mpServer["Node"] = raftCfg.Servers[i].Node
		mpServer["Address"] = raftCfg.Servers[i].Address
		mpServer["Leader"] = raftCfg.Servers[i].Leader
		mpServer["Voter"] = raftCfg.Servers[i].Voter
		mpServer["RaftProtocol"] = raftCfg.Servers[i].RaftProtocol

		keyServer := "server" + strconv.Itoa(i)
		mp[keyServer] = mpServer
	}

	return mp, nil
}

func (n *Nomad) ReadVars() (map[string]interface{}, error) {
	mp := make(map[string]interface{})

	varsObj := n.client.Variables()
	varsInfo, _, err := varsObj.List(nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(varsInfo); i++ {
		mpPathvars := make(map[string]interface{})
		vItems, _, err := varsObj.GetVariableItems(varsInfo[i].Path, nil)
		if err != nil {
			return nil, err
		}

		for k := range vItems {
			mpPathvars[k] = vItems[k]
		}

		mp[varsInfo[i].Path] = mpPathvars
	}

	return mp, nil
}
