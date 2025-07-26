package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("GC_SECRET_KEY", "secret")
	t.Setenv("GC_SECRET_ID", "id")
	t.Setenv("YNAB_TOKEN", "token")
	t.Setenv("JOBS", "GC_ACCOUNT_ID,YNAB_BUDGET_ID,YNAB_ACCOUNT_ID|GC_ACCOUNT_ID2,YNAB_BUDGET_ID2,YNAB_ACCOUNT_ID2")
	c, err := LoadConfigFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, "secret", c.GCSecretKey)
	assert.Equal(t, "id", c.GCSecretID)
	assert.Equal(t, "token", c.YNABToken)
}
