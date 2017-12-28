package tests_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jtarchie/syslog/pkg/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Forwarding loglines from files to a TCP syslog drain", func() {
	DeploymentName := func() string {
		return "syslog-tests"
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
		Eventually(session).Should(gexec.Exit(0))
		return session
	}

	ForwarderLog := func() *gexec.Session {
		session := BoshCmd("ssh", "storer", fmt.Sprintf("--command=%q", "cat /var/vcap/store/syslog_storer/syslog.log | grep '47450'"), "--json", "-r")
		Eventually(session).Should(gexec.Exit())
		return session
	}

	SendLogMessage := func(msg string) {
		session := BoshCmd("ssh", "forwarder", "-c", fmt.Sprintf("logger %s", msg))
		Eventually(session).Should(gexec.Exit(0))
	}

	Cleanup := func() {
		session := BoshCmd("delete-deployment")
		Eventually(session).Should(gexec.Exit(0))
	}

	BeforeEach(func() {
		Cleanup()
		Deploy("manifest.yml")
	})

	Context("When a message is written to the default watch dir", func() {

		It("is received in rfc5424 format on the configured drain", func() {
			Eventually(func() *gexec.Session {
				SendLogMessage("test-rfc5424")
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
	})
})
