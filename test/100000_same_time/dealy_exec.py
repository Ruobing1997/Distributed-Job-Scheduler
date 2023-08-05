import requests
import threading
from datetime import datetime, timedelta

BASE_URL = "http://localhost:8080/api/generate"
TOTAL_REQUESTS = 10
DELAY_IN_MINUTES = 1  # You can adjust this value to delay the task by different minutes

def send_request(job_name):
    headers = {
        "Content-Type": "application/json"
    }

    # Calculate cron expression for DELAY_IN_MINUTES later
    future_time = datetime.now() + timedelta(minutes=DELAY_IN_MINUTES)
    cronExpr = f"{future_time.minute} {future_time.hour} * * *"

    data = {
        "jobName": job_name,
        "jobType": 0,  # Assuming you want the 'Recur' option
        "cronExpr": cronExpr,  # This will execute the job DELAY_IN_MINUTES minutes from now
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