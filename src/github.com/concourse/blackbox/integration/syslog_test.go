package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/concourse/blackbox/integration"

	sl "github.com/ziutek/syslog"

	"github.com/concourse/blackbox"
	"github.com/concourse/blackbox/syslog"
)

var _ = Describe("Blackbox", func() {
	var (
		logDir  string
		tagName string
		logFile *os.File
	)

	BeforeEach(func() {
		var err error
		logDir, err = ioutil.TempDir("", "syslog-test")
		Expect(err).NotTo(HaveOccurred())

		tagName = "test-tag"
		err = os.Mkdir(filepath.Join(logDir, tagName), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		logFile, err = os.OpenFile(
			filepath.Join(logDir, tagName, "tail.log"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		logFile.Close()

		err := os.RemoveAll(logDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the syslog server is already running", func() {
		var (
			syslogServer   *SyslogServer
			blackboxRunner *BlackboxRunner
			inbox          *Inbox
		)

		BeforeEach(func() {
			inbox = NewInbox()
			syslogServer = NewSyslogServer(inbox)
			syslogServer.Start()

			blackboxRunner = NewBlackboxRunner(blackboxPath)
		})

		buildConfigHostname := func(hostname string, dirToWatch string) blackbox.Config {
			return blackbox.Config{
				Hostname: hostname,
				Syslog: blackbox.SyslogConfig{
					Destination: syslog.Drain{
						Transport: "udp",
						Address:   syslogServer.Addr,
					},
					SourceDir: dirToWatch,
				},
			}
		}

		buildConfig := func(dirToWatch string) blackbox.Config {
			return buildConfigHostname("", dirToWatch)
		}

		AfterEach(func() {
			syslogServer.Stop()
		})

		It("logs any new lines of a file in source directory to syslog with subdirectory name as tag", func() {
			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 1)

			logFile.WriteString("hello\n")
			logFile.WriteString("world\n")
			logFile.Sync()
			logFile.Close()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			Eventually(inbox.Messages, "2s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("world"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			blackboxRunner.Stop()
		})

		It("can have a custom hostname", func() {
			config := buildConfigHostname("fake-hostname", logDir)
			blackboxRunner.StartWithConfig(config, 1)

			logFile.WriteString("hello\n")
			logFile.Sync()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring("fake-hostname"))

			blackboxRunner.Stop()
		})

		It("does not log existing messages", func() {
			logFile.WriteString("already present\n")
			logFile.Sync()

			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 1)

			logFile.WriteString("hello\n")
			logFile.Sync()

			var message *sl.Message
			Eventually(inbox.Messages, "2s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))

			blackboxRunner.Stop()
		})

		It("tracks logs in multiple files in source directory", func() {
			anotherLogFile, err := os.OpenFile(
				filepath.Join(logDir, tagName, "another-tail.log"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())
			defer anotherLogFile.Close()

			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 2)

			logFile.WriteString("hello\n")
			logFile.Sync()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			anotherLogFile.WriteString("hello from the other side\n")
			anotherLogFile.Sync()

			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello from the other side"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))
		})

		It("skips files not ending in .log", func() {
			anotherLogFile, err := os.OpenFile(
				filepath.Join(logDir, tagName, "another-tail.log"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())
			defer anotherLogFile.Close()

			notALogFile, err := os.OpenFile(
				filepath.Join(logDir, tagName, "not-a-log-file.log.1"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())
			defer notALogFile.Close()

			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 2)

			logFile.WriteString("hello\n")
			logFile.Sync()

			notALogFile.WriteString("john cena\n")
			notALogFile.Sync()

			var message *sl.Message
			Eventually(inbox.Messages, "30s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			Consistently(inbox.Messages).ShouldNot(Receive())

			anotherLogFile.WriteString("hello from the other side\n")
			anotherLogFile.Sync()

			notALogFile.WriteString("my time is now\n")
			notALogFile.Sync()

			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello from the other side"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			Consistently(inbox.Messages).ShouldNot(Receive())
		})

		It("tracks files in multiple directories using multiple tags", func() {
			tagName2 := "2-test-2-tag"
			err := os.Mkdir(filepath.Join(logDir, tagName2), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			anotherLogFile, err := os.OpenFile(
				filepath.Join(logDir, tagName2, "another-tail.log"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())
			defer anotherLogFile.Close()

			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 2)

			logFile.WriteString("hello\n")
			logFile.Sync()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			anotherLogFile.WriteString("hello from the other side\n")
			anotherLogFile.Sync()

			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello from the other side"))
			Expect(message.Content).To(ContainSubstring("2-test-2-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))
		})

		It("starts tracking logs in newly created files", func() {
			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 1)

			anotherLogFile, err := os.OpenFile(
				filepath.Join(logDir, tagName, "another-tail.log"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())
			defer anotherLogFile.Close()

			// wait for tailer to pick up file, twice the interval
			time.Sleep(10 * time.Second)

			anotherLogFile.WriteString("hello from the other side\n")
			anotherLogFile.Sync()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello from the other side"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			By("keeping track of old files")
			logFile.WriteString("hello\n")
			logFile.Sync()

			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))
		})

		It("continues discovering new files after the original files get deleted", func() {
			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 1)

			logFile.WriteString("hello\n")
			logFile.Sync()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			err := os.Remove(filepath.Join(logDir, tagName, "tail.log"))
			Expect(err).NotTo(HaveOccurred())

			// wait for tail process to die, tailer interval is 1 sec
			time.Sleep(2 * time.Second)

			anotherLogFile, err := os.OpenFile(
				filepath.Join(logDir, tagName, "tail.log"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())
			defer anotherLogFile.Close()

			// wait for tailer to pick up file, twice the interval
			time.Sleep(10 * time.Second)

			anotherLogFile.WriteString("bye\n")
			anotherLogFile.Sync()

			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("bye"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))
		})

		It("ignores subdirectories in tag directories", func() {
			err := os.Mkdir(filepath.Join(logDir, tagName, "ignore-me"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(
				filepath.Join(logDir, tagName, "ignore-me", "and-my-son.log"),
				[]byte("some-data"),
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())

			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 1)

			logFile.WriteString("hello\n")
			logFile.Sync()
			logFile.Close()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			blackboxRunner.Stop()
		})

		It("ignores files in source directory", func() {
			err := ioutil.WriteFile(
				filepath.Join(logDir, "not-a-tag-dir.log"),
				[]byte("some-data"),
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())

			config := buildConfig(logDir)
			blackboxRunner.StartWithConfig(config, 1)

			logFile.WriteString("hello\n")
			logFile.Sync()
			logFile.Close()

			var message *sl.Message
			Eventually(inbox.Messages, "5s").Should(Receive(&message))
			Expect(message.Content).To(ContainSubstring("hello"))
			Expect(message.Content).To(ContainSubstring("test-tag"))
			Expect(message.Content).To(ContainSubstring(Hostname()))

			blackboxRunner.Stop()
		})
	})

	Context("when the syslog server is not already running", func() {
		var serverProcess ifrit.Process

		AfterEach(func() {
			ginkgomon.Interrupt(serverProcess)
		})

		It("tails files when server takes a long time to start", func() {
			address := fmt.Sprintf("127.0.0.1:%d", 9090+GinkgoParallelNode())

			config := blackbox.Config{
				Hostname: "",
				Syslog: blackbox.SyslogConfig{
					Destination: syslog.Drain{
						Transport: "tcp",
						Address:   address,
					},
					SourceDir: logDir,
				},
			}

			configPath := CreateConfigFile(config)

			blackboxCmd := exec.Command(blackboxPath, "-config", configPath)

			session, err := gexec.Start(blackboxCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(2 * syslog.ServerPollingInterval)

			buffer := gbytes.NewBuffer()
			serverProcess = ginkgomon.Invoke(&TcpSyslogServer{
				Addr:   address,
				Buffer: buffer,
			})

			Eventually(session.Err, "10s").Should(gbytes.Say("Seeked"))

			logFile.WriteString("hello\n")
			logFile.WriteString("world\n")
			logFile.Sync()

			Eventually(buffer, "5s").Should(gbytes.Say("hello"))
			Eventually(buffer, "5s").Should(gbytes.Say("world"))

			ginkgomon.Interrupt(serverProcess)

			logFile.WriteString("can't log this\n")
			logFile.Sync()

			time.Sleep(2 * syslog.ServerPollingInterval)

			serverProcess = ginkgomon.Invoke(&TcpSyslogServer{
				Addr:   address,
				Buffer: buffer,
			})

			logFile.WriteString("more\n")
			logFile.Sync()
			logFile.Close()

			Eventually(buffer, "5s").Should(gbytes.Say("can't log this"))
			Eventually(buffer, "5s").Should(gbytes.Say("more"))

			session.Signal(os.Interrupt)
			session.Wait()
		})
	})
})
