mod api_gateway 'api_gateway/justfile'
mod auth 'auth/justfile'
mod money_movement 'money_movement/justfile'

# Generate auth protobuf go files from auth_svc.proto template
gen-auth-proto:
  just auth generate-proto

# Generate money_movement protobuf go files from money_movement_svc.proto template
gen-mm-proto:
  just money_movement generate-proto

# Generate API Gateway Swagger docs
gen-api-gateway-docs:
  just api_gateway gen-swagger-docs

# Start the services
# Ensure to do `minikube start`
# and in a separate tab: `minikube mount .:/go-payment-micro`
# Setup kafka queue using Strimzi: https://strimzi.io/quickstarts/
#
# After the services are up, run `minikube tunnel` in a separate tab
# to access the api service.
start-services:
  kubectl apply -R -f mysql_auth/manifests -n kafka
  kubectl apply -R -f mysql_money_movement/manifests -n kafka
  kubectl apply -R -f mysql_ledger/manifests -n kafka
  kubectl apply -R -f auth/manifests -n kafka
  kubectl apply -R -f money_movement/manifests -n kafka
  kubectl apply -R -f ledger/manifests -n kafka
  kubectl apply -R -f email/manifests -n kafka
  kubectl apply -R -f api_gateway/manifests -n kafka

stop-services:
  kubectl delete -R -f api_gateway/manifests -n kafka
  kubectl delete -R -f email/manifests -n kafka
  kubectl delete -R -f ledger/manifests -n kafka
  kubectl delete -R -f money_movement/manifests -n kafka
  kubectl delete -R -f auth/manifests -n kafka
  kubectl delete -R -f mysql_ledger/manifests -n kafka
  kubectl delete -R -f mysql_money_movement/manifests -n kafka
  kubectl delete -R -f mysql_auth/manifests -n kafka
