package dtos

type (
	// VirusTotalScanFileResponse represents the scan file endpoint reponse payload of virustotal
	VirusTotalScanFileResponse struct {
		ScanID       string `json:"scan_id"`
		Sha1         string `json:"sha1"`
		Resource     string `json:"resource"`
		ResponseCode int    `json:"response_code"`
		Sha256       string `json:"sha256"`
		Permalink    string `json:"permalink"`
		Md5          string `json:"md5"`
		VerboseMsg   string `json:"verbose_msg"`
	}
	// Scanner represents the antivirus scanner data in the report response payload of virustotal
	Scanner struct {
		Detected bool   `json:"detected"`
		Version  string `json:"version"`
		Result   string `json:"result"`
		Update   string `json:"update"`
	}
	// VirusTotalReportResponse represents the report reponse payload of the virustotal service
	VirusTotalReportResponse struct {
		Scans struct {
			Bkav                  Scanner `json:"Bkav"`
			TotalDefense          Scanner `json:"TotalDefense"`
			MicroWorldEScan       Scanner `json:"MicroWorld-eScan"`
			FireEye               Scanner `json:"FireEye"`
			CATQuickHeal          Scanner `json:"CAT-QuickHeal"`
			McAfee                Scanner `json:"McAfee"`
			Malwarebytes          Scanner `json:"Malwarebytes"`
			VIPRE                 Scanner `json:"VIPRE"`
			AegisLab              Scanner `json:"AegisLab"`
			Sangfor               Scanner `json:"Sangfor"`
			K7AntiVirus           Scanner `json:"K7AntiVirus"`
			BitDefender           Scanner `json:"BitDefender"`
			K7GW                  Scanner `json:"K7GW"`
			Baidu                 Scanner `json:"Baidu"`
			FProt                 Scanner `json:"F-Prot"`
			SymantecMobileInsight Scanner `json:"SymantecMobileInsight"`
			ESETNOD32             Scanner `json:"ESET-NOD32"`
			APEX                  Scanner `json:"APEX"`
			Avast                 Scanner `json:"Avast"`
			ClamAV                Scanner `json:"ClamAV"`
			Kaspersky             Scanner `json:"Kaspersky"`
			Alibaba               Scanner `json:"Alibaba"`
			NANOAntivirus         Scanner `json:"NANO-Antivirus"`
			ViRobot               Scanner `json:"ViRobot"`
			Rising                Scanner `json:"Rising"`
			AdAware               Scanner `json:"Ad-Aware"`
			Emsisoft              Scanner `json:"Emsisoft"`
			Comodo                Scanner `json:"Comodo"`
			FSecure               Scanner `json:"F-Secure"`
			DrWeb                 Scanner `json:"DrWeb"`
			Zillya                Scanner `json:"Zillya"`
			TrendMicro            Scanner `json:"TrendMicro"`
			McAfeeGWEdition       Scanner `json:"McAfee-GW-Edition"`
			CMC                   Scanner `json:"CMC"`
			Sophos                Scanner `json:"Sophos"`
			SentinelOne           Scanner `json:"SentinelOne"`
			Cyren                 Scanner `json:"Cyren"`
			Jiangmin              Scanner `json:"Jiangmin"`
			Webroot               Scanner `json:"Webroot"`
			Avira                 Scanner `json:"Avira"`
			Fortinet              Scanner `json:"Fortinet"`
			AntiyAVL              Scanner `json:"Antiy-AVL"`
			Kingsoft              Scanner `json:"Kingsoft"`
			Endgame               Scanner `json:"Endgame"`
			Arcabit               Scanner `json:"Arcabit"`
			SUPERAntiSpyware      Scanner `json:"SUPERAntiSpyware"`
			ZoneAlarm             Scanner `json:"ZoneAlarm"`
			AvastMobile           Scanner `json:"Avast-Mobile"`
			Microsoft             Scanner `json:"Microsoft"`
			TACHYON               Scanner `json:"TACHYON"`
			AhnLabV3              Scanner `json:"AhnLab-V3"`
			VBA32                 Scanner `json:"VBA32"`
			ALYac                 Scanner `json:"ALYac"`
			MAX                   Scanner `json:"MAX"`
			Zoner                 Scanner `json:"Zoner"`
			TrendMicroHouseCall   Scanner `json:"TrendMicro-HouseCall"`
			Tencent               Scanner `json:"Tencent"`
			Yandex                Scanner `json:"Yandex"`
			Ikarus                Scanner `json:"Ikarus"`
			MaxSecure             Scanner `json:"MaxSecure"`
			GData                 Scanner `json:"GData"`
			BitDefenderTheta      Scanner `json:"BitDefenderTheta"`
			AVG                   Scanner `json:"AVG"`
			Panda                 Scanner `json:"Panda"`
			Qihoo360              Scanner `json:"Qihoo-360"`
		} `json:"scans"`
		ScanID       string `json:"scan_id"`
		Sha1         string `json:"sha1"`
		Resource     string `json:"resource"`
		ResponseCode int    `json:"response_code"`
		ScanDate     string `json:"scan_date"`
		Permalink    string `json:"permalink"`
		VerboseMsg   string `json:"verbose_msg"`
		Total        int    `json:"total"`
		Positives    int    `json:"positives"`
		Sha256       string `json:"sha256"`
		Md5          string `json:"md5"`
	}
)
