import requests
import threading

BASE_URL = "http://localhost:8080/api/generate"
TOTAL_REQUESTS = 10000

def send_request(job_name):
    headers = {
        "Content-Type": "application/json"
    }

    data = {
        "jobName": job_name,
        "jobType": 1,  # Assuming you want the 'Recur' option
        "cronExpr": "* * * * *",
        "format": 1,  # Python
        "script": '''
import time
start_time = time.time()
total = 0
for i in range(1, 10000001):
    total += i
end_time = time.time()
print("The sum of 1 to 10000000 is:", total)
print("Execution time:", end_time - start_time, "seconds")
''',
        "retries": 3
    }

    response = requests.post(BASE_URL, json=data, headers=headers)
    if response.status_code != 200:
        print(f"Failed for job: {job_name}")
        print("Server response:", response.text)

def main():
    threads = []
    for i in range(TOTAL_REQUESTS):
        job_name = f"New-Job{i+1}"
        t = threading.Thread(target=send_request, args=(job_name,))
        threads.append(t)
        t.start()

    for t in threads:
        t.join()

    print(f"{TOTAL_REQUESTS} requests sent!")

if __name__ == "__main__":
    main()