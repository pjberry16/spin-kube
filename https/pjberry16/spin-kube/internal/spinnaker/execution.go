package spinnaker

import "github.com/spinnaker/spin/cmd/gateclient"

type Execution struct {
	gc          *gateclient.GatewayClient
	Application string              `json:"application"`
	Name        string              `json:"name"`
	BuildTime   int64               `json:"buildTime"`
	Stages      []Stage             `json:"stages"`
	StartTime   int64               `json:"startTime"`
	EndTime     int64               `json:"endTime"`
	Status      string              `json:"status"`
	Trigger     Trigger             `json:"trigger"`
	Variables   []ExecutionVariable `json:"variables"`
}

type ExecutionVariable struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Trigger struct {
	Parameters Parameters `json:"parameters"`
	Type       string     `json:"type"`
	User       string     `json:"user"`
}

type Parameters struct {
	CR                string `json:"Service Now Record"`
	ParentExecutionId string `json:"parentExecutionId"`
}

type Stage struct {
	Type    string       `json:"type"`
	Name    string       `json:"name"`
	Context StageContext `json:"context,omitempty"`
	Status  string       `json:"status"`
}

type StageContext struct {
	Account   string         `json:"account"`
	Version   string         `json:"version"`
	Kube      string         `json:"kube"`
	Errors    []interface{}  `json:"errors"`
	Target    string         `json:"target,omitempty"` // Target contains the store name to deploy a kube into.
	Clusters  []StageCluster `json:"clusters,omitempty"`
	Exception Exception      `json:"exception"`
}

type Exception struct {
	Details Details `json:"details"`
}

type Details struct {
	Errors []string `json:"errors"`
}

type StageCluster struct {
	Account     string `json:"account"`
	Application string `json:"application"`
}
