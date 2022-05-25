## Config file

- ### **Reading from env variable if value is not in toml file**:

  - This feature is supported for strings, int, bool, time.duration and string slices only.

  - To use this feature you have to do the next:

    - Assume that there is an env variable called **LOG_LEVEL** and you want to assign **LOG_LEVEL** value to **app.log_level**, you will change the value of:

      ```
      log_level= "debug"
      ```

      to:

      ```
      log_level= "$_LOG_LEVEL"
      ```

    - If you want to add an array as an env variable in your machine, please add backslash before special characters and "\n" instead of new lines, you can use this [tool](https://www.freeformatter.com/json-escape.html) to do that

      example:

      ```
      export ARRAY= "[\"txt\", \"pdf\", \"dmg\", \"exe\", \"com\", \"rar\", \"unknown\"]"
      ```

      Don't forget to put the value between double quotes in case there are white spaces in the value.

    - Before you use this feature please make sure that the env variable that you want to use is globally in your machine and not just exported in a local session.

- ### **Policy**

  - Please before adding policy JSON file add "\" before special characters and "\n" instead of new lines by this [tool](https://www.freeformatter.com/json-escape.html).

  - You shouldn't remove the policy variable from config file.

  - if you don't want to set a value for policy you should do like that

    ```
    policy = ""
    ```

  - Don't forget to put the value between double quotes incase you want to add it as an env variable in your machine because there are white spaces in the value.

    

## Testing REQMOD and RESPMOD

- ### REQMOD

  Try to upload a PDF file using this [website](https://filebin.net/), then download it to check if it was scanned by Glasswall API or not.

  

- ### RESPMOD 

  Try to google "PDF sample download", then open any result which includes a PDF file and check if it was scanned by Glasswall API or not.