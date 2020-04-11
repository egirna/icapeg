package transformers

import (
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"strconv"
)

// TransformVmrayToSampleInfo transforms a vmray sample response to geenric sample info
func TransformVmrayToSampleInfo(sr *dtos.GetVmraySampleResponse) *dtos.SampleInfo {
	return &dtos.SampleInfo{
		FileName:       sr.Data.SampleFilename,
		SampleType:     sr.Data.SampleType,
		SampleSeverity: sr.Data.SampleSeverity,
		VTIScore:       fmt.Sprintf("%v/100", sr.Data.SampleVtiScore),
		FileSizeStr:    fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(sr.Data.SampleFilesize)),
	}
}

// TransformVmrayToSubmitResponse transforms a vmray submit response to generic submit response
func TransformVmrayToSubmitResponse(vsr *dtos.VmraySubmitResponse) *dtos.SubmitResponse {
	sr := &dtos.SubmitResponse{}
	if len(vsr.Data.Submissions) > 0 {
		sr.SubmissionExists = true
	}

	if sr.SubmissionExists {
		sr.SubmissionID = strconv.Itoa(vsr.Data.Submissions[0].SubmissionID)
		sr.SubmissionSampleID = strconv.Itoa(vsr.Data.Submissions[0].SubmissionSampleID)
	}

	return sr
}

// TransformVmrayToSubmissionStatusResponse transforms a vmray submission status response to the generic one
func TransformVmrayToSubmissionStatusResponse(stsr *dtos.VmraySubmissionStatusResponse) *dtos.SubmissionStatusResponse {
	return &dtos.SubmissionStatusResponse{
		SubmissionFinished: stsr.Data.SubmissionFinished,
	}
}
