package healthcheck

import (
	"fmt"
	"database/sql"
)

const (
	JOINING_STATE        = "1"
	DONOR_DESYNCED_STATE = "2"
	JOINED_STATE         = "3"
	SYNCED_STATE         = "4"
)

type Healthchecker struct {
	db     *sql.DB
	config HealthcheckerConfig
}

type HealthcheckerConfig struct {
	AvailableWhenDonor    bool
	AvailableWhenReadOnly bool
}

type HealthResult struct {
	Healthy             bool   `json:healthy`
	ClusterConfId       string `json:wsrep_cluster_conf_id`
	ClusterSize         string `json:wsrep_cluster_size`
	ClusterStateUUID    string `json:wsrep_cluster_state_uuid`
	ClusterStatus       string `json:wsrep_cluster_status`
	Connected           string `json:wsrep_connected`
	LocalState          string `json:wsrep_local_state`
	LocalStateComment   string `json:wsrep_local_state_comment`
	Messages          []string `json:messages`
	ReadOnly            string `json:read_only`
	Ready               string `json:wsrep_ready`
}

func New(db *sql.DB, config HealthcheckerConfig) *Healthchecker {
	return &Healthchecker{
		db:     db,
		config: config,
	}
}

func (h *Healthchecker) Check() *HealthResult {
	var variable_name string
	var result = &HealthResult{Messages:make([]string, 0)}

	var statusValues = map[string]*string{
		"wsrep_local_state": &result.LocalState,
		"wsrep_local_state_comment": &result.LocalStateComment,
		"wsrep_cluster_conf_id": &result.ClusterConfId,
		"wsrep_cluster_size": &result.ClusterSize,
		"wsrep_cluster_state_uuid": &result.ClusterStateUUID,
		"wsrep_cluster_status": &result.ClusterStatus,
		"wsrep_connected": &result.Connected,
		"wsrep_ready": &result.Ready,
	}

	for key := range statusValues {
		err := h.db.QueryRow(fmt.Sprintf("SHOW STATUS LIKE '%s'", key)).Scan(&variable_name, statusValues[key])
		if err != nil {
			if (err.Error() == "sql: no rows in result set") {
				*statusValues[key] = "--"
			} else {
				result.Messages = append(result.Messages, fmt.Sprintf("Could not get %s value: %s", key, err.Error()))
			}
		}
	}

	err := h.db.QueryRow("SHOW GLOBAL VARIABLES LIKE 'read_only'").Scan(&variable_name, &result.ReadOnly)
	if err != nil {
		if (err.Error() == "sql: no rows in result set") {
			result.ReadOnly = "--"
		} else {
			result.Messages = append(result.Messages, fmt.Sprintf("Could not get read_only value: %s", err.Error()))
		}
	}

	if len(result.Messages) == 0 &&
	   (result.LocalState == SYNCED_STATE || (result.LocalState == DONOR_DESYNCED_STATE && h.config.AvailableWhenDonor)) {
		if (result.ClusterStatus == "Primary") {
			result.Healthy = true
		} else {
			result.Messages = append(result.Messages, "Node is not part of the primary component!")
		}

		if !h.config.AvailableWhenReadOnly && result.ReadOnly == "ON" {
			result.Healthy = false
			result.Messages = append(result.Messages, "Node is read-only")
		}
	}

	return result
}
