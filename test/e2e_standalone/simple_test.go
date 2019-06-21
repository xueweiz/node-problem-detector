package e2e_standalone_test

import (
	"os"
	"path"
	"testing"

	"k8s.io/kubernetes/test/e2e/framework/ssh"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
)

func TestNPD(t *testing.T) {
	junitFile := "junit.xml"
	artifacts_dir := os.Getenv("ARTIFACTS")
	if artifacts_dir != "" {
		junitFile = path.Join(artifacts_dir, junitFile)
	}

	junitReporter := reporters.NewJUnitReporter(junitFile)
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "NPD Standalone Suite", []ginkgo.Reporter{junitReporter})
}

var _ = ginkgo.Describe("NPD", func() {
	ginkgo.It("dummy test foo", func() {
		ssh.SSH("pwd", "34.83.159.130", "gce")
		return
	})
	ginkgo.It("dummy test bar", func() {
		return
	})
})
