package dtos

type (

	// VmraySampleData represents the data field of the vmray sample response
	VmraySampleData struct {
		SampleChildSampleIds         []interface{} `json:"sample_child_sample_ids"`
		SampleClassifications        []interface{} `json:"sample_classifications"`
		SampleContainerType          interface{}   `json:"sample_container_type"`
		SampleCreated                string        `json:"sample_created"`
		SampleFilename               string        `json:"sample_filename"`
		SampleFilesize               int           `json:"sample_filesize"`
		SampleHighestVtiScore        int           `json:"sample_highest_vti_score"`
		SampleHighestVtiSeverity     string        `json:"sample_highest_vti_severity"`
		SampleID                     int           `json:"sample_id"`
		SampleImphash                interface{}   `json:"sample_imphash"`
		SampleIsMultipart            bool          `json:"sample_is_multipart"`
		SampleLastMdScore            interface{}   `json:"sample_last_md_score"`
		SampleLastReputationSeverity string        `json:"sample_last_reputation_severity"`
		SampleLastVtScore            interface{}   `json:"sample_last_vt_score"`
		SampleMd5Hash                string        `json:"sample_md5hash"`
		SampleParentSampleIds        []interface{} `json:"sample_parent_sample_ids"`
		SamplePriority               int           `json:"sample_priority"`
		SampleScore                  int           `json:"sample_score"`
		SampleSeverity               string        `json:"sample_severity"`
		SampleSha1Hash               string        `json:"sample_sha1hash"`
		SampleSha256Hash             string        `json:"sample_sha256hash"`
		SampleSsdeephash             string        `json:"sample_ssdeephash"`
		SampleType                   string        `json:"sample_type"`
		SampleURL                    interface{}   `json:"sample_url"`
		SampleVtiScore               int           `json:"sample_vti_score"`
		SampleWebifURL               string        `json:"sample_webif_url"`
	}
	// GetVmraySampleResponse represents the get sample response payload
	GetVmraySampleResponse struct {
		Data   VmraySampleData `json:"data"`
		Result string          `json:"result"`
	}

	// VmraySubmissions represents the submissions field of submit response of vmray
	VmraySubmissions struct {
		SubmissionAnalyzerModeAnalyzerMode                string        `json:"submission_analyzer_mode_analyzer_mode"`
		SubmissionAnalyzerModeArchiveAction               string        `json:"submission_analyzer_mode_archive_action"`
		SubmissionAnalyzerModeDetonateLinksInEmails       bool          `json:"submission_analyzer_mode_detonate_links_in_emails"`
		SubmissionAnalyzerModeEnableReputation            bool          `json:"submission_analyzer_mode_enable_reputation"`
		SubmissionAnalyzerModeEnableWhois                 bool          `json:"submission_analyzer_mode_enable_whois"`
		SubmissionAnalyzerModeID                          int           `json:"submission_analyzer_mode_id"`
		SubmissionAnalyzerModeKnownBenign                 bool          `json:"submission_analyzer_mode_known_benign"`
		SubmissionAnalyzerModeKnownMalicious              bool          `json:"submission_analyzer_mode_known_malicious"`
		SubmissionAnalyzerModeMaxDynamicAnalysesPerSample string        `json:"submission_analyzer_mode_max_dynamic_analyses_per_sample"`
		SubmissionAnalyzerModeMaxRecursiveSamples         string        `json:"submission_analyzer_mode_max_recursive_samples"`
		SubmissionAnalyzerModeReanalyze                   bool          `json:"submission_analyzer_mode_reanalyze"`
		SubmissionAnalyzerModeTriage                      string        `json:"submission_analyzer_mode_triage"`
		SubmissionAnalyzerModeTriageErrorHandling         interface{}   `json:"submission_analyzer_mode_triage_error_handling"`
		SubmissionAPIKeyID                                int           `json:"submission_api_key_id"`
		SubmissionBillingType                             string        `json:"submission_billing_type"`
		SubmissionComment                                 interface{}   `json:"submission_comment"`
		SubmissionCreated                                 string        `json:"submission_created"`
		SubmissionDeletionDate                            interface{}   `json:"submission_deletion_date"`
		SubmissionDllCallMode                             interface{}   `json:"submission_dll_call_mode"`
		SubmissionDllCalls                                interface{}   `json:"submission_dll_calls"`
		SubmissionDocumentPassword                        interface{}   `json:"submission_document_password"`
		SubmissionEnableLocalAv                           bool          `json:"submission_enable_local_av"`
		SubmissionFilename                                string        `json:"submission_filename"`
		SubmissionFinishTime                              interface{}   `json:"submission_finish_time"`
		SubmissionFinished                                bool          `json:"submission_finished"`
		SubmissionHasErrors                               interface{}   `json:"submission_has_errors"`
		SubmissionID                                      int           `json:"submission_id"`
		SubmissionIPID                                    int           `json:"submission_ip_id"`
		SubmissionIPIP                                    string        `json:"submission_ip_ip"`
		SubmissionKnownConfiguration                      bool          `json:"submission_known_configuration"`
		SubmissionOriginalFilename                        string        `json:"submission_original_filename"`
		SubmissionOriginalURL                             interface{}   `json:"submission_original_url"`
		SubmissionPrescriptForceAdmin                     bool          `json:"submission_prescript_force_admin"`
		SubmissionPrescriptID                             interface{}   `json:"submission_prescript_id"`
		SubmissionPriority                                int           `json:"submission_priority"`
		SubmissionReputationMode                          string        `json:"submission_reputation_mode"`
		SubmissionRetentionPeriod                         int           `json:"submission_retention_period"`
		SubmissionSampleID                                int           `json:"submission_sample_id"`
		SubmissionSampleMd5                               string        `json:"submission_sample_md5"`
		SubmissionSampleSha1                              string        `json:"submission_sample_sha1"`
		SubmissionSampleSha256                            string        `json:"submission_sample_sha256"`
		SubmissionSampleSsdeep                            string        `json:"submission_sample_ssdeep"`
		SubmissionScore                                   interface{}   `json:"submission_score"`
		SubmissionSeverity                                interface{}   `json:"submission_severity"`
		SubmissionShareable                               bool          `json:"submission_shareable"`
		SubmissionSubmissionMetadata                      string        `json:"submission_submission_metadata"`
		SubmissionSystemTime                              interface{}   `json:"submission_system_time"`
		SubmissionTags                                    []interface{} `json:"submission_tags"`
		SubmissionTriageErrorHandling                     interface{}   `json:"submission_triage_error_handling"`
		SubmissionType                                    string        `json:"submission_type"`
		SubmissionUserAccountID                           int           `json:"submission_user_account_id"`
		SubmissionUserAccountName                         string        `json:"submission_user_account_name"`
		SubmissionUserAccountSubscriptionMode             string        `json:"submission_user_account_subscription_mode"`
		SubmissionUserEmail                               string        `json:"submission_user_email"`
		SubmissionUserID                                  int           `json:"submission_user_id"`
		SubmissionWebifURL                                string        `json:"submission_webif_url"`
		SubmissionWhoisMode                               string        `json:"submission_whois_mode"`
	}

	// VmraySubmitData represents the data field of the vmray submit response
	VmraySubmitData struct {
		Errors []struct {
			ErrMsg string `json:"ErrMsg"`
		} `json:"errors"`
		Jobs []struct {
			JobAnalyzerID          int         `json:"job_analyzer_id"`
			JobAnalyzerName        string      `json:"job_analyzer_name"`
			JobBillID              int         `json:"job_bill_id"`
			JobBillType            string      `json:"job_bill_type"`
			JobConfigurationID     int         `json:"job_configuration_id"`
			JobConfigurationName   string      `json:"job_configuration_name"`
			JobCreated             string      `json:"job_created"`
			JobDocumentPassword    interface{} `json:"job_document_password"`
			JobEnableLocalAv       bool        `json:"job_enable_local_av"`
			JobID                  int         `json:"job_id"`
			JobJobruleID           int         `json:"job_jobrule_id"`
			JobJobruleSampletype   string      `json:"job_jobrule_sampletype"`
			JobParentAnalysisID    interface{} `json:"job_parent_analysis_id"`
			JobPrescriptForceAdmin bool        `json:"job_prescript_force_admin"`
			JobPrescriptID         interface{} `json:"job_prescript_id"`
			JobPriority            int         `json:"job_priority"`
			JobReputationJobID     int         `json:"job_reputation_job_id"`
			JobSampleID            int         `json:"job_sample_id"`
			JobSampleMd5           string      `json:"job_sample_md5"`
			JobSampleSha1          string      `json:"job_sample_sha1"`
			JobSampleSha256        string      `json:"job_sample_sha256"`
			JobSampleSsdeep        string      `json:"job_sample_ssdeep"`
			JobSnapshotID          int         `json:"job_snapshot_id"`
			JobSnapshotName        string      `json:"job_snapshot_name"`
			JobStaticConfigID      interface{} `json:"job_static_config_id"`
			JobStatus              string      `json:"job_status"`
			JobStatuschanged       string      `json:"job_statuschanged"`
			JobSubmissionID        int         `json:"job_submission_id"`
			JobSystemTime          interface{} `json:"job_system_time"`
			JobTrackingState       string      `json:"job_tracking_state"`
			JobType                string      `json:"job_type"`
			JobUserEmail           string      `json:"job_user_email"`
			JobUserID              int         `json:"job_user_id"`
			JobVMID                int         `json:"job_vm_id"`
			JobVMName              string      `json:"job_vm_name"`
			JobVmhostID            interface{} `json:"job_vmhost_id"`
			JobVminstanceNum       interface{} `json:"job_vminstance_num"`
			JobVncToken            string      `json:"job_vnc_token"`
		} `json:"jobs"`
		MdJobs         []interface{} `json:"md_jobs"`
		ReputationJobs []struct {
			ReputationJobBillID        interface{} `json:"reputation_job_bill_id"`
			ReputationJobCreated       string      `json:"reputation_job_created"`
			ReputationJobID            int         `json:"reputation_job_id"`
			ReputationJobPriority      int         `json:"reputation_job_priority"`
			ReputationJobSampleID      int         `json:"reputation_job_sample_id"`
			ReputationJobSampleMd5     string      `json:"reputation_job_sample_md5"`
			ReputationJobSampleSha1    string      `json:"reputation_job_sample_sha1"`
			ReputationJobSampleSha256  string      `json:"reputation_job_sample_sha256"`
			ReputationJobSampleSsdeep  string      `json:"reputation_job_sample_ssdeep"`
			ReputationJobStatus        string      `json:"reputation_job_status"`
			ReputationJobStatuschanged string      `json:"reputation_job_statuschanged"`
			ReputationJobSubmissionID  int         `json:"reputation_job_submission_id"`
			ReputationJobUserEmail     string      `json:"reputation_job_user_email"`
			ReputationJobUserID        int         `json:"reputation_job_user_id"`
		} `json:"reputation_jobs"`
		Samples []struct {
			SampleChildSampleIds  []interface{} `json:"sample_child_sample_ids"`
			SampleContainerType   interface{}   `json:"sample_container_type"`
			SampleCreated         string        `json:"sample_created"`
			SampleFilename        string        `json:"sample_filename"`
			SampleFilesize        int           `json:"sample_filesize"`
			SampleID              int           `json:"sample_id"`
			SampleImphash         interface{}   `json:"sample_imphash"`
			SampleIsMultipart     bool          `json:"sample_is_multipart"`
			SampleMd5Hash         string        `json:"sample_md5hash"`
			SampleParentSampleIds []interface{} `json:"sample_parent_sample_ids"`
			SamplePriority        int           `json:"sample_priority"`
			SampleSha1Hash        string        `json:"sample_sha1hash"`
			SampleSha256Hash      string        `json:"sample_sha256hash"`
			SampleSsdeephash      string        `json:"sample_ssdeephash"`
			SampleType            string        `json:"sample_type"`
			SampleURL             interface{}   `json:"sample_url"`
			SampleWebifURL        string        `json:"sample_webif_url"`
			SubmissionFilename    string        `json:"submission_filename"`
		} `json:"samples"`
		StaticJobs []struct {
			JobAnalyzerID          int         `json:"job_analyzer_id"`
			JobAnalyzerName        string      `json:"job_analyzer_name"`
			JobBillID              interface{} `json:"job_bill_id"`
			JobBillType            string      `json:"job_bill_type"`
			JobConfigurationID     interface{} `json:"job_configuration_id"`
			JobCreated             string      `json:"job_created"`
			JobDocumentPassword    interface{} `json:"job_document_password"`
			JobEnableLocalAv       bool        `json:"job_enable_local_av"`
			JobID                  int         `json:"job_id"`
			JobJobruleID           interface{} `json:"job_jobrule_id"`
			JobParentAnalysisID    interface{} `json:"job_parent_analysis_id"`
			JobPrescriptForceAdmin bool        `json:"job_prescript_force_admin"`
			JobPrescriptID         interface{} `json:"job_prescript_id"`
			JobPriority            int         `json:"job_priority"`
			JobReputationJobID     int         `json:"job_reputation_job_id"`
			JobSampleID            int         `json:"job_sample_id"`
			JobSampleMd5           string      `json:"job_sample_md5"`
			JobSampleSha1          string      `json:"job_sample_sha1"`
			JobSampleSha256        string      `json:"job_sample_sha256"`
			JobSampleSsdeep        string      `json:"job_sample_ssdeep"`
			JobSnapshotID          interface{} `json:"job_snapshot_id"`
			JobStaticConfigID      int         `json:"job_static_config_id"`
			JobStaticConfigName    string      `json:"job_static_config_name"`
			JobStatus              string      `json:"job_status"`
			JobStatuschanged       string      `json:"job_statuschanged"`
			JobSubmissionID        int         `json:"job_submission_id"`
			JobSystemTime          interface{} `json:"job_system_time"`
			JobTrackingState       string      `json:"job_tracking_state"`
			JobType                string      `json:"job_type"`
			JobUserEmail           string      `json:"job_user_email"`
			JobUserID              int         `json:"job_user_id"`
			JobVMID                interface{} `json:"job_vm_id"`
			JobVmhostID            interface{} `json:"job_vmhost_id"`
			JobVminstanceNum       interface{} `json:"job_vminstance_num"`
			JobVncToken            string      `json:"job_vnc_token"`
		} `json:"static_jobs"`
		Submissions []VmraySubmissions `json:"submissions"`
		VtJobs      []interface{}      `json:"vt_jobs"`
		WhoisJobs   []interface{}      `json:"whois_jobs"`
	}

	// VmraySubmitResponse represents the submit sample file enpoint response payload
	VmraySubmitResponse struct {
		Data   VmraySubmitData `json:"data"`
		Result string          `json:"result"`
	}

	// VmraySubmissionData represents the data field of the vmray submission status response
	VmraySubmissionData struct {
		SubmissionAnalyzerModeAnalyzerMode                string        `json:"submission_analyzer_mode_analyzer_mode"`
		SubmissionAnalyzerModeArchiveAction               string        `json:"submission_analyzer_mode_archive_action"`
		SubmissionAnalyzerModeDetonateLinksInEmails       bool          `json:"submission_analyzer_mode_detonate_links_in_emails"`
		SubmissionAnalyzerModeEnableReputation            bool          `json:"submission_analyzer_mode_enable_reputation"`
		SubmissionAnalyzerModeEnableWhois                 bool          `json:"submission_analyzer_mode_enable_whois"`
		SubmissionAnalyzerModeID                          int           `json:"submission_analyzer_mode_id"`
		SubmissionAnalyzerModeKnownBenign                 bool          `json:"submission_analyzer_mode_known_benign"`
		SubmissionAnalyzerModeKnownMalicious              bool          `json:"submission_analyzer_mode_known_malicious"`
		SubmissionAnalyzerModeMaxDynamicAnalysesPerSample string        `json:"submission_analyzer_mode_max_dynamic_analyses_per_sample"`
		SubmissionAnalyzerModeMaxRecursiveSamples         string        `json:"submission_analyzer_mode_max_recursive_samples"`
		SubmissionAnalyzerModeReanalyze                   bool          `json:"submission_analyzer_mode_reanalyze"`
		SubmissionAnalyzerModeTriage                      string        `json:"submission_analyzer_mode_triage"`
		SubmissionAnalyzerModeTriageErrorHandling         interface{}   `json:"submission_analyzer_mode_triage_error_handling"`
		SubmissionAPIKeyID                                int           `json:"submission_api_key_id"`
		SubmissionBillingType                             string        `json:"submission_billing_type"`
		SubmissionComment                                 interface{}   `json:"submission_comment"`
		SubmissionCreated                                 string        `json:"submission_created"`
		SubmissionDeletionDate                            interface{}   `json:"submission_deletion_date"`
		SubmissionDllCallMode                             interface{}   `json:"submission_dll_call_mode"`
		SubmissionDllCalls                                interface{}   `json:"submission_dll_calls"`
		SubmissionDocumentPassword                        interface{}   `json:"submission_document_password"`
		SubmissionEnableLocalAv                           bool          `json:"submission_enable_local_av"`
		SubmissionFilename                                string        `json:"submission_filename"`
		SubmissionFinishTime                              string        `json:"submission_finish_time"`
		SubmissionFinished                                bool          `json:"submission_finished"`
		SubmissionHasErrors                               bool          `json:"submission_has_errors"`
		SubmissionID                                      int           `json:"submission_id"`
		SubmissionIPID                                    int           `json:"submission_ip_id"`
		SubmissionIPIP                                    string        `json:"submission_ip_ip"`
		SubmissionKnownConfiguration                      bool          `json:"submission_known_configuration"`
		SubmissionOriginalFilename                        string        `json:"submission_original_filename"`
		SubmissionOriginalURL                             interface{}   `json:"submission_original_url"`
		SubmissionPrescriptForceAdmin                     bool          `json:"submission_prescript_force_admin"`
		SubmissionPrescriptID                             interface{}   `json:"submission_prescript_id"`
		SubmissionPriority                                int           `json:"submission_priority"`
		SubmissionReputationMode                          string        `json:"submission_reputation_mode"`
		SubmissionRetentionPeriod                         int           `json:"submission_retention_period"`
		SubmissionSampleID                                int           `json:"submission_sample_id"`
		SubmissionSampleMd5                               string        `json:"submission_sample_md5"`
		SubmissionSampleSha1                              string        `json:"submission_sample_sha1"`
		SubmissionSampleSha256                            string        `json:"submission_sample_sha256"`
		SubmissionSampleSsdeep                            string        `json:"submission_sample_ssdeep"`
		SubmissionScore                                   int           `json:"submission_score"`
		SubmissionSeverity                                string        `json:"submission_severity"`
		SubmissionShareable                               bool          `json:"submission_shareable"`
		SubmissionSubmissionMetadata                      string        `json:"submission_submission_metadata"`
		SubmissionSystemTime                              interface{}   `json:"submission_system_time"`
		SubmissionTags                                    []interface{} `json:"submission_tags"`
		SubmissionTriageErrorHandling                     interface{}   `json:"submission_triage_error_handling"`
		SubmissionType                                    string        `json:"submission_type"`
		SubmissionUserAccountID                           int           `json:"submission_user_account_id"`
		SubmissionUserAccountName                         string        `json:"submission_user_account_name"`
		SubmissionUserAccountSubscriptionMode             string        `json:"submission_user_account_subscription_mode"`
		SubmissionUserEmail                               string        `json:"submission_user_email"`
		SubmissionUserID                                  int           `json:"submission_user_id"`
		SubmissionWebifURL                                string        `json:"submission_webif_url"`
		SubmissionWhoisMode                               string        `json:"submission_whois_mode"`
	}

	// VmraySubmissionStatusResponse represents the submission status response payload
	VmraySubmissionStatusResponse struct {
		Data   VmraySubmissionData `json:"data"`
		Result string              `json:"result"`
	}
)
