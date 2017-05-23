package timeouts_test

import (
	"fmt"
	"os/exec"
	"strconv"

	"code.cloudfoundry.org/go-db-helpers/db"
	"code.cloudfoundry.org/go-db-helpers/testsupport"
	"code.cloudfoundry.org/silk/controller"
	"code.cloudfoundry.org/silk/controller/config"
	"code.cloudfoundry.org/silk/controller/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var testTimeoutInSeconds = 5

var (
	session    *gexec.Session
	conf       config.Config
	dbConf     db.Config
	testClient *controller.Client
)

var _ = BeforeEach(func() {
	dbConf = testsupport.GetDBConfig()
	dbConf.DatabaseName = fmt.Sprintf("test_database_%x", GinkgoParallelNode())
	dbConf.Timeout = testTimeoutInSeconds - 1
	testsupport.CreateDatabase(dbConf)

	conf = helpers.DefaultTestConfig(dbConf, "../fixtures")
	testClient = helpers.TestClient(conf, "../fixtures")
	session = helpers.StartAndWaitForServer(controllerBinaryPath, conf, testClient)
})

var _ = AfterEach(func() {
	helpers.StopServer(session)
	err := testsupport.RemoveDatabase(dbConf)
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Timeouts", func() {
	Context("when the database is unreachable", func() {
		BeforeEach(func() {
			By("blocking access to port " + strconv.Itoa(int(dbConf.Port)))
			mustSucceed("iptables", "-A", "INPUT", "-p", "tcp", "--dport", strconv.Itoa(int(dbConf.Port)), "-j", "DROP")
		})
		AfterEach(func() {
			By("allowing access to port " + strconv.Itoa(int(dbConf.Port)))
			mustSucceed("iptables", "-D", "INPUT", "-p", "tcp", "--dport", strconv.Itoa(int(dbConf.Port)), "-j", "DROP")
		})

		It("AcquireSubnetLease times out with an error", func(done Done) {
			_, err := testClient.AcquireSubnetLease("10.244.4.5")
			Expect(err).To(MatchError(ContainSubstring("http status 500")))

			close(done)
		}, float64(testTimeoutInSeconds))

		It("GetRoutableLeases times out with an error", func(done Done) {
			_, err := testClient.GetRoutableLeases()
			Expect(err).To(MatchError(ContainSubstring("http status 500")))

			close(done)
		}, float64(testTimeoutInSeconds))

		It("ReleaseSubnetLease times out with an error", func(done Done) {
			err := testClient.ReleaseSubnetLease("10.244.4.5")
			Expect(err).To(MatchError(ContainSubstring("http status 500")))

			close(done)
		}, float64(testTimeoutInSeconds))

		It("RenewSubnetLease times out with an error", func(done Done) {
			err := testClient.RenewSubnetLease(controller.Lease{
				UnderlayIP:          "10.244.4.5",
				OverlaySubnet:       "10.255.10.0/24",
				OverlayHardwareAddr: "ee:ee:0a:ff:0a:00",
			})
			Expect(err).To(MatchError(ContainSubstring("http status 500")))

			close(done)
		}, float64(testTimeoutInSeconds))
	})
})

func mustSucceed(binary string, args ...string) string {
	cmd := exec.Command(binary, args...)
	sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, helpers.DEFAULT_TIMEOUT).Should(gexec.Exit(0))
	return string(sess.Out.Contents())
}
