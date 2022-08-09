from ast import Is
from concurrent.futures import process
from email.header import Header
from inspect import istraceback
import os.path
import subprocess
import sys
import hashlib
import csv
from setuptools import Command
import toml
import time


passed_tests = 0
failed_tests = 0 
# start ICAPeg 
subprocess.run(['./icapeg 2> /dev/null &'],shell=True)
time.sleep(10)



class style:
    def header(text):
        COLOR = '\033[1;37;45m'
        ENDC = '\033[0m'
        print('\033[1m' + COLOR + text + ENDC)

    def ok(text, explane = ""):
        COLOR = '\033[92m'
        ENDC = '\033[0m'
        print('\033[1m' + COLOR + '\u2705 ' + text + ENDC + explane)

    def fail(text, explane = ""):
        COLOR = '\033[91m'
        ENDC = '\033[0m'
        print('\033[1m' + COLOR + '\u274c ' + text + ENDC + explane)

    def warning(text, explane = ""):
        COLOR = '\033[93m'
        ENDC = '\033[0m'
        print('\033[1m' + COLOR + '\u26A0 ' + text + ENDC + explane)
    def yellow(text, explane = ""):
        COLOR = '\033[93m'
        ENDC = '\033[0m'
        print('\033[1m' + COLOR + text + ENDC + explane)

def icap_client(command):
    subprocess.run(['touch ./testing/output && rm -f ./testing/output'],shell=True)
    proc = subprocess.run([command], stderr=subprocess.PIPE, shell=True)
    output = str(proc.stderr).strip().replace('\\t','').replace("\n","\\n").split("\\n")
    for i in output:
        if i.startswith('ICAP/1.0'):
            statusCode = i[9:12]
            statusMessage = i[13:]
            return statusCode, statusMessage
    return 'No output','No output'

def hashfile(file):
	# A arbitrary (but fixed) buffer
	# size (change accordingly)
	# 65536 = 65536 bytes = 64 kilobytes
	BUF_SIZE = 65536

	# Initializing the sha256() method
	sha256 = hashlib.sha256()

	# Opening the file provided as
	# the first commandline argument
	with open(file, 'rb') as f:
		
		while True:
			
			# reading data = BUF_SIZE from
			# the file and saving it in a variable
			data = f.read(BUF_SIZE)

			# True if eof = 1
			if not data:
				break
	
			# Passing that data to that sh256 hash
			# function (updating the function with that data)
			sha256.update(data)

	
	# sha256.hexdigest() hashes all the input
	# data passed to the sha256() via sha256.update()
	# Acts as a finalize method, after which
	# all the input data gets hashed hexdigest()
	# hashes the data, and returns the output in hexadecimal format
	return sha256.hexdigest()

def Compare_files(file1, file2):
    file1_exist = os.path.exists(file1)
    file2_exist = os.path.exists(file2)
    if (file1_exist and file2_exist):
        f1_hash = hashfile(file1)
        f2_hash = hashfile(file2)
        return f1_hash == f2_hash
    else:
        return False
  
def is_service_exist(testCase, command):
    global passed_tests, failed_tests
    test_statusCode = testCase[1]
    test_statusMessage = testCase[2]
    result_statusCode, result_statusMessage = icap_client(command)
    out = " --> result: " + result_statusCode + " " + result_statusMessage + "; expected: " + test_statusCode + " " + test_statusMessage
    if (test_statusCode == result_statusCode and test_statusMessage == result_statusMessage):
        style.ok("Test passed", out)
        passed_tests = passed_tests + 1
    else : 
        style.fail("Test Failed", out)
        failed_tests += 1

def test_service_name():
    file = open("./testing/service_name.csv")
    csvData = csv.reader(file)
    data = list(csvData)
    # test with 204 
    style.header("***** Test wrong service Name with 204 *****")
    for row in data:
        service = row[0]
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f "./testing/book.pdf" -o ./testing/output -v'

        is_service_exist(row, command)

        


    # test without 204 
    style.header("***** Test wrong service Name without 204 *****")
    for row in data:
        service = row[0]
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f "./testing/book.pdf" -o ./testing/output -v -no204'
        is_service_exist(row, command)


    file.close()

def reconfigure(service, model, value):
    subprocess.run(['cp ./config.toml ./testing'],shell=True)
    data = toml.load("./testing/config.toml") 
    data[service][model] = value 
    f = open("./config.toml",'w')
    toml.dump(data, f)
    f.close()

    # kill icap and restart 
    subprocess.run(['kill -9 $(pidof icapeg)'],shell=True)
    subprocess.run(['./icapeg 2> /dev/null &'],shell=True)
    time.sleep(10)
    
def is_mode_working(test_filename,test_result, command):
    global passed_tests, failed_tests
    subprocess.run(['touch ./testing/output && rm ./testing/output'],shell=True)

    result_statusCode, result_statusMessage = icap_client(command)
    ismatched = Compare_files('./testing/'+test_filename, './testing/output')
    if (ismatched):
        result = "OK"
    else : 
        result = "FAILED"
    if (result == "OK"):
        resultMessage = "File Recieved and status code " + result_statusCode + " " + result_statusMessage
    else:
        resultMessage = "File Not Recieved and status code " + result_statusCode + " " + result_statusMessage

    out = " -->File: " + test_filename +" result: " + resultMessage + " " + "; expected: " + test_result 
    
    if (result == test_result.strip() and result_statusCode + result_statusMessage == "200OK"):
        style.ok("Test passed", out)
        passed_tests = passed_tests + 1
    else : 
        style.fail("Test Failed", out)
        failed_tests += 1

def is_mode_working_with204(test_filename,test_result, command):
    global passed_tests, failed_tests
    result_statusCode, result_statusMessage = icap_client(command)


    resultMessage = "result Header: " + result_statusCode + " " + result_statusMessage + "; expected: " + test_result

    
    if (result_statusCode +" " + result_statusMessage == test_result):
        style.ok("Test passed", resultMessage)
        passed_tests = passed_tests + 1
    else : 
        style.fail("Test Failed", resultMessage)
        failed_tests += 1

def test_mode(mode=''):
    if (mode == "req"):
        options = '-req http://www.example.com'
        modeName = 'Request'
    else:
        options = ''
        modeName = 'Response'
    file = open("./testing/test_size.csv")
    csvData = csv.reader(file)
    data = list(csvData)

    # test with 204 
    style.header("***** Test " + modeName + " mode echo service with 204 *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -v'
         is_mode_working_with204(fileName,'200 OK', command)


    # test without 204 
    style.header("***** Test " + modeName + " mode echo service without 204 *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -v -no204'
        is_mode_working(fileName,'OK', command)

    # test with max file size 
    # subprocess.run(['mv ./testing/config.toml ./config.toml'],shell=True)
    reconfigure("echo", "max_filesize", 100)
    style.header("***** Test " + modeName + " mode echo service with max file size *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        expected = row[1]
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -v -no204'
        is_mode_working(fileName,expected, command)
    subprocess.run(['mv ./testing/config.toml ./config.toml'],shell=True)
    subprocess.run(['kill -9 $(pidof icapeg)'],shell=True)
    subprocess.run(['./icapeg 2> /dev/null &'],shell=True)
    time.sleep(10)

        # test  Without Preview (Server Side) 
    reconfigure("echo", "preview_enabled", False)
    style.header("***** Test " + modeName + " mode echo service Without Preview (Server Side) *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        expected = "OK"
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -v -no204'
        is_mode_working(fileName,expected, command)
    subprocess.run(['mv ./testing/config.toml ./config.toml'],shell=True)
    subprocess.run(['kill -9 $(pidof icapeg)'],shell=True)
    subprocess.run(['./icapeg 2> /dev/null &'],shell=True)
    time.sleep(10)

    # test  Without Preview (client Side) 
    style.header("***** Test " + modeName + " mode echo service Without Preview (client Side) *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        expected = "OK"
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -nopreview -v -no204'
        is_mode_working(fileName,expected, command)


    # test  With Preview 0 (client Side) 
    style.header("***** Test " + modeName + " mode echo service With Preview 0 (client Side) *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        expected = "OK"
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -w 0 -v -no204'
        is_mode_working(fileName,expected, command)

    # test  With preview exceeding limit and file size sent 
    style.header("***** Test " + modeName + " mode echo service With Preview exceeding limit  *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        expected = "OK"
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -w 2048 -v -no204'
        is_mode_working(fileName,expected, command)

    # test  With preview less than file size sent  
    style.header("***** Test " + modeName + " mode echo service With Preview exceeding file size sent  *****")
    for row in data:
        service = 'echo'
        fileName = row[0]
        expected = "OK"
        inputfile = './testing/' + fileName
        command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -w 100 -v -no204'
        is_mode_working(fileName,expected, command)

    # test  With request methods GET POST PUT CONNECT  
    style.header("***** Test " + modeName + " mode echo service With HTTP Methods *****")
    # methods = ['GET', 'POST', 'PUT', 'CONNECT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS', 'TRACE', 'FackeMehod'] 
    methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS', 'TRACE', 'FackeMehod'] 
    for method in methods:
        style.yellow("Method: " + method)
        for row in data:
            service = 'echo'
            fileName = row[0]
            expected = "OK"
            inputfile = './testing/' + fileName
            command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output '+ options +' -method '+ method +' -w 100 -v -no204'
            is_mode_working(fileName,expected, command)

def icap_client_istag(command):
    subprocess.run(['touch ./testing/output && rm -f ./testing/output'],shell=True)
    proc = subprocess.run([command], stderr=subprocess.PIPE, shell=True)
    output = str(proc.stderr).strip().replace('\\t','').replace("\n","\\n").split("\\n")
    for i in output:
        if i.startswith('ISTag'):
            IStag = i[7:]
            return IStag
    return 'No output'

def get_IStag(service):
    data = toml.load("./config.toml") 
    IStag = data[service]['service_tag'] 
    return IStag

def get_icap_services():
    data = toml.load("./config.toml") 
    services = data['app']['services'] 
    return services        

def is_tags_matched(tag1,tag2):
    global passed_tests, failed_tests
    out = " --> result: " + tag1 + "; expected: " + tag2
    if (tag1 == tag2):
        style.ok("Test passed", out)
        passed_tests = passed_tests + 1
    else : 
        style.fail("Test Failed", out)
        failed_tests += 1

def test_istag():
    command = ''
    file = open("./testing/test_size.csv")
    csvData = csv.reader(file)
    data = list(csvData)

    services = get_icap_services()

    style.header("***** Test IStag *****")
    for service in services:
        tag = get_IStag(service)


        # test with 204 
        style.yellow(service + ":  with 204")
        for row in data:
            fileName = row[0]
            inputfile = './testing/' + fileName
            command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output  -v'
            IStag = icap_client_istag(command)
            is_tags_matched(IStag, tag)

        # test without  204 
        style.yellow(service + ":  with no204")
        for row in data:
            fileName = row[0]
            inputfile = './testing/' + fileName
            command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output  -no204 -v'
            IStag = icap_client_istag(command)
            is_tags_matched(IStag, tag)

        # test with preview 0 
        style.yellow(service + ":  with preview 0")
        for row in data:
            fileName = row[0]
            inputfile = './testing/' + fileName
            command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output  -w 0 -v'
            IStag = icap_client_istag(command)
            is_tags_matched(IStag, tag)

        # test with preview under and apove file size 
        style.yellow(service + ":  with under and apove file size")
        for row in data:
            fileName = row[0]
            inputfile = './testing/' + fileName
            command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output  -w 110 -v'
            IStag = icap_client_istag(command)
            is_tags_matched(IStag, tag)

        # test with no preview (client side)
        style.yellow(service + ":  with no preview (client side)")
        for row in data:
            fileName = row[0]
            inputfile = './testing/' + fileName
            command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output  -nopreview -v'
            IStag = icap_client_istag(command)
            is_tags_matched(IStag, tag)

        # test with no preview (server side)
        reconfigure("echo", "preview_enabled", False)
        style.yellow(service + ":  with no preview (server side)")
        for row in data:
            fileName = row[0]
            inputfile = './testing/' + fileName
            command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output -v'
            IStag = icap_client_istag(command)
            is_tags_matched(IStag, tag)
        subprocess.run(['mv ./testing/config.toml ./config.toml'],shell=True)
        subprocess.run(['kill -9 $(pidof icapeg)'],shell=True)
        subprocess.run(['./icapeg 2> /dev/null &'],shell=True)
        time.sleep(10)


        # test  With request methods GET POST PUT CONNECT  
        methods = ['GET', 'POST', 'PUT', 'CONNECT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS', 'TRACE', 'FackeMehod']
        for method in methods:
            style.yellow("HTTP Method: " + method)
            for row in data:
                fileName = row[0]
                expected = "OK"
                inputfile = './testing/' + fileName
                command = 'c-icap-client -i 127.0.0.1  -p 1344 -s '+ service + ' -f '+ inputfile +' -o ./testing/output ' + ' -method '+ method +' -v'
                is_tags_matched(IStag, tag)

#--------------------------------------



# -------------------------------------
# add txt extensions to process_extensions
reconfigure("echo", "bypass_extensions", ["pdf"])
reconfigure("echo", "process_extensions", ["*"])
time.sleep(10)
# test_service_name()
test_mode('resp')
test_mode('req')
# test_istag()

# =====================================

# subprocess.run(['mv ./testing/config.toml ./config.toml'],shell=True)
subprocess.run(['kill -9 $(pidof icapeg)'],shell=True)
Total = passed_tests + failed_tests 
print('\n\033[1m \033[1;32;46m ######### conclusion ######## \033[0m')
print('\033[93m Total: \033[0m ' + str(Total))
style.ok("Passed: ", str(passed_tests))
style.fail("Failed: ", str(failed_tests))

if(failed_tests > 0 ):
    sys.exit(50)
else : 
    sys.exit()
