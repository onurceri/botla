package integration

import "testing"

func TestTeardownTestEnv_NoPanic(t *testing.T) {
    TeardownTestEnv(nil)
    TeardownTestEnv(&TestEnv{})
}

