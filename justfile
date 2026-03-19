mod auth 'auth/justfile'
mod money_movement 'money_movement/justfile'

# Generate auth protobuf go files from auth_svc.proto template
gen-auth-proto:
  just auth generate-proto

# Generate money_movement protobuf go files from money_movement_svc.proto template
gen-mm-proto:
  just money_movement generate-proto
