package healthcheck_test

import (
	"errors"
	"fmt"

	"database/sql"
	testdb "github.com/erikstmartin/go-testdb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/BrandYourself/galera-healthcheck/healthcheck"
)

var _ = Describe("GaleraHealthChecker", func() {

	Describe("Check", func() {
		Context("when HealthCheck is called", func() {
			It("reads the correct variables from the server", func() {
				config := HealthcheckTestHelperConfig{
					wsrep_local_state: "4",
					wsrep_local_state_comment: "Synced",
					wsrep_cluster_conf_id: "15",
					wsrep_cluster_size: "3",
					wsrep_cluster_state_uuid: "1234-5678-90ab",
					wsrep_cluster_status: "Primary",
					wsrep_connected: "ON",
					wsrep_ready: "ON",
					read_only: "OFF",
					available_when_donor: false,
					available_when_read_only: false,
				}

				result := HealthcheckTestHelper(config)

				Expect(result.Healthy).To(BeTrue())
				Expect(len(result.Messages)).To(Equal(0))
				Expect(result.ClusterConfId).To(Equal(config.wsrep_cluster_conf_id))
				Expect(result.ClusterSize).To(Equal(config.wsrep_cluster_size))
				Expect(result.ClusterStateUUID).To(Equal(config.wsrep_cluster_state_uuid))
				Expect(result.ClusterStatus).To(Equal(config.wsrep_cluster_status))
				Expect(result.Connected).To(Equal(config.wsrep_connected))
				Expect(result.LocalState).To(Equal(config.wsrep_local_state))
				Expect(result.LocalStateComment).To(Equal(config.wsrep_local_state_comment))
				Expect(result.ReadOnly).To(Equal(config.read_only))
				Expect(result.Ready).To(Equal(config.wsrep_ready))
			})
		})

		Context("when WSREP_STATUS is joining", func() {
			It("returns false and Joining", func() {
				config := HealthcheckTestHelperConfig{
					wsrep_local_state: "1",
					read_only: "OFF",
					available_when_donor: false,
					available_when_read_only: false,
				}

				result := HealthcheckTestHelper(config)

				Expect(result.Healthy).To(BeFalse())
				Expect(len(result.Messages)).To(Equal(0))
			})
		})

		Context("when WSREP_STATUS is joined", func() {
			It("returns false and Joined", func() {
				config := HealthcheckTestHelperConfig{
					wsrep_local_state: "3",
					read_only: "OFF",
					available_when_donor: false,
					available_when_read_only: false,
				}

				result := HealthcheckTestHelper(config)

				Expect(result.Healthy).To(BeFalse())
				Expect(len(result.Messages)).To(Equal(0))
			})
		})

		Context("when WSREP_STATUS is donor", func() {
			Context("when not AVAILABLE_WHEN_DONOR", func() {
				It("returns false and Donor", func() {
					config := HealthcheckTestHelperConfig{
						wsrep_local_state: "2",
						read_only: "OFF",
						available_when_donor: false,
						available_when_read_only: false,
					}

					result := HealthcheckTestHelper(config)

					Expect(result.Healthy).To(BeFalse())
					Expect(len(result.Messages)).To(Equal(0))
				})
			})

			Context("when AVAILABLE_WHEN_DONOR", func() {
				Context("when READ_ONLY is ON and AVAILABLE_WHEN_READONLY is true", func() {
					It("returns true and synced", func() {
						config := HealthcheckTestHelperConfig{
							wsrep_local_state: "2",
							read_only: "ON",
							wsrep_cluster_status: "Primary",
							available_when_donor: true,
							available_when_read_only: true,
						}

						result := HealthcheckTestHelper(config)

						Expect(result.Healthy).To(BeTrue())
						Expect(len(result.Messages)).To(Equal(0))
					})
				})

				Context("when READ_ONLY is ON and AVAILABLE_WHEN_READONLY is false", func() {
					It("returns false and read-only", func() {
						config := HealthcheckTestHelperConfig{
							wsrep_local_state: "2",
							read_only: "ON",
							wsrep_cluster_status: "Primary",
							available_when_donor: true,
							available_when_read_only: false,
						}

						result := HealthcheckTestHelper(config)

						Expect(result.Healthy).To(BeFalse())
						Expect(len(result.Messages)).To(Equal(1))
						Expect(result.Messages[0]).To(Equal("Node is read-only"))
					})
				})

				Context("when READ_ONLY is OFF", func() {
					It("returns true and synced", func() {
						config := HealthcheckTestHelperConfig{
							wsrep_local_state: "2",
							read_only: "OFF",
							wsrep_cluster_status: "Primary",
							available_when_donor: true,
							available_when_read_only: false,
						}

						result := HealthcheckTestHelper(config)

						Expect(result.Healthy).To(BeTrue())
						Expect(len(result.Messages)).To(Equal(0))
					})
				})
			})

		})

		Context("when WSREP_STATUS is synced", func() {

			Context("when READ_ONLY is ON and AVAILABLE_WHEN_READONLY is true", func() {
				It("returns true and synced", func() {

					config := HealthcheckTestHelperConfig{
						wsrep_local_state: "4",
						read_only: "ON",
						wsrep_cluster_status: "Primary",
						available_when_donor: false,
						available_when_read_only: true,
					}

					result := HealthcheckTestHelper(config)

					Expect(result.Healthy).To(BeTrue())
					Expect(len(result.Messages)).To(Equal(0))
				})
			})

			Context("when READ_ONLY is ON and AVAILABLE_WHEN_READONLY is false", func() {
				It("returns false and read-only", func() {

					config := HealthcheckTestHelperConfig{
						wsrep_local_state: "4",
						read_only: "ON",
						wsrep_cluster_status: "Primary",
						available_when_donor: false,
						available_when_read_only: false,
					}

					result := HealthcheckTestHelper(config)

					Expect(result.Healthy).To(BeFalse())
					Expect(len(result.Messages)).To(Equal(1))
					Expect(result.Messages[0]).To(Equal("Node is read-only"))
				})
			})

			Context("when READ_ONLY is OFF", func() {
				It("returns true and synced", func() {
					config := HealthcheckTestHelperConfig{
						wsrep_local_state: "4",
						read_only: "OFF",
						wsrep_cluster_status: "Primary",
						available_when_donor: false,
						available_when_read_only: false,
					}

					result := HealthcheckTestHelper(config)

					Expect(result.Healthy).To(BeTrue())
					Expect(len(result.Messages)).To(Equal(0))
				})
			})

			Context("when the node is not in the primary component", func() {
				It("returns false and the correct error", func() {
					config := HealthcheckTestHelperConfig{
						wsrep_local_state: "4",
						read_only: "OFF",
						wsrep_cluster_status: "Non-Primary",
						available_when_donor: false,
						available_when_read_only: false,
					}

					result := HealthcheckTestHelper(config)

					Expect(result.Healthy).To(BeFalse())
					Expect(len(result.Messages)).To(Equal(1))
					Expect(result.Messages[0]).To(Equal("Node is not part of the primary component!"))
				})
			})
		})

		Context("when SHOW STATUS has errors", func() {
			It("returns false and the error message", func() {

				db, _ := sql.Open("testdb", "")

				sql := "SHOW STATUS LIKE 'wsrep_local_state'"
				testdb.StubQueryError(sql, errors.New("test error"))

				config := healthcheck.HealthcheckerConfig{
					AvailableWhenDonor:    false,
					AvailableWhenReadOnly: false,
				}

				healthchecker := healthcheck.New(db, config)

				result := healthchecker.Check()

				Expect(result.Healthy).To(BeFalse())
				Expect(len(result.Messages)).To(Equal(1))
				Expect(result.Messages[0]).To(Equal("Could not get wsrep_local_state value: test error"))
			})
		})

		Context("when SHOW STATUS has errors", func() {
			It("returns false and the error message", func() {

				db, _ := sql.Open("testdb", "")

				sql := "SHOW STATUS LIKE 'wsrep_local_state'"
				columns := []string{"Variable_name", "Value"}
				value := "wsrep_local_state,4"
				testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, value))

				sql = "SHOW GLOBAL VARIABLES LIKE 'read_only'"
				testdb.StubQueryError(sql, errors.New("another test error"))

				config := healthcheck.HealthcheckerConfig{
					AvailableWhenDonor:    false,
					AvailableWhenReadOnly: false,
				}

				healthchecker := healthcheck.New(db, config)

				result := healthchecker.Check()

				Expect(result.Healthy).To(BeFalse())
				Expect(len(result.Messages)).To(Equal(1))
				Expect(result.Messages[0]).To(Equal("Could not get read_only value: another test error"))
			})
		})
	})
})

type HealthcheckTestHelperConfig struct {
	wsrep_local_state string
	wsrep_local_state_comment string
	wsrep_cluster_conf_id string
	wsrep_cluster_size string
	wsrep_cluster_state_uuid string
	wsrep_cluster_status string
	wsrep_connected string
	wsrep_ready string
	read_only string
	available_when_donor     bool
	available_when_read_only bool
}

func HealthcheckTestHelper(testConfig HealthcheckTestHelperConfig) *healthcheck.HealthResult {
	db, _ := sql.Open("testdb", "")

	sql := "SHOW STATUS LIKE 'wsrep_local_state'"
	columns := []string{"Variable_name", "Value"}
	result := fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_local_state)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW STATUS LIKE 'wsrep_local_state_comment'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_local_state_comment)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW STATUS LIKE 'wsrep_cluster_conf_id'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_cluster_conf_id)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW STATUS LIKE 'wsrep_cluster_size'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_cluster_size)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW STATUS LIKE 'wsrep_cluster_state_uuid'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_cluster_state_uuid)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW STATUS LIKE 'wsrep_cluster_status'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_cluster_status)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW STATUS LIKE 'wsrep_connected'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_connected)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW STATUS LIKE 'wsrep_ready'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("wsrep_local_state,%s", testConfig.wsrep_ready)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql = "SHOW GLOBAL VARIABLES LIKE 'read_only'"
	columns = []string{"Variable_name", "Value"}
	result = fmt.Sprintf("read_only,%s", testConfig.read_only)
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	config := healthcheck.HealthcheckerConfig{
		AvailableWhenDonor:    testConfig.available_when_donor,
		AvailableWhenReadOnly: testConfig.available_when_read_only,
	}

	healthchecker := healthcheck.New(db, config)

	return healthchecker.Check()
}
