mod auth 'auth/justfile'

# Generate auth protobuf go files from auth_svc.proto template
gen-auth-proto:
  just auth generate-proto
