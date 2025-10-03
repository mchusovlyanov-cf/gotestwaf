package statistic

import (
	"sort"
)

type Statistics struct {
	IsGrpcAvailable    bool
	IsGraphQLAvailable bool

	Paths ScannedPaths

	TestCasesFingerprint string

	TruePositiveTests TestsSummary
	TrueNegativeTests TestsSummary

	Score struct {
		ApiSec  Score
		AppSec  Score
		Average float64
	}
}

type TestsSummary struct {
	SummaryTable []*SummaryTableRow
	Blocked      []*TestDetails
	Bypasses     []*TestDetails
	Unresolved   []*TestDetails
	Failed       []*FailedDetails

	ReqStats       RequestStats
	ApiSecReqStats RequestStats
	AppSecReqStats RequestStats

	UnresolvedRequestsPercentage       float64
	ResolvedBlockedRequestsPercentage  float64
	ResolvedBypassedRequestsPercentage float64
	FailedRequestsPercentage           float64
}

type SummaryTableRow struct {
	TestSet    string  `json:"test_set" validate:"required,printascii,max=256"`
	TestCase   string  `json:"test_case" validate:"required,printascii,max=256"`
	Percentage float64 `json:"percentage" validate:"min=0,max=100"`
	Sent       int     `json:"sent" validate:"min=0"`
	Blocked    int     `json:"blocked" validate:"min=0"`
	Bypassed   int     `json:"bypassed" validate:"min=0"`
	Unresolved int     `json:"unresolved" validate:"min=0"`
	Failed     int     `json:"failed" validate:"min=0"`
}

type TestDetails struct {
	Payload            string
	TestCase           string
	TestSet            string
	Encoder            string
	Placeholder        string
	ResponseStatusCode int
	AdditionalInfo     []string
	Type               string
}

type FailedDetails struct {
	Payload     string   `json:"payload" validate:"required"`
	TestCase    string   `json:"test_case" validate:"required,printascii"`
	TestSet     string   `json:"test_set" validate:"required,printascii"`
	Encoder     string   `json:"encoder" validate:"required,printascii"`
	Placeholder string   `json:"placeholder" validate:"required,printascii"`
	Reason      []string `json:"reason" validate:"omitempty,dive,required"`
	Type        string   `json:"type" validate:"omitempty"`
}

type RequestStats struct {
	AllRequestsNumber        int
	BlockedRequestsNumber    int
	BypassedRequestsNumber   int
	UnresolvedRequestsNumber int
	FailedRequestsNumber     int
	ResolvedRequestsNumber   int
}

type Score struct {
	TruePositive float64
	TrueNegative float64
	Average      float64
}

type Path struct {
	Method string `json:"method" validate:"required,printascii,max=32"`
	Path   string `json:"path" validate:"required,printascii,max=1024"`
}

type ScannedPaths []*Path

var _ sort.Interface = (ScannedPaths)(nil)

func (sp ScannedPaths) Len() int {
	return len(sp)
}

func (sp ScannedPaths) Less(i, j int) bool {
	if sp[i].Path > sp[j].Path {
		return false
	} else if sp[i].Path < sp[j].Path {
		return true
	}

	return sp[i].Method < sp[j].Method
}

func (sp ScannedPaths) Swap(i, j int) {
	sp[i], sp[j] = sp[j], sp[i]
}

func (sp ScannedPaths) Sort() {
	sort.Sort(sp)
}
