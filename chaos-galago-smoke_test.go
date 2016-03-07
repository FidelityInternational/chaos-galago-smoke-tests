package main_test

import (
	"crypto/tls"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

var _ = Describe("Assuming chaos-galago is deployed", func() {
	var client *http.Client

	BeforeEach(func() {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	})

	Describe("service instances", func() {
		Context("when the service instance exists", func() {
			var (
				serviceInstanceName = "galago_smoke_test"
				dashboardURL        string
			)

			BeforeEach(func() {
				exec.Command("cf", "create-service", "chaos-galago", "default", serviceInstanceName).Run()
				output, err := exec.Command("cf", "service", serviceInstanceName).Output()
				freakOutDebug(output, err)
				firstSplit := strings.SplitAfter(string(output), "Dashboard: ")[1]
				dashboardURL = strings.TrimSpace(strings.SplitAfter(firstSplit, "\n")[0])
			})

			AfterEach(func() {
				exec.Command("cf", "delete-service", "-f", serviceInstanceName).Run()
			})

			It("creates a service instance", func() {
				output, _ := exec.Command("cf", "create-service", "chaos-galago", "default", serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).To(MatchRegexp("already exists"))
				output, _ = exec.Command("cf", "services").Output()
				Expect(string(output)).To(MatchRegexp(serviceInstanceName))
			})

			It("updates a service instance", func() {
				resp, _ := client.PostForm(dashboardURL, url.Values{"probability": {"1"}, "frequency": {"1"}})
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				Expect(string(body)).To(MatchRegexp("Probability: 1"))
				Expect(string(body)).To(MatchRegexp("Frequency: 1"))
			})

			It("deletes a service instance", func() {
				output, _ := exec.Command("cf", "delete-service", "-f", serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).ToNot(MatchRegexp("does not exist"))
				output, _ = exec.Command("cf", "services").Output()
				Expect(string(output)).ToNot(MatchRegexp(serviceInstanceName))
			})
		})

		Context("When the service instance does not exist", func() {
			var serviceInstanceName = "galago_smoke_test"

			AfterEach(func() {
				exec.Command("cf", "delete-service", "-f", serviceInstanceName).Run()
			})

			It("creates a service instance", func() {
				output, _ := exec.Command("cf", "create-service", "chaos-galago", "default", serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).ToNot(MatchRegexp("already exists"))
				output, _ = exec.Command("cf", "services").Output()
				Expect(string(output)).To(MatchRegexp(serviceInstanceName))
			})

			It("deletes a service instance", func() {
				output, _ := exec.Command("cf", "delete-service", "-f", serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).To(MatchRegexp("does not exist"))
				output, _ = exec.Command("cf", "services").Output()
				Expect(string(output)).ToNot(MatchRegexp(serviceInstanceName))
			})
		})

	})

	Describe("service bindings", func() {
		var (
			serviceInstanceName = "galago_smoke_test"
			appName             = "galago_smoke_test"
		)

		BeforeEach(func() {
			exec.Command("cf", "create-service", "chaos-galago", "default", serviceInstanceName).Run()
			exec.Command("cf", "push", "-f", "fixtures/galago_smoke_test/manifest.yml", "--no-start").Run()
		})

		AfterEach(func() {
			exec.Command("cf", "delete", "-f", appName).Run()
			exec.Command("cf", "delete-service", "-f", serviceInstanceName).Run()
		})

		Context("when an app is bound", func() {
			BeforeEach(func() {
				exec.Command("cf", "bind-service", appName, serviceInstanceName).Run()
			})

			It("bind a service instance", func() {
				output, _ := exec.Command("cf", "bind-service", appName, serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).To(MatchRegexp("already bound"))
				output, _ = exec.Command("cf", "env", appName).Output()
				Expect(string(output)).To(MatchRegexp(`label": "chaos-galago"`))
				Expect(string(output)).To(MatchRegexp(`frequency": 5`))
				Expect(string(output)).To(MatchRegexp(`probability": 0.2`))
			})

			It("unbind a service instance", func() {
				output, _ := exec.Command("cf", "unbind-service", appName, serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).ToNot(MatchRegexp("did not exist"))
				output, _ = exec.Command("cf", "env", appName).Output()
				Expect(string(output)).ToNot(MatchRegexp(`label": "chaos-galago"`))
				Expect(string(output)).ToNot(MatchRegexp(`frequency": 5`))
				Expect(string(output)).ToNot(MatchRegexp(`probability": 0.2`))
			})

			Context("the processor", func() {
				var (
					appGUID      string
					dashboardURL string
				)

				BeforeEach(func() {
					output, err := exec.Command("cf", "app", appName, "--guid").Output()
					freakOutDebug(output, err)
					appGUID = strings.TrimSpace(string(output))
					output, err = exec.Command("cf", "service", serviceInstanceName).Output()
					freakOutDebug(output, err)
					firstSplit := strings.SplitAfter(string(output), "Dashboard: ")[1]
					dashboardURL = strings.TrimSpace(strings.SplitAfter(firstSplit, "\n")[0])
					client.PostForm(dashboardURL, url.Values{"probability": {"1"}, "frequency": {"1"}})
					exec.Command("cf", "start", appName).Run()
					exec.Command("cf", "target", "-o", "chaos-galago", "-s", "chaos-galago").Run()
				})

				AfterEach(func() {
					exec.Command("cf", "target", "-o", orgName, "-s", spaceName).Run()
				})

				It("acts on bound aplications", func() {
					Eventually(func() string {
						output, _ := exec.Command("cf", "curl", fmt.Sprintf("v2/apps/%s/instances", appGUID)).Output()
						return string(output)
					}, "50s", "5s").Should(MatchRegexp(`"state": "RUNNING"`))
					Eventually(func() string {
						output, _ := exec.Command("cf", "logs", "chaos-galago-processor", "--recent").Output()
						return string(output)
					}, "120s", "10s").Should(MatchRegexp(fmt.Sprintf("About to kill app instance: %s at index: 0", appGUID)))
					Eventually(func() string {
						output, _ := exec.Command("cf", "curl", fmt.Sprintf("v2/apps/%s/instances", appGUID)).Output()
						return string(output)
					}, "30s", "3s").Should(MatchRegexp(`"state": "DOWN"`))
				})
			})
		})

		Context("when an app is not bound", func() {
			It("bind a service instance", func() {
				output, _ := exec.Command("cf", "bind-service", appName, serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).ToNot(MatchRegexp("already bound"))
				output, _ = exec.Command("cf", "env", appName).Output()
				Expect(string(output)).To(MatchRegexp(`label": "chaos-galago"`))
				Expect(string(output)).To(MatchRegexp(`frequency": 5`))
				Expect(string(output)).To(MatchRegexp(`probability": 0.2`))
			})

			It("unbind a service instance", func() {
				output, _ := exec.Command("cf", "unbind-service", appName, serviceInstanceName).Output()
				Expect(string(output)).To(MatchRegexp("OK"))
				Expect(string(output)).To(MatchRegexp("did not exist"))
				output, _ = exec.Command("cf", "env", appName).Output()
				Expect(string(output)).ToNot(MatchRegexp(`label": "chaos-galago"`))
				Expect(string(output)).ToNot(MatchRegexp(`frequency": 5`))
				Expect(string(output)).ToNot(MatchRegexp(`probability": 0.2`))
			})
		})
	})
})
