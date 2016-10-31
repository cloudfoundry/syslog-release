package blackbox_test

import (
	"time"

	. "github.com/concourse/blackbox"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Duration", func() {
		It("can be unmarshalled from YAML as a string", func() {
			var duration Duration
			err := yaml.Unmarshal([]byte("10s"), &duration)
			Expect(err).ToNot(HaveOccurred())

			Expect(time.Duration(duration)).To(Equal(10 * time.Second))
		})

		It("can be unmarshalled from YAML as an integer", func() {
			var duration Duration
			err := yaml.Unmarshal([]byte("10"), &duration)
			Expect(err).ToNot(HaveOccurred())

			Expect(time.Duration(duration)).To(Equal(10 * time.Nanosecond))
		})
	})
})
