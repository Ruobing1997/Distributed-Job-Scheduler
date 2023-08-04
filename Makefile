.PHONY: clean-service clean-deployments get-pods clean-all

clean-service:
	kubectl delete service --all -n supernova

clean-deployments:
	kubectl delete deployments --all -n supernova

clean-daemonset:
	kubectl delete daemonset --all -n supernova

clean:
	kubectl delete service --all -n supernova
	kubectl delete deployments --all -n supernova
	kubectl delete daemonset --all -n supernova

get-pods:
	kubectl get pods -n supernova