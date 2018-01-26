package syslog_acceptance_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/jtarchie/syslog/pkg/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Impact on the local VM", func() {

	AfterEach(func() {
		Cleanup()
	})

	Context("When starting up", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/udp-blackbox.yml")
			AddFakeOldConfig()
			restartsession := BoshCmd("restart", "forwarder")
			Eventually(restartsession).Should(gexec.Exit(0))
		})

		It("Cleans up any file at the old config file location", func() {
			session := ForwarderSshCmd("stat /etc/rsyslog.d/rsyslog.conf")
			Eventually(session).Should(gbytes.Say("stat: cannot stat ‘/etc/rsyslog.d/rsyslog.conf’: No such file or directory"))
		})
	})

	PContext("When processing logs from blackbox", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/udp-blackbox.yml")
		})

		It("doesn't write them to any of the standard linux logfiles", func() {
			By("waiting for logs to be forwarded")
			Eventually(WriteToTestFile("test-logger-isolation")).Should(gbytes.Say("test-logger-isolation"))

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
	TestSharedBehavior := func() {
		Context("When messages are written to UDP with logger", func() {
			It("receives messages in rfc5424 format on the configured drain", func() {

				type LogOutput struct {
					Tables []struct {
						Rows []struct {
							Stdout string
						}
					}
				}

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
				session := ForwarderSshCmd("sudo touch /var/vcap/sys/log/syslog_forwarder/file.log")
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
