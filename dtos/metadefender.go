package dtos

type (
	// MetaDefenderScanFileResponse represents the scan file endpoint response payload of MetaDefender
	MetaDefenderScanFileResponse struct {
		DataID        string `json:"data_id"`
		Status        string `json:"status"`
		InQueue       uint   `json:"in_queue"`
		QueuePriority string `json:"queue_priority"`
		Sha1          string `json:"sha1"`
		Sha256        string `json:"sha256"`
	}
	// MetaDefenderErrorResponse represents the error response for the MetaDefender service
	MetaDefenderErrorResponse struct {
		Error struct {
			Code     int      `json:"code"`
			Messages []string `json:"messages"`
		} `json:"error"`
	}
	// MDScan scan detail for MetaDefender
	MDScan struct {
		ScanTime    int    `json:"scan_time"`
		DefTime     string `json:"def_time"`
		ScanResultI int    `json:"scan_result_i"`
		ThreatFound string `json:"threat_found"`
	}
	// MetaDefenderReportResponse represents the report response payload of the MetaDefender service
	MetaDefenderReportResponse struct {
		DataID      string `json:"data_id"`
		Status      string `json:"status"`
		InQueue     string `json:"in_queue"`
		LastUpdated string `json:"last_updated"`
		FileID      string `json:"file_id"`
		ScanResults struct {
			ScanDetails struct {
				LavaSoft         MDScan `json:"Lavasoft"`
				STOPzilla        MDScan `json:"STOPzilla"`
				Zillya           MDScan `json:"Zillya!"`
				VirusBlokAda     MDScan `json:"VirusBlokAda"`
				TrendMicro       MDScan `json:"TrendMicro"`
				SuperAntiSpyware MDScan `json:"SUPERAntiSpyware"`
				NProtect         MDScan `json:"nProtect"`
				NANOAV           MDScan `json:"NANOAV"`
				FSecure          MDScan `json:"F-secure"`
				Eset             MDScan `json:"ESET"`
				BitDefender      MDScan `json:"BitDefender"`
				Baidu            MDScan `json:"Baidu"`
				Ahnlab           MDScan `json:"Ahnlab"`
				AegisLab         MDScan `json:"AegisLab"`
				Zoner            MDScan `json:"Zoner"`
				ThreatTrack      MDScan `json:"ThreatTrack"`
				Sophos           MDScan `json:"Sophos"`
				Preventon        MDScan `json:"Preventon"`
				Mcafee           MDScan `json:"McAfee"`
				K7               MDScan `json:"K7"`
				Jiangmin         MDScan `json:"Jiangmin"`
				Hauri            MDScan `json:"Hauri"`
				Fprot            MDScan `json:"F-prot"`
				Fortinet         MDScan `json:"Fortinet"`
				Filseclab        MDScan `json:"Filseclab"`
				Emsisoft         MDScan `json:"Emsisoft"`
				ClamAV           MDScan `json:"ClamAV"`
				ByteHero         MDScan `json:"ByteHero"`
				Avira            MDScan `json:"Avira"`
				AVG              MDScan `json:"AVG"`
				Agnitum          MDScan `json:"Agnitum"`
				Ikarus           MDScan `json:"Ikarus"`
				Cyren            MDScan `json:"Cyren"`
				MicrosoftSE      MDScan `json:"Microsoft Security Essentials"`
				QuickHeal        MDScan `json:"Quick Heal"`
				TotalDefense     MDScan `json:"Total Defense"`
				TrendMicroHC     MDScan `json:"TrendMicro House Call"`
				XvirusPG         MDScan `json:"Xvirus Personal Guard"`
				DrWebGateway     MDScan `json:"Dr.Web Gateway"`
				VirITeXplorer    MDScan `json:"Vir.IT eXplorer"`
			} `json:"scan_details"`
			RescanAvailable    bool   `json:"rescan_available"`
			ScanAllResultI     int    `json:"scan_all_result_i"`
			StartTime          string `json:"start_time"`
			TotalTime          int64  `json:"total_time"`
			TotalAvs           int    `json:"total_avs"`
			TotalDetectedAvs   int    `json:"total_detected_avs"`
			ProgressPercentage int    `json:"progress_percentage"`
			ScanAllRes         string `json:"scan_all_result_a"`
		} `json:"scan_results"`

		FileInfo struct {
			FileSize        int64  `json:"file_size"`
			UploadTimeStamp string `json:"upload_timestamp"`
			MD5             string `json:"md5"`
			Sha1            string `json:"sha1"`

			Sha256 string `json:"sha256"`

			FileTC      string `json:"file_type_category"`
			FileTD      string `json:"file_type_description"`
			FileTE      string `json:"file_type_extension"`
			DisplayName string `json:"display_name"`
		} `json:"file_info"`
		ShareFile      int      `json:"share_file"`
		RestVersion    string   `json:"rest_version"`
		AdditionalInfo []string `json:"additional_info"`
		Votes          struct {
			Up   int `json:"up"`
			Down int `json:"down"`
		}
	}
)
