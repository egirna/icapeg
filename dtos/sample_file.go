package dtos

type (
	// SampleInfo holds the informations regarding a sample file
	SampleInfo struct {
		FileName           string
		SampleType         string
		SampleSeverity     string
		VTIScore           string
		FileSizeStr        string
		SubmissionFinished bool
	}
	// SubmitResponse holds the informations regarding the submit response
	SubmitResponse struct {
		SubmissionExists   bool
		SubmissionID       string
		SubmissionSampleID string
	}
	// SubmissionStatusResponse holds the information regarding the submission status response
	SubmissionStatusResponse struct {
		SubmissionFinished bool
	}
	// TemplateData represents the data needed to be show in the custom html template
	TemplateData struct {
		FileName     string `json:"file_name"`
		Severity     string `json:"severity"`
		RequestedURL string `json:"requested_url"`
		Score        string `json:"score"`
		FileType     string `json:"file_type"`
		FileSizeStr  string `json:"file_size_str"`
		ResultsBy    string `json:"results_by"`
	}

	// FileMetaInfo represents the meta data regarding the concerned file
	FileMetaInfo struct {
		FileName string
		FileType string
		FileSize float64
	}
)
