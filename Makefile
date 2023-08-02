.PHONY: clean-service clean-deployments get-pods

clean-service:
	kubectl delete service --all -n supernova

clean-deployments:
	kubectl delete deployments --all -n supernova

get-pods:
	kubectl get pods -n supernova