## Log Levels

ICAPeg now has log levels, to determine what will be logged for a user to see. There are 4 levels at this moment:

### Info

The logs will contain whatever is necessary to indicate the overall work progress of the app like request/response received & final result of scan etc. Here is what you'll do in the ``config.toml`` file under ``app``:

  ```bash
   log_level = 'info'
  ```

### Error

The logs will contain just the error statements like the unexpected response status codes or any other failure throughout the app.

  ```bash
    log_level = 'error'
  ```

### Debug

The logs will contain everything including the ``info`` & the ``error`` logs, along with some of the dumps used in the app just for debugging purposes.

  ```bash
    log_level = 'debug'
  ```

### None

Just like its name suggests, the logs will contain nothing at all.

  ```bash
   log_level = 'none'
  ```



**NOTE**: Apart from the app start/end indicators, ICAPeg now logs everything in a ``logs.txt`` file which it creates during the starting of the server.
