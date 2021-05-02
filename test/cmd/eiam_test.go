package cmd

import (
	"log"
	"os"
	"testing"

	"github.com/spf13/viper"

	testutil "github.com/rigup/ephemeral-iam/test"
)

var allSettings map[string]interface{}

func TestMain(m *testing.M) {
	if err := testutil.InitEiam(); err != nil {
		log.Fatal(err)
	}

	testutil.LoadEnv()

	allSettings = make(map[string]interface{}, len(viper.AllKeys()))
	for _, configKey := range viper.AllKeys() {
		allSettings[configKey] = viper.Get(configKey)
	}

	os.Exit(m.Run())
}
