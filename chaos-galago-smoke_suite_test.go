package main_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"os"
	"os/exec"
	"testing"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	doSetup()
	RunSpecs(t, "main test suite")
	doTeardown()
}

var (
	cfHome              string
	err                 error
	guid                = uuid.NewV4()
	orgName             = fmt.Sprintf("chaos-galago-smoke-%s", guid)
	spaceName           = orgName
	output              []byte
	serviceInstanceName = fmt.Sprintf("galago_smoke_test_%s", guid)
	appName             = fmt.Sprintf("galago_smoke_test_%s", guid)
)

func doSetup() {
	cfHome = os.Getenv("CF_HOME")
	isEnvSet("CF_HOME", cfHome)
	if _, homeErr := os.Stat(fmt.Sprintf("%s/.cf", cfHome)); homeErr == nil {
		fmt.Println("CF_HOME must be a temp directory")
		os.Exit(1)
	}

	cfPassword := os.Getenv("CF_PASSWORD")
	cfUsername := os.Getenv("CF_USERNAME")
	cfDomain := os.Getenv("CF_DOMAIN")
	isEnvSet("CF_PASSWORD", cfPassword)
	isEnvSet("CF_USERNAME", cfUsername)
	isEnvSet("CF_DOMAIN", cfDomain)

	output, err = exec.Command("cf", "login", "-a", fmt.Sprintf("https://api.%s", cfDomain), "-u", cfUsername, "-p", cfPassword, "--skip-ssl-validation").Output()
	freakOutDebug(output, err)
	output, err = exec.Command("cf", "create-org", orgName).Output()
	freakOutDebug(output, err)
	output, err = exec.Command("cf", "target", "-o", orgName).Output()
	freakOutDebug(output, err)
	output, err = exec.Command("cf", "create-space", spaceName).Output()
	freakOutDebug(output, err)
	output, err = exec.Command("cf", "target", "-s", spaceName).Output()
	freakOutDebug(output, err)
	output, err = exec.Command("cf", "push", "-f", "fixtures/galago_smoke_test/manifest.yml").Output()
	freakOutDebug(output, err)
}

func doTeardown() {
	output, err = exec.Command("cf", "delete", "-f", appName).Output()
	freakOutDebug(output, err)
	output, err = exec.Command("cf", "delete-space", "-f", spaceName).Output()
	freakOutDebug(output, err)
	output, err = exec.Command("cf", "delete-org", "-f", orgName).Output()
	freakOutDebug(output, err)
	os.RemoveAll(fmt.Sprintf("%s/.cf", cfHome))
}

func isEnvSet(envName string, env string) {
	if env == "" {
		fmt.Printf("\n\nEnvironment Variable \"%s\" was not set\n", envName)
		os.Exit(1)
	}
}

func freakOutDebug(output []byte, err error) {
	if err != nil {
		fmt.Printf("\nAn unexpected error occurred: \n%s\n", output)
	}
}
