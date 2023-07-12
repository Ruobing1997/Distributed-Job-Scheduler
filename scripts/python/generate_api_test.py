import requests


if __name__ == '__main__':
    url = "http://localhost:8080/generate"

    data = {
        "name": "Test Task",
        "jobType": "OneTime",
        "schedule": "0 0 * * *",
        "payload": "Sample payload",
    }

    response = requests.post(url, json=data)

    if response.status_code == 200:
        print("API request was successful!")
        print("Response data:", response.json())
    else:
        print("API request failed with status code:", response.status_code)
        print("Error message:", response.text)
