package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"icapeg/config"
	"icapeg/dtos"
	"icapeg/logger"
	"icapeg/service"
	"icapeg/utils"

	zLog "github.com/rs/zerolog/log"
)

func doScan(scannerName, serviceName string, filename string, fmi dtos.FileMetaInfo, buf *bytes.Buffer, fileURL string, logger *logger.ZLogger) (int, *dtos.SampleInfo) {

	if config.App().RespScannerVendorShadow != utils.NoVendor || config.App().ReqScannerVendorShadow != utils.NoVendor {
		go doShadowScan(scannerName, serviceName, filename, fmi, buf, fileURL, logger)
	}

	newBuf := utils.CopyBuffer(buf)

	var sts int
	var si *dtos.SampleInfo

	localService := service.IsServiceLocal(scannerName, serviceName, logger)

	if localService && buf != nil { // if the scanner is installed locally
		sts, si = doLocalScan(scannerName, serviceName, fmi, newBuf, logger)
	}

	if !localService { // if the scanner is an external service requiring API calls.

		if buf == nil && fileURL != "" { // indicates this is a URL scan request
			sts, si = doRemoteURLScan(scannerName, serviceName, filename, fmi, fileURL, logger)
		}

		if buf != nil && fileURL == "" { // indicates this is a File scan request
			sts, si = doRemoteFileScan(scannerName, serviceName, filename, fmi, newBuf, logger)
		}

	}

	return sts, si
}

func doLocalScan(scannerName string, serviceName string, fmi dtos.FileMetaInfo, buf *bytes.Buffer, logger *logger.ZLogger) (int, *dtos.SampleInfo) {
	lsvc := service.GetLocalService(scannerName, serviceName)

	if lsvc == nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("No such scanner vendors: %s", scannerName)).Msgf("no_vendor_supported")
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	if !lsvc.RespSupported() {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("The vendor %s does not support respmod of icap\n", scannerName)).Msgf("resp_mode_not_supported_by_vendor")
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	sampleInfo, err := lsvc.ScanFileStream(buf, fmi)
	if err != nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Couldn't fetch sample information for local service").Msgf("fail_to_fetch_sample_info_from_local_service")
		return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
	}

	if !utils.InStringSlice(sampleInfo.SampleSeverity, lsvc.GetOkFileStatus()) { // checking is the sample severity is amongst the allowable file status
		return http.StatusOK, sampleInfo
	}

	return http.StatusNoContent, nil
}

func doRemoteFileScan(scannerName, serviceName string, filename string, fmi dtos.FileMetaInfo, buf *bytes.Buffer, logger *logger.ZLogger) (int, *dtos.SampleInfo) {

	svc := service.GetService(scannerName, serviceName, logger)

	if svc == nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("No such scanner vendors: %s", scannerName)).Msgf("no_vendor_supported")
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	if !svc.RespSupported() {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("The vendor %s does not support respmod of icap\n", scannerName)).Msgf("resp_mode_not_supported_by_vendor")
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	// The submit file api call is commented out for safety for now
	submitResp, err := svc.SubmitFile(buf, filename) // submitting the file for analysing
	if err != nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", fmt.Sprintf("Failed to submit file to %s", scannerName)).Msgf("submit_file_failed_to_scanner")
		return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
	}

	if !submitResp.SubmissionExists {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "No submissions for the file").Msgf("no_file_submission")
		return http.StatusNoContent, nil
	}

	submissionFinished := false
	statusCheckFinishTime := time.Now().Add(svc.GetStatusCheckTimeout()) // the time after which, the system is to stop checking for submission finish
	var sampleInfo *dtos.SampleInfo
	sampleID := submitResp.SubmissionSampleID // "4715575"

	for !submissionFinished && time.Now().Before(statusCheckFinishTime) { // while the time dedicated for status checking has no expired
		submissionID := submitResp.SubmissionID // "5651578"

		switch svc.StatusEndpointExists() { // this blocks acts depending of the fact that the scanner has a seperate endpoint for checking file scan status or not, somne of them has and some don't
		case true:
			submissionStatus, err := svc.GetSubmissionStatus(submissionID) // getting the file submission status by the submission id received by submitting the file
			if err != nil {
				elapsed := time.Since(logger.LogStartTime)
				zLog.Error().Dur("duration", elapsed).Err(err).Str("value", fmt.Sprintf("Failed to get submission status from %s", scannerName)).Msgf("fail_to_retrieve_submission_status")
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}

			submissionFinished = submissionStatus.SubmissionFinished
		case false: // if it doesn;t the file report result will contain the information
			var err error
			sampleInfo, err = svc.GetSampleFileInfo(sampleID, fmi)
			if err != nil {
				elapsed := time.Since(logger.LogStartTime)
				zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Couldn't fetch sample information during status check").Msgf("fail_to_fetch_file_info")
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}
			submissionFinished = sampleInfo.SubmissionFinished
		default:
			elapsed := time.Since(logger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", "Put the status_endpoint_exists field in the config file under the scanner vendor").Msgf("status_endpoint_missing_in_config_file")
		}

		if !submissionFinished { // if the submission is not finished, wait for a certain time and then call again
			time.Sleep(svc.GetStatusCheckInterval())
		}
	}

	if !submissionFinished {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "File submission is taking too long to finish").Msgf("file_submission_delayed")
		return http.StatusNoContent, nil
	}

	if sampleInfo == nil {
		var err error
		sampleInfo, err = svc.GetSampleFileInfo(sampleID, fmi) // getting the results after scanner is done analysing the file
		if err != nil {
			elapsed := time.Since(logger.LogStartTime)
			zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Couldn't fetch sample information after submission finish").Msgf("fail_to_fetch_file_info")
			return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
		}
	}

	if !utils.InStringSlice(sampleInfo.SampleSeverity, svc.GetOkFileStatus()) { // checking is the sample severity is amongst the allowable file status

		return http.StatusOK, sampleInfo
	}

	return http.StatusNoContent, nil
}

func doRemoteURLScan(scannerName, serrviceName string, filename string, fmi dtos.FileMetaInfo, fileURL string, logger *logger.ZLogger) (int, *dtos.SampleInfo) {
	svc := service.GetService(scannerName, serrviceName, logger)
	var elapsed time.Duration
	if svc == nil {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("No such scanner vendors: %s", scannerName)).Msgf("no_vendor_supported")
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	if !svc.ReqSupported() {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("The vendor %s does not support reqmod of icap\n", scannerName)).Msgf("resp_mode_not_supported_by_vendor")
		return utils.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil
	}

	// The submit file api call is commented out for safety for now
	submitResp, err := svc.SubmitURL(fileURL, filename) // submitting the file for analysing
	if err != nil {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", fmt.Sprintf("Failed to submit file to %s", scannerName)).Msgf("submit_file_failed_to_scanner")
		return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
	}

	if !submitResp.SubmissionExists {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "No submissions for the file").Msgf("no_file_submission")
		return http.StatusNoContent, nil
	}

	submissionFinished := false
	statusCheckFinishTime := time.Now().Add(svc.GetStatusCheckTimeout()) // the time after which, the system is to stop checking for submission finish
	var sampleInfo *dtos.SampleInfo
	sampleID := submitResp.SubmissionSampleID // "4715575"

	for !submissionFinished && time.Now().Before(statusCheckFinishTime) { // while the time dedicated for status checking has no expired
		submissionID := submitResp.SubmissionID // "5651578"

		switch svc.StatusEndpointExists() { // this blocks acts depending of the fact that the scanner has a seperate endpoint for checking file scan status or not, somne of them has and some don't
		case true:
			submissionStatus, err := svc.GetSubmissionStatus(submissionID) // getting the url submission status by the submission id received by submitting the url
			if err != nil {
				elapsed = time.Since(logger.LogStartTime)
				zLog.Error().Dur("duration", elapsed).Err(err).Str("value", fmt.Sprintf("Failed to get submission status from %s", scannerName)).Msgf("fail_to_retrieve_submission_status")
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}

			submissionFinished = submissionStatus.SubmissionFinished
		case false: // if it doesn;t the url report result will contain the information
			var err error
			sampleInfo, err = svc.GetSampleURLInfo(sampleID, fmi)
			if err != nil {
				elapsed = time.Since(logger.LogStartTime)
				zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Couldn't fetch sample information during status check").Msgf("fail_to_fetch_file_info")
				return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
			}
			submissionFinished = sampleInfo.SubmissionFinished
		default:
			elapsed = time.Since(logger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", "Put the status_endpoint_exists field in the config file under the scanner vendor").Msgf("status_endpoint_missing_in_config_file")
		}

		if !submissionFinished { // if the submission is not finished, wait for a certain time and then call again
			time.Sleep(svc.GetStatusCheckInterval())
		}
	}

	if !submissionFinished {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "File submission is taking too long to finish").Msgf("file_submission_delayed")
		return http.StatusNoContent, nil
	}

	if sampleInfo == nil {
		var err error
		sampleInfo, err = svc.GetSampleURLInfo(sampleID, fmi) // getting the results after scanner is done analysing the file
		if err != nil {
			elapsed = time.Since(logger.LogStartTime)
			zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Couldn't fetch sample information after submission finish").Msgf("fail_to_fetch_file_info")
			return utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil
		}
	}

	if !utils.InStringSlice(sampleInfo.SampleSeverity, svc.GetOkFileStatus()) { // checking is the sample severity is amongst the allowable url status
		return http.StatusOK, sampleInfo
	}
	return http.StatusNoContent, nil
}

// DoCDR send req to api return resp to client
func DoCDR(scannerName string, serviceName string, f *bytes.Buffer, filename string, reqURL string, logger *logger.ZLogger) (*http.Response, int, bool, string, error) {

	svc := service.GetService(scannerName, serviceName, logger)
	if svc == nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("No such scanner vendors: %s", scannerName)).Msgf("no_vendor_supported")
		err := errors.New("No such scanner vendors: %s" + scannerName)
		return nil, utils.BadRequestStatusCodeStr, false, "", err
	}

	submitResp, ICAPStatusCode, html, x_adaption_id, err := svc.SendFileApi(f, filename, reqURL) // submitting the file for analysing
	if err != nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", fmt.Sprintf("Failed to submit file to %s", scannerName)).Msgf("submit_file_failed_to_scanner")
		return nil, utils.BadRequestStatusCodeStr, false, "", err
	}

	return submitResp, ICAPStatusCode, html, x_adaption_id, err
}
