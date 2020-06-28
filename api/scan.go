package api

import (
	"bytes"
	"icapeg/dtos"
	"icapeg/service"
	"icapeg/utils"
	"net/http"
	"time"
)

func doScan(scannerName, filename string, fmi dtos.FileMetaInfo, buf *bytes.Buffer, fileURL string) (int, *dtos.SampleInfo) {

	var sts int
	var si *dtos.SampleInfo

	localService := service.IsServiceLocal(scannerName)

	if localService && buf != nil { // if the scanner is installed locally
		sts, si = doLocalScan(scannerName, fmi, buf)
	}

	if !localService { // if the scanner is an external service requiring API calls.

		if buf == nil && fileURL != "" { // indicates this is a URL scan request
			sts, si = doRemoteURLScan(scannerName, filename, fmi, fileURL)
		}

		if buf != nil && fileURL == "" { // indicates this is a File scan request
			sts, si = doRemoteFileScan(scannerName, filename, fmi, buf)
		}

	}

	return sts, si
}

func doLocalScan(scannerName string, fmi dtos.FileMetaInfo, buf *bytes.Buffer) (int, *dtos.SampleInfo) {
	lsvc := service.GetLocalService(scannerName)

	if lsvc == nil {
		debugLogger.LogToFile("No such scanner vendors:", scannerName)
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	if !lsvc.RespSupported() {
		debugLogger.LogfToFile("The vendor %s does not support respmod of icap\n", scannerName)
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	sampleInfo, err := lsvc.ScanFileStream(buf, fmi)
	if err != nil {
		errorLogger.LogToFile("Couldn't fetch sample information for local service: ", err.Error())
		return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
	}

	debugLogger.DumpToFile("result", sampleInfo)

	if !utils.InStringSlice(sampleInfo.SampleSeverity, lsvc.GetOkFileStatus()) { // checking is the sample severity is amongst the allowable file status
		return http.StatusOK, sampleInfo
	}

	return http.StatusNoContent, nil
}

func doRemoteFileScan(scannerName, filename string, fmi dtos.FileMetaInfo, buf *bytes.Buffer) (int, *dtos.SampleInfo) {

	svc := service.GetService(scannerName)

	if svc == nil {
		debugLogger.LogToFile("No such scanner vendors:", scannerName)
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	if !svc.RespSupported() {
		debugLogger.LogfToFile("The vendor %s does not support respmod of icap\n", scannerName)
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	// The submit file api call is commented out for safety for now
	submitResp, err := svc.SubmitFile(buf, filename) // submitting the file for analysing
	if err != nil {
		errorLogger.LogfToFile("Failed to submit file to %s: %s\n", scannerName, err.Error())
		return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
	}

	if !submitResp.SubmissionExists {
		debugLogger.LogToFile("No submissions for the file")
		return http.StatusNoContent, nil
	}

	debugLogger.DumpToFile("submit response", submitResp)

	submissionFinished := false
	statusCheckFinishTime := time.Now().Add(svc.GetStatusCheckTimeout()) // the time after which, the system is to stop checking for submission finish
	var sampleInfo *dtos.SampleInfo
	sampleID := submitResp.SubmissionSampleID //"4715575"

	for !submissionFinished && time.Now().Before(statusCheckFinishTime) { // while the time dedicated for status checking has no expired
		submissionID := submitResp.SubmissionID //"5651578"

		switch svc.StatusEndpointExists() { // this blocks acts depending of the fact that the scanner has a seperate endpoint for checking file scan status or not, somne of them has and some don't
		case true:
			submissionStatus, err := svc.GetSubmissionStatus(submissionID) // getting the file submission status by the submission id received by submitting the file
			if err != nil {
				errorLogger.LogfToFile("Failed to get submission status from %s: %s\n", scannerName, err.Error())
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}

			debugLogger.DumpToFile("submission status resp", submissionStatus)

			submissionFinished = submissionStatus.SubmissionFinished
		case false: // if it doesn;t the file report result will contain the information
			var err error
			sampleInfo, err = svc.GetSampleFileInfo(sampleID, fmi)
			if err != nil {
				errorLogger.LogToFile("Couldn't fetch sample information during status check: ", err.Error())
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}
			submissionFinished = sampleInfo.SubmissionFinished
		default:
			debugLogger.LogToFile("Put the status_endpoint_exists field in the config file under the scanner vendor")
		}

		if !submissionFinished { // if the submission is not finished, wait for a certain time and then call again
			time.Sleep(svc.GetStatusCheckInterval())
		}
	}

	if !submissionFinished {
		debugLogger.LogToFile("File submission is taking too long to finish")
		return http.StatusNoContent, nil
	}

	if sampleInfo == nil {
		var err error
		sampleInfo, err = svc.GetSampleFileInfo(sampleID, fmi) // getting the results after scanner is done analysing the file
		if err != nil {
			errorLogger.LogToFile("Couldn't fetch sample information after submission finish: ", err.Error())
			return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
		}
	}

	debugLogger.DumpToFile("result", sampleInfo)

	if !utils.InStringSlice(sampleInfo.SampleSeverity, svc.GetOkFileStatus()) { // checking is the sample severity is amongst the allowable file status

		return http.StatusOK, sampleInfo
	}

	return http.StatusNoContent, nil
}

func doRemoteURLScan(scannerName, filename string, fmi dtos.FileMetaInfo, fileURL string) (int, *dtos.SampleInfo) {
	svc := service.GetService(scannerName)

	if svc == nil {
		debugLogger.LogToFile("No such scanner vendors:", scannerName)
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	if !svc.ReqSupported() {
		debugLogger.LogfToFile("The vendor %s does not support reqmod of icap\n", scannerName)
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	// The submit file api call is commented out for safety for now
	submitResp, err := svc.SubmitURL(fileURL, filename) // submitting the file for analysing
	if err != nil {
		errorLogger.LogfToFile("Failed to submit url to %s: %s\n", scannerName, err.Error())
		return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
	}

	if !submitResp.SubmissionExists {
		debugLogger.LogToFile("No submissions for the url")
		return http.StatusNoContent, nil
	}

	debugLogger.DumpToFile("submit response", submitResp)

	submissionFinished := false
	statusCheckFinishTime := time.Now().Add(svc.GetStatusCheckTimeout()) // the time after which, the system is to stop checking for submission finish
	var sampleInfo *dtos.SampleInfo
	sampleID := submitResp.SubmissionSampleID //"4715575"

	for !submissionFinished && time.Now().Before(statusCheckFinishTime) { // while the time dedicated for status checking has no expired
		submissionID := submitResp.SubmissionID //"5651578"

		switch svc.StatusEndpointExists() { // this blocks acts depending of the fact that the scanner has a seperate endpoint for checking file scan status or not, somne of them has and some don't
		case true:
			submissionStatus, err := svc.GetSubmissionStatus(submissionID) // getting the url submission status by the submission id received by submitting the url
			if err != nil {
				errorLogger.LogfToFile("Failed to get submission status from %s: %s\n", scannerName, err.Error())
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}

			debugLogger.DumpToFile("submission status resp", submissionStatus)

			submissionFinished = submissionStatus.SubmissionFinished
		case false: // if it doesn;t the url report result will contain the information
			var err error
			sampleInfo, err = svc.GetSampleURLInfo(sampleID, fmi)
			if err != nil {
				errorLogger.LogToFile("Couldn't fetch sample information during status check: ", err.Error())
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}
			submissionFinished = sampleInfo.SubmissionFinished
		default:
			debugLogger.LogToFile("Put the status_endpoint_exists field in the config file under the scanner vendor")
		}

		if !submissionFinished { // if the submission is not finished, wait for a certain time and then call again
			time.Sleep(svc.GetStatusCheckInterval())
		}
	}

	if !submissionFinished {
		debugLogger.LogToFile("URL submission is taking too long to finish")
		return http.StatusNoContent, nil
	}

	if sampleInfo == nil {
		var err error
		sampleInfo, err = svc.GetSampleURLInfo(sampleID, fmi) // getting the results after scanner is done analysing the file
		if err != nil {
			errorLogger.LogToFile("Couldn't fetch sample information after submission finish: ", err.Error())
			return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
		}
	}

	debugLogger.DumpToFile("result", sampleInfo)

	if !utils.InStringSlice(sampleInfo.SampleSeverity, svc.GetOkFileStatus()) { // checking is the sample severity is amongst the allowable url status
		return http.StatusOK, sampleInfo
	}
	return http.StatusNoContent, nil
}
