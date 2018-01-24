package tests_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jtarchie/syslog/pkg/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Impact on the local VM", func() {
	DeploymentName := func() string {
		return fmt.Sprintf("syslog-tests-%d", GinkgoParallelNode())
	}

	BoshCmd := func(args ...string) *gexec.Session {
		boshArgs := []string{"-n", "-d", DeploymentName()}
		boshArgs = append(boshArgs, args...)
		boshCmd := exec.Command("bosh", boshArgs...)
		By("Performing command: bosh " + strings.Join(boshArgs, " "))
		session, err := gexec.Start(boshCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		return session
	}

	SendLogMessage := func(msg string) {
		session := BoshCmd("ssh", "forwarder", "-c", fmt.Sprintf("logger %s -t vcap.", msg))
		Eventually(session).Should(gexec.Exit(0))
	}

	Cleanup := func() {
		session := BoshCmd("delete-deployment")
		Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	}

	Deploy := func(manifest string) *gexec.Session {
		session := BoshCmd("deploy", manifest, "-v", fmt.Sprintf("deployment=%s", DeploymentName()))
		Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		return session
	}

	ForwarderLog := func() *gexec.Session {
		// 47450 is CF's "enterprise ID" and uniquely identifies messages sent by our system
		session := BoshCmd("ssh", "storer", fmt.Sprintf("--command=%q", "cat /var/vcap/store/syslog_storer/syslog.log | grep '47450'"), "--json", "-r")
		Eventually(session).Should(gexec.Exit())
		return session
	}

	WriteToTestFile := func(message string) func() *gexec.Session {
		return func() *gexec.Session {
			session := BoshCmd("ssh", "forwarder", "-c", fmt.Sprintf("echo %s | sudo tee -a /var/vcap/sys/log/syslog_forwarder/file.log", message))
			Eventually(session).Should(gexec.Exit(0))
			return ForwarderLog()
		}
	}

	DefaultLogfiles := func() *gexec.Session {
		session := BoshCmd("ssh", "forwarder", fmt.Sprintf("--command=%q", "sudo cat /var/log/{messages,syslog,user.log}"), "--json", "-r")
		Eventually(session).Should(gexec.Exit())
		return session
	}

	PContext("When starting up", func() {
		It("Cleans up any file at the old config file location", func() {
		})
	})

	PContext("When processing logs from blackbox", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/udp-blackbox.yml")
			WriteToTestFile("test-logger-isolation")
		})

		It("doesn't write them to any of the standard linux logfiles", func() {
			By("waiting for logs to be forwarded")
			Eventually(func() *gexec.Session {
				return ForwarderLog()
			}).Should(gbytes.Say("test-logger-isolation"))

			By("checking that the logs don't appear in local logfiles")
			Expect(DefaultLogfiles()).NotTo(gbytes.Say("test-logger-isolation"))
		})
	})

	Context("When processing forwarded logs from jobs using logger", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/udp-blackbox.yml")
			SendLogMessage("test-logger-isolation")
		})

		It("doesn't write them to the logfiles specified in the stemcell config", func() {

			By("waiting for logs to be forwarded")
			Eventually(func() *gexec.Session {
				return ForwarderLog()
			}).Should(gbytes.Say("test-logger-isolation"))

			By("checking that the logs don't appear in local logfiles")
			Expect(DefaultLogfiles()).NotTo(gbytes.Say("test-logger-isolation"))
		})
	})
})

var _ = Describe("Forwarding loglines to a TCP syslog drain", func() {
	DeploymentName := func() string {
		return fmt.Sprintf("syslog-tests-%d", GinkgoParallelNode())
	}

	BoshCmd := func(args ...string) *gexec.Session {
		boshArgs := []string{"-n", "-d", DeploymentName()}
		boshArgs = append(boshArgs, args...)
		boshCmd := exec.Command("bosh", boshArgs...)
		By("Performing command: bosh " + strings.Join(boshArgs, " "))
		session, err := gexec.Start(boshCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		return session
	}

	type LogOutput struct {
		Tables []struct {
			Rows []struct {
				Stdout string
			}
		}
	}

	Deploy := func(manifest string) *gexec.Session {
		session := BoshCmd("deploy", manifest, "-v", fmt.Sprintf("deployment=%s", DeploymentName()))
		Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		return session
	}

	ForwarderLog := func() *gexec.Session {
		// 47450 is CF's "enterprise ID" and uniquely identifies messages sent by our system
		session := BoshCmd("ssh", "storer", fmt.Sprintf("--command=%q", "cat /var/vcap/store/syslog_storer/syslog.log | grep '47450'"), "--json", "-r")
		Eventually(session).Should(gexec.Exit())
		return session
	}

	SendLogMessage := func(msg string) {
		session := BoshCmd("ssh", "forwarder", "-c", fmt.Sprintf("logger %s -t vcap.", msg))
		Eventually(session).Should(gexec.Exit(0))
	}

	WriteToTestFile := func(message string) func() *gexec.Session {
		return func() *gexec.Session {
			session := BoshCmd("ssh", "forwarder", "-c", fmt.Sprintf("echo %s | sudo tee -a /var/vcap/sys/log/syslog_forwarder/file.log", message))
			Eventually(session).Should(gexec.Exit(0))
			return ForwarderLog()
		}
	}

	Cleanup := func() {
		session := BoshCmd("delete-deployment")
		Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	}

	TestSharedBehavior := func() {
		Context("When messages are written to UDP with logger", func() {
			It("receives messages in rfc5424 format on the configured drain", func() {
				SendLogMessage("test-rfc5424")
				Eventually(func() *gexec.Session {
					return ForwarderLog()
				}).Should(gbytes.Say("test-rfc5424"))

				output := LogOutput{}
				err := json.Unmarshal(ForwarderLog().Out.Contents(), &output)
				Expect(err).ToNot(HaveOccurred())

				logs := bytes.NewBufferString(output.Tables[0].Rows[0].Stdout)
				reader := bufio.NewReader(logs)

				for {
					line, _, err := reader.ReadLine()
					Expect(err).ToNot(HaveOccurred())
					if len(line) == 0 {
						break
					}
					logLine, err := syslog.Parse(line)
					Expect(err).ToNot(HaveOccurred())
					if string(logLine.Message()) == "test-rfc5424" {
						sdata := logLine.StructureData()
						Expect(string(sdata.ID())).To(Equal("instance@47450"))
						break
					}
				}
			})

			It("receives messages over 1k long on the configured drain", func() {
				message := counterString(1025, "A")
				SendLogMessage(message)
				Eventually(func() *gexec.Session {
					return ForwarderLog()
				}).Should(gbytes.Say(message))
			})
		})

		Context("when a file is created in the watched directory structure", func() {
			BeforeEach(func() {
				session := BoshCmd("ssh", "forwarder", "-c", "sudo touch /var/vcap/sys/log/syslog_forwarder/file.log")
				Eventually(session).Should(gexec.Exit(0))
			})

			It("forwards new lines written to the file through syslog", func() {
				Eventually(WriteToTestFile("test-blackbox-forwarding")).Should(gbytes.Say("test-blackbox-forwarding"))
			})
		})
	}

	Context("when file forwarding is configured to use UDP", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/udp-blackbox.yml")
		})
		AfterEach(func() {
			Cleanup()
		})

		TestSharedBehavior()
	})

	Context("when file forwarding is configured to use TCP", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/tcp-blackbox.yml")
		})
		AfterEach(func() {
			Cleanup()
		})

		TestSharedBehavior()
		It("fowards messages of over 1KB", func() {
			message := counterString(1025, "A")

			Eventually(WriteToTestFile(message)).Should(gbytes.Say(message))
		})
	})
})

func counterString(l int, s string) string {
	counterstring := ""
	for len(counterstring) < l {
		counterstring += s
		counterstring += strconv.Itoa(len(counterstring))
	}

	return counterstring[:l]
}
