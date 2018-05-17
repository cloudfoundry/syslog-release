package syslog_acceptance_test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func DeploymentName() string {
	return fmt.Sprintf("syslog-tests-%d", GinkgoParallelNode())
}

func BoshCmd(args ...string) *gexec.Session {
	boshArgs := []string{"-n", "-d", DeploymentName()}
	boshArgs = append(boshArgs, args...)
	boshCmd := exec.Command("bosh", boshArgs...)
	By("Performing command: bosh " + strings.Join(boshArgs, " "))
	session, err := gexec.Start(boshCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	return session
}

func ForwarderSshCmd(command string) *gexec.Session {
	return BoshCmd("ssh", "forwarder", "-c", command)
}

func SendLogMessage(msg string) {
	session := ForwarderSshCmd(fmt.Sprintf("logger %s -t vcap.", msg))
	Eventually(session).Should(gexec.Exit(0))
}

func Cleanup() {
	BoshCmd("locks")
	session := BoshCmd("delete-deployment")
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	Eventually(BoshCmd("locks")).ShouldNot(gbytes.Say(DeploymentName()))
}

func Deploy(manifest string) *gexec.Session {
	session := BoshCmd("deploy", manifest, "-v", fmt.Sprintf("deployment=%s", DeploymentName()))
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	Eventually(BoshCmd("locks")).ShouldNot(gbytes.Say(DeploymentName()))
	return session
}

func DeployWithVarsStore(manifest string) *gexec.Session {
	session := BoshCmd("deploy", manifest, "-v", fmt.Sprintf("deployment=%s", DeploymentName()), fmt.Sprintf("--vars-store=/tmp/%s-vars.yml", DeploymentName()))
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	Eventually(BoshCmd("locks")).ShouldNot(gbytes.Say(DeploymentName()))
	return session
}

func ForwarderLog() *gexec.Session {
	// 47450 is CF's "enterprise ID" and uniquely identifies messages sent by our system
	session := BoshCmd("ssh", "storer", fmt.Sprintf("--command=%q", "cat /var/vcap/store/syslog_storer/syslog.log | grep '47450'"), "--json", "-r")
	Eventually(session).Should(gexec.Exit())
	return session
}

type LogOutput struct {
	Tables []struct {
		Rows []struct {
			Stdout string
		}
	}
}

func ForwardedLogs() string {
	// 47450 is CF's "enterprise ID" and uniquely identifies messages sent by our system
	return OutputFromBoshCommand("storer", "cat /var/vcap/store/syslog_storer/syslog.log | grep '47450'")
}

func ForwarderMonitSummary() string {
	return OutputFromBoshCommand("forwarder", "sudo /var/vcap/bash/bin/monit summary")
}

func OutputFromBoshCommand(job, command string) string {
	// 47450 is CF's "enterprise ID" and uniquely identifies messages sent by our system
	session := BoshCmd("ssh", job, "--command="+command, "--json", "-r")
	Eventually(session).Should(gexec.Exit())
	stdoutContents := session.Out.Contents()
	var logOutput LogOutput
	err := json.Unmarshal(stdoutContents, &logOutput)
	Expect(err).ToNot(HaveOccurred())
	return logOutput.Tables[0].Rows[0].Stdout
}

func AddFakeOldConfig() {
	By("Adding files all the places where the config used to live")
	session := ForwarderSshCmd("sudo bash -c 'echo fakeConfig=true > /etc/rsyslog.d/rsyslog.conf && echo fakeConfig=true > /etc/rsyslog.d/30-syslog-release.conf && echo fakeConfig=true > /etc/rsyslog.d/20-syslog-release-custom-rules.conf'")
	Eventually(session, 5*time.Minute).Should(gexec.Exit(0))
}

func WriteToTestFile(message string) func() *gexec.Session {
	return func() *gexec.Session {
		session := ForwarderSshCmd(fmt.Sprintf("echo %s | sudo tee -a /var/vcap/sys/log/syslog_forwarder/file.log", message))
		Eventually(session).Should(gexec.Exit(0))
		return ForwarderLog()
	}
}

func WriteToPrivateTestFile(message string) func() *gexec.Session {
	return func() *gexec.Session {
		session := ForwarderSshCmd(fmt.Sprintf("sudo bash -c '"+
			"touch /var/vcap/sys/log/syslog_forwarder/file.log; "+
			"chmod 0700 /var/vcap/sys/log/syslog_forwarder/file.log; "+
			"echo %s >> /var/vcap/sys/log/syslog_forwarder/file.log'", message))
		Eventually(session).Should(gexec.Exit(0))
		return ForwarderLog()
	}
}

func DefaultLogfiles() *gexec.Session {
	session := BoshCmd("ssh", "forwarder", fmt.Sprintf("--command=%q", "sudo cat /var/log/{messages,syslog,user.log}"), "--json", "-r")
	Eventually(session).Should(gexec.Exit())
	return session
}
