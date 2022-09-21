# ICAPeg monitoring using ELK stack

This is guid to setup ELK (Elasticsearch, Logstash and kibana) monitoring tool to monitor ICAPeg  

## Steps:
**1.[Setup ELK Stack](#setup-elk-stack)**<br>

**2.[Configure Logstash With ICAPeg Logs](#configure-logstash-with-icapeg-logs)**<br>

**3.[Run Logstash](#run-logstash)**<br>

**4.[Summary](#summary)**<br>


## Setup ELK Stack

 **You can download and install ELK from official website :** 
 - Elasticsearch 8.4 : [Link](https://www.elastic.co/guide/en/elasticsearch/reference/8.4/install-elasticsearch.html)
 - Logstash 8.4: [Link](https://www.elastic.co/guide/en/logstash/8.4/installing-logstash.html)
 - Kibana 8.4: [Link](https://www.elastic.co/guide/en/kibana/8.4/install.html)

 Follow setup instruction then start running Elasticsearch and kibana
  
## Configure Logstash With ICAPeg Logs

   Create `logstash.conf` file with this configurations:
   ```
   input {
		file {
			type => "json"
			codec => "json"
			path => "<ICAPeg_Path>/logs/logs.json"
			start_position => beginning
				}
			}

output {
		elasticsearch {
			hosts => ["<ElasticSearch_Address>"]
			index => "<Index_Name>"
			   		}
			}
   ```
   **Replace the following parts with correct data:**
   
   `<ICAPeg_Path>` : The absolute path of ICAPeg directory  ex: "/home/icapeg"

   `<ElasticSearch_Address>`: The address of elasticsearch ex: "localhost:9200"

   `<Index_Name>`: The Elasticsearch index name that created ex: "ICAPeg-logs"
   
## Run Logstash
  - **Go to logstsh Directory:**
   Run  `cd /usr/share/logstash` or go to logstash directory that you installed on.
   
   - **Run Logstash**
      ```
      sudo bin/logstash -f <logstash.conf path>
      ```
      
      Replace `<logstash.conf path>` with the absolute path to logstash.conf


      ## Summary
Now All stuff are working together, after any update on ICAPeg logs file will captured by Logstash and sored in Elasticsearch index.