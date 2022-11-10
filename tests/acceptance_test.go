package syslog_acceptance_test

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"time"

	logrfc "github.com/jtarchie/syslog/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var DEFAULT_LOGGER_SIZE = 1024

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

		It("Cleans up any files at old config file locations", func() {
			session := ForwarderSshCmd("stat /etc/rsyslog.d/rsyslog.conf || stat /etc/rsyslog.d/30-syslog-release.conf || stat /etc/rsyslog.d/20-syslog-release-custom-rules.conf")
			Eventually(session).Should(gbytes.Say(`stat: cannot stat[x]? '\/etc\/rsyslog\.d\/rsyslog\.conf'`))
			Eventually(session).Should(gbytes.Say(`stat: cannot stat[x]? '\/etc\/rsyslog\.d\/30-syslog-release\.conf'`))
			Eventually(session).Should(gbytes.Say(`stat: cannot stat[x]? '\/etc\/rsyslog\.d\/20-syslog-release-custom-rules\.conf'`))
		})
	})

	Context("When processing logs from blackbox", func() {
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
			SendLogMessage("test-logger-isolation", DEFAULT_LOGGER_SIZE)
		})

		It("doesn't write them to the logfiles specified in the stemcell config", func() {

			By("waiting for logs to be forwarded")
			Eventually(func() string {
				return ForwardedLogs()
			}).Should(ContainSubstring("test-logger-isolation"))

			By("checking that the logs don't appear in local logfiles")
			Expect(DefaultLogfiles()).NotTo(gbytes.Say("test-logger-isolation"))
		})
	})
})

var _ = Describe("Forwarding loglines to a TCP syslog drain", func() {
	TestSharedBehavior := func() {
		Context("When messages are written to UDP with logger", func() {
			It("receives messages in rfc5424 format on the configured drain", func() {
				SendLogMessage("test-rfc5424", DEFAULT_LOGGER_SIZE)
				Eventually(func() string {
					return ForwardedLogs()
				}).Should(ContainSubstring("test-rfc5424"))

				logs := bytes.NewBufferString(ForwardedLogs())
				reader := bufio.NewReader(logs)

				for {
					line, _, err := reader.ReadLine()
					Expect(err).ToNot(HaveOccurred())
					if len(line) == 0 {
						break
					}
					logLine, _, err := logrfc.Parse(line)
					Expect(err).ToNot(HaveOccurred())
					if string(logLine.Message()) == "test-rfc5424" {
						sdata := logLine.StructureData()[0]
						Expect(string(sdata.ID())).To(Equal("instance@47450"))
						properties := sdata.Properties()
						Expect(properties).To(ContainElement(logrfc.Property{Key: ("director"), Value: ("")}))
						Expect(properties).To(ContainElement(logrfc.Property{Key: ("deployment"), Value: (DeploymentName())}))
						Expect(properties).To(ContainElement(logrfc.Property{Key: ("group"), Value: ("forwarder")}))
						break
					}
				}
			})

			It("receives messages over 1k long on the configured drain", func() {
				message := counterString(1025, "A")
				SendLogMessage(message, 1025)
				Eventually(func() string {
					return ForwardedLogs()
				}).Should(ContainSubstring(message))
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

		It("has a valid config", func() {
			session := ForwarderSshCmd("sudo rsyslogd -N1")
			Eventually(session).Should(gexec.Exit(0))
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

		It("receives truncated messages at max syslog size of 2k", func() {
			message := counterString(3000, "C")
			SendLogMessage(message, 8192)

			Consistently(
				ForwardedLogs,
			).ShouldNot(ContainSubstring(message[2048:]))
		})
	})

	Context("when file forwarding is configured with good rules", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/good-rules.yml")
		})
		AfterEach(func() {
			Cleanup()
		})

		It("filters out messages that match the rule", func() {
			message := "This is a DEBUG message that we filter out"
			SendLogMessage(message, DEFAULT_LOGGER_SIZE)
			Consistently(func() string {
				return ForwardedLogs()
			}).ShouldNot(ContainSubstring(message))
		})

		TestSharedBehavior()
	})

	Context("when an invalid configuration is supplied", func() {
		BeforeEach(func() {
			Cleanup()
		})
		AfterEach(func() {
			Cleanup()
		})

		It("will fail the pre-start script", func() {
			By("Deploying")

			session := BoshCmd("deploy", "manifests/broken-rules.yml",
				"-v", fmt.Sprintf("deployment=%s", DeploymentName()),
				"-v", fmt.Sprintf("stemcell-os=%s", StemcellOS()))
			Eventually(session, 10*time.Minute).Should(gexec.Exit(1))
			Eventually(BoshCmd("locks")).ShouldNot(gbytes.Say(DeploymentName()))
		})
	})

	Context("when TLS is configured", func() {
		BeforeEach(func() {
			Cleanup()
			DeployWithVarsStore("manifests/tls-forwarding.yml")
		})
		AfterEach(func() {
			Cleanup()
		})

		TestSharedBehavior()
	})

	Context("when TLS is configured over relp", func() {
		BeforeEach(func() {
			Cleanup()
			DeployWithVarsStore("manifests/relp-tls.yml")
		})
		AfterEach(func() {
			Cleanup()
		})

		TestSharedBehavior()
	})

	Context("when TLS is configured and mTLS is enforced", func() {
		BeforeEach(func() {
			Cleanup()
			DeployWithVarsStore("manifests/tls-forwarding-mtls.yml")
		})
		AfterEach(func() {
			Cleanup()
		})

		TestSharedBehavior()
	})
})

var _ = Describe("Optional features to reduce CF log volume", func() {
	AfterEach(func() {
		Cleanup()
	})

	Context("when vcap filtering is enabled to eliminate duplication", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/vcap-filtering.yml")
		})
		It("filters logs from the vcap side of the tee while forwarding other logs", func() {
			By("continuing to forward logs from the filesystem")
			fileMessage := "Old-style CF tee-based message, file side"
			Eventually(WriteToTestFile(fileMessage)).Should(gbytes.Say(fileMessage))

			By("not forwarding logs written as vcap. user")
			loggerMessage := "Old-style CF tee-based message, logger side"
			SendLogMessage(loggerMessage, DEFAULT_LOGGER_SIZE)
			Consistently(func() string {
				return ForwardedLogs()
			}).ShouldNot(ContainSubstring(loggerMessage))
		})
	})
	Context("when DEBUG filtering is enabled to reduce volume", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/debug-filtering.yml")
		})
		It("filters logs that start with DEBUG while forwarding other logs", func() {
			By("continuing to forward logs from the filesystem")
			normalMessage := "INFO is not debug or DEBUG"
			Eventually(WriteToTestFile(normalMessage)).Should(gbytes.Say(normalMessage))

			By("not forwarding logs that start with DEBUG")
			debugMessage := "DEBUG is debug, however"
			SendLogMessage(debugMessage, DEFAULT_LOGGER_SIZE)
			Consistently(func() string {
				return ForwardedLogs()
			}).ShouldNot(ContainSubstring(debugMessage))
		})
	})
})

var _ = Describe("When syslog is configured to run in unprivileged mode", func() {
	BeforeEach(func() {
		Cleanup()
		Deploy("manifests/blackbox-unpriv.yml")
	})
	AfterEach(func() {
		Cleanup()
	})

	It("forwards normal log lines", func() {
		messagegenerator := "{1..10}"
		message := "1 2 3 4 5 6 7 8 9 10"
		Eventually(WriteToTestFile(messagegenerator)).Should(gbytes.Say(message))
	})

	It("does not forward logs not visable to syslog user", func() {
		//we can't use a literal message or we'll get false matches from the command which created the log line
		messagegenerator := "{1..10}"
		message := "1 2 3 4 5 6 7 8 9 10"
		Consistently(WriteToPrivateTestFile(messagegenerator)).ShouldNot(gbytes.Say(message))
	})
})

var _ = Describe("When syslog is disabled", func() {
	Context("when the storer is configured to accept logs", func() {
		BeforeEach(func() {
			Cleanup()
			Deploy("manifests/disabled.yml")
		})

		It("does not forward logs", func() {
			SendLogMessage("test-logger-isolation", DEFAULT_LOGGER_SIZE)
			Consistently(func() string {
				return ForwardedLogs()
			}).Should(HaveLen(0))
		})

		It("doesn't start blackbox", func() {
			Expect(ForwarderMonitSummary()).ToNot(ContainSubstring("blackbox"))
		})
	})

	Context("when there is not configuration provided via links or properties", func() {
		BeforeEach(func() {
			Cleanup()
		})

		It("starts successfully", func() {
			Deploy("manifests/disabled-no-config.yml")
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
