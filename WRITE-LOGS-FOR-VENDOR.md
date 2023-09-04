## **How to write logs for a vendor?**

```go
logging.Logger.Debug(utils.PrepareLogMsg(<the X-ICAP-Metadata of the service>, <The message you want to log>))
```
