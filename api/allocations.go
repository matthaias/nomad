package api

import (
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/go-cleanhttp"
)

// Allocations is used to query the alloc-related endpoints.
type Allocations struct {
	client *Client
}

// Allocations returns a handle on the allocs endpoints.
func (c *Client) Allocations() *Allocations {
	return &Allocations{client: c}
}

// List returns a list of all of the allocations.
func (a *Allocations) List(q *QueryOptions) ([]*AllocationListStub, *QueryMeta, error) {
	var resp []*AllocationListStub
	qm, err := a.client.query("/v1/allocations", &resp, q)
	if err != nil {
		return nil, nil, err
	}
	sort.Sort(AllocIndexSort(resp))
	return resp, qm, nil
}

func (a *Allocations) PrefixList(prefix string) ([]*AllocationListStub, *QueryMeta, error) {
	return a.List(&QueryOptions{Prefix: prefix})
}

// Info is used to retrieve a single allocation.
func (a *Allocations) Info(allocID string, q *QueryOptions) (*Allocation, *QueryMeta, error) {
	var resp Allocation
	qm, err := a.client.query("/v1/allocation/"+allocID, &resp, q)
	if err != nil {
		return nil, nil, err
	}
	return &resp, qm, nil
}

func (a *Allocations) Stats(alloc *Allocation, q *QueryOptions) (map[string]*TaskResourceUsage, error) {
	node, _, err := a.client.Nodes().Info(alloc.NodeID, q)
	if err != nil {
		return nil, err
	}
	if node.HTTPAddr == "" {
		return nil, fmt.Errorf("http addr of the node where alloc %q is running is not advertised", alloc.ID)
	}
	client, err := NewClient(&Config{
		Address:    fmt.Sprintf("http://%s", node.HTTPAddr),
		HttpClient: cleanhttp.DefaultClient(),
	})
	if err != nil {
		return nil, err
	}
	resp := make(map[string][]*TaskResourceUsage)
	_, err = client.query("/v1/client/allocation/"+alloc.ID+"/stats", &resp, nil)
	res := make(map[string]*TaskResourceUsage)
	for task, ru := range resp {
		res[task] = ru[0]
	}
	return res, err
}

// Allocation is used for serialization of allocations.
type Allocation struct {
	ID                 string
	EvalID             string
	Name               string
	NodeID             string
	JobID              string
	Job                *Job
	TaskGroup          string
	Resources          *Resources
	TaskResources      map[string]*Resources
	Services           map[string]string
	Metrics            *AllocationMetric
	DesiredStatus      string
	DesiredDescription string
	ClientStatus       string
	ClientDescription  string
	TaskStates         map[string]*TaskState
	CreateIndex        uint64
	ModifyIndex        uint64
	CreateTime         int64
}

// AllocationMetric is used to deserialize allocation metrics.
type AllocationMetric struct {
	NodesEvaluated     int
	NodesFiltered      int
	NodesAvailable     map[string]int
	ClassFiltered      map[string]int
	ConstraintFiltered map[string]int
	NodesExhausted     int
	ClassExhausted     map[string]int
	DimensionExhausted map[string]int
	Scores             map[string]float64
	AllocationTime     time.Duration
	CoalescedFailures  int
}

// AllocationListStub is used to return a subset of an allocation
// during list operations.
type AllocationListStub struct {
	ID                 string
	EvalID             string
	Name               string
	NodeID             string
	JobID              string
	TaskGroup          string
	DesiredStatus      string
	DesiredDescription string
	ClientStatus       string
	ClientDescription  string
	TaskStates         map[string]*TaskState
	CreateIndex        uint64
	ModifyIndex        uint64
	CreateTime         int64
}

// AllocIndexSort reverse sorts allocs by CreateIndex.
type AllocIndexSort []*AllocationListStub

func (a AllocIndexSort) Len() int {
	return len(a)
}

func (a AllocIndexSort) Less(i, j int) bool {
	return a[i].CreateIndex > a[j].CreateIndex
}

func (a AllocIndexSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
