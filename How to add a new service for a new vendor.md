## **How to add a new service for a new vendor?**

Assume that the name of the vendor is **abc** and the name of the service **xyz**, Let's start with the steps: 

- ### **Config.toml**

  You should add asection with the name of the service in [**config.toml**](./config.toml).

  You should add the required variables of the service.

  For example if the name of the service is **xyz** and its vendor is **abc**, so the section will be like the following section:

  ```toml
  [xyz]
  vendor = "abc"
  req_mode=true
  resp_mode=true
  base_url = "http://abc/" #
  scan_endpoint = "xyz.com"
  api_key = "<api key>"
  timeout  = 300 #seconds , ICAP will return 408 - Request timeout
  fail_threshold = 2
  #max file size value from 1 to 9223372036854775807, and value of zero means unlimited
  max_filesize = 0 #bytes
  ```

- ### **Vendor package**

  - Add a new package named **abc** in the [service](./service) directory.

  - Add file named **abc.go** in **abc** package.

  - Create a struct named **Abc** with the required attributes of any service to this vendor.

    ```go
    type Abc struct {
    	httpMsg                           *utils.HttpMsg
    	elapsed                           time.Duration
    	serviceName                       string
    	methodName                        string
    	maxFileSize                       int
    	BaseURL                           string
    	Timeout                           time.Duration
    	APIKey                            string
    	ScanEndpoint                      string
    	ReportEndpoint                    string
    	FailThreshold                     int
        generalFunc                       *general_functions.GeneralFunc //this not neccessary, you just can use it to make use of its function
    	logger                            *logger.ZLogger
    }
    ```

  - Add function named **NewAbcService** which create service from abc vendor

    ```go
    // NewGlasswallService returns a new populated instance of the Glasswall service
    func NewGlasswallService(serviceName, methodName string, httpMsg *utils.HttpMsg, elapsed time.Duration, logger *logger.ZLogger) *Abc {
    	abcService := &Glasswall{
    		httpMsg:                           httpMsg,
    		elapsed:                           elapsed,
    		serviceName:                       serviceName,
    		methodName:                        methodName,
    		maxFileSize:                       readValues.ReadValuesInt(serviceName + ".max_filesize"),
    		BaseURL:                           readValues.ReadValuesString(serviceName + ".base_url"),
    		Timeout:                           readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
    		APIKey:                            readValues.ReadValuesString(serviceName + ".api_key"),
    		ScanEndpoint:                      readValues.ReadValuesString(serviceName + ".scan_endpoint"),
    		ReportEndpoint:                    "/",
    		FailThreshold:                     readValues.ReadValuesInt(serviceName + ".fail_threshold"),
    		statusCheckInterval:               2 * time.Second,
    		respSupported:                     readValues.ReadValuesBool(serviceName + 
    		logger:                            logger,
    	}
    	return abcService
    }
    ```

  - Add function **Processing** to **abc** struct to implement the interface

    ```go
    //Processing is a func used for to processing the http message
    func (g *Glasswall) Processing() (int, interface{}, map[string]string) {
    	// implementation
    }
    
    ```

    

- ### **service.go**

  - Add a name of the new vendor as a constant variable in [service.go](./service/servoce.go) at the start of the file in the const file.

    ```go
    //Vendors names
    const (
    	VendorGlasswall = "glasswall"
        VendorABC = "abc"
    )
    ```

  - You should add a case in the switch case in GetService function by the new vendor
  
    ```go
    func GetService(vendor, serviceName, methodName string, httpMsg *utils.HttpMsg, elapsed time.Duration, logger *logger.ZLogger) Service {
    	switch vendor {
    	case VendorGlasswall:
    		return glasswall.NewGlasswallService(serviceName, methodName, httpMsg, elapsed, logger)
    	case VendorABC:
    		return echo.NewABCService(serviceName, methodName, httpMsg, elapsed, logger)
    	}
    	return nil
    }
    ```

Now you can run the application and try it with you service.