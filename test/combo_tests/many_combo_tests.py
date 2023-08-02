import requests
import threading
import itertools

BASE_URL = "http://localhost:8080/api/generate"

JOB_NAMES = [f"Job{i}" for i in range(1, 11)]  # 10 different job names for example
JOB_TYPES = [0, 1]  # 0 for one-time, 1 for recurring
CRON_EXPRS = ["* * * * *", "0 * * * *", "0 0 * * *"]  # Every minute, every hour, every day
FORMATS = {
    0: ["echo 'Hello, World!'", "ls -l"],  # Shell
    1: ["print('Hello from Python!')", "print(sum(range(1, 10001)))"]  # Python
}
RETRIES = [0, 1, 2, 3]

def send_request(data):
    headers = {
        "Content-Type": "application/json"
    }

    response = requests.post(BASE_URL, json=data, headers=headers)
    if response.status_code != 200:
        print(f"Failed for job: {data['jobName']}")
        print("Server response:", response.text)

def main():
    threads = []

    # For each format, we generate combinations using scripts relevant to that format.
    for fmt, scripts_for_format in FORMATS.items():
        combinations = itertools.product(JOB_NAMES, JOB_TYPES, CRON_EXPRS, [fmt], scripts_for_format, RETRIES)
        for combo in combinations:
            data = {
                "jobName": combo[0],
                "jobType": combo[1],
                "cronExpr": combo[2],
                "format": combo[3],
                "script": combo[4],
                "retries": combo[5]
            }

            t = threading.Thread(target=send_request, args=(data,))
            threads.append(t)
            t.start()

    for t in threads:
        t.join()

    print(f"All valid combinations sent!")

if __name__ == "__main__":
    main()