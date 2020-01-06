package migrations

import (
	"database/sql"
	"encoding/json"

	"github.com/concourse/concourse/atc"
)

func (self *migrations) Up_1574452410() error {
	tx, err := self.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	rows, err := tx.Query("SELECT pipeline_id, config, nonce FROM jobs WHERE active = true")
	if err != nil {
		return err
	}

	pipelineJobConfigs := make(map[int]atc.JobConfigs)
	for rows.Next() {
		var configBlob []byte
		var nonce sql.NullString
		var pipelineID int

		err = rows.Scan(&pipelineID, &configBlob, &nonce)
		if err != nil {
			return err
		}

		var noncense *string
		if nonce.Valid {
			noncense = &nonce.String
		}

		decrypted, err := self.Strategy.Decrypt(string(configBlob), noncense)
		if err != nil {
			return err
		}

		var config atc.JobConfig
		err = json.Unmarshal(decrypted, &config)
		if err != nil {
			return err
		}

		pipelineJobConfigs[pipelineID] = append(pipelineJobConfigs[pipelineID], config)
	}

	for pipelineID, jobConfigs := range pipelineJobConfigs {
		resourceNameToID := make(map[string]int)
		jobNameToID := make(map[string]int)

		rows, err := tx.Query("SELECT id, name FROM resources WHERE pipeline_id = $1", pipelineID)
		if err != nil {
			return err
		}

		for rows.Next() {
			var id int
			var name string

			err = rows.Scan(&id, &name)
			if err != nil {
				return err
			}

			resourceNameToID[name] = id
		}

		rows, err = tx.Query("SELECT id, name FROM jobs WHERE pipeline_id = $1", pipelineID)
		if err != nil {
			return err
		}

		for rows.Next() {
			var id int
			var name string

			err = rows.Scan(&id, &name)
			if err != nil {
				return err
			}

			jobNameToID[name] = id
		}

		for _, jobConfig := range jobConfigs {
			for _, plan := range jobConfig.Plan {
				if plan.Get != "" {
					err = insertJobInput(tx, plan, jobConfig.Name, resourceNameToID, jobNameToID)
					if err != nil {
						return err
					}
				} else if plan.Put != "" {
					err = insertJobOutput(tx, plan, jobConfig.Name, resourceNameToID, jobNameToID)
					if err != nil {
						return err
					}
				}
			}

			err = updateJob(tx, jobConfig, jobNameToID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func insertJobInput(tx *sql.Tx, plan atc.PlanConfig, jobName string, resourceNameToID map[string]int, jobNameToID map[string]int) error {
	if len(plan.Passed) != 0 {
		for _, passedJob := range plan.Passed {
			var resourceID int
			if plan.Resource != "" {
				resourceID = resourceNameToID[plan.Resource]
			} else {
				resourceID = resourceNameToID[plan.Get]
			}

			var version sql.NullString
			if plan.Version != nil {
				versionJSON, err := plan.Version.MarshalJSON()
				if err != nil {
					return err
				}

				version = sql.NullString{Valid: true, String: string(versionJSON)}
			}

			_, err := tx.Exec("INSERT INTO job_inputs (name, job_id, resource_id, passed_job_id, trigger, version) VALUES ($1, $2, $3, $4, $5, $6)", plan.Get, jobNameToID[jobName], resourceID, jobNameToID[passedJob], plan.Trigger, version)
			if err != nil {
				return err
			}
		}
	} else {
		var resourceID int
		if plan.Resource != "" {
			resourceID = resourceNameToID[plan.Resource]
		} else {
			resourceID = resourceNameToID[plan.Get]
		}

		var version sql.NullString
		if plan.Version != nil {
			versionJSON, err := plan.Version.MarshalJSON()
			if err != nil {
				return err
			}

			version = sql.NullString{Valid: true, String: string(versionJSON)}
		}

		_, err := tx.Exec("INSERT INTO job_inputs (name, job_id, resource_id, trigger, version) VALUES ($1, $2, $3, $4, $5)", plan.Get, jobNameToID[jobName], resourceID, plan.Trigger, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertJobOutput(tx *sql.Tx, plan atc.PlanConfig, jobName string, resourceNameToID map[string]int, jobNameToID map[string]int) error {
	var resourceID int
	if plan.Resource != "" {
		resourceID = resourceNameToID[plan.Resource]
	} else {
		resourceID = resourceNameToID[plan.Put]
	}

	_, err := tx.Exec("INSERT INTO job_outputs (name, job_id, resource_id) VALUES ($1, $2, $3)", plan.Put, jobNameToID[jobName], resourceID)
	if err != nil {
		return err
	}

	return nil
}

func updateJob(tx *sql.Tx, jobConfig atc.JobConfig, jobNameToID map[string]int) error {
	_, err := tx.Exec("UPDATE jobs SET public = $1, max_in_flight = $2, disable_manual_trigger = $3 WHERE id = $4", jobConfig.Public, jobConfig.MaxInFlight(), jobConfig.DisableManualTrigger, jobNameToID[jobConfig.Name])
	if err != nil {
		return err
	}

	return nil
}