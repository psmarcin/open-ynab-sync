package main

import (
	"fmt"
	"strings"
)

type job struct {
	GCAccountID   string
	YNABAccountID string
	YNABBudgetID  string
}

// envToJobs parses a delimited string to construct a slice of job structs or returns an error for invalid input format.
// example source: GCAccountID1,YNABBudgetID1,YNABAccountID1|GCAccountID2,YNABBudgetID2,YNABAccountID2|...
func envToJobs(source string) (jobs []job, err error) {
	if source == "" {
		return nil, fmt.Errorf("empty source string")
	}

	jobConfigs := strings.Split(source, "|")
	jobs = make([]job, 0, len(jobConfigs))

	for _, cfg := range jobConfigs {
		parts := strings.Split(cfg, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid job configuration: %s", cfg)
		}

		jobs = append(jobs, job{
			GCAccountID:   strings.TrimSpace(parts[0]),
			YNABBudgetID:  strings.TrimSpace(parts[1]),
			YNABAccountID: strings.TrimSpace(parts[2]),
		})
	}

	return jobs, nil
}
