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
		FileName     string
		Severity     string
		RequestedURL string
		VTIScore     string
		FileType     string
		FileSizeStr  string
	}

	// FileMetaInfo represents the meta data regarding the concerned file
	FileMetaInfo struct {
		FileName string
		FileType string
		FileSize float64
	}
)
