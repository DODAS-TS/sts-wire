package template

// ClientTemplate ..
const ClientTemplate = `{
	"redirect_uris": [
	  "http://{{ .Host }}:{{ .Port }}/oauth2/callback"
	],
	"client_name": "{{ .ClientName }}",
	"contacts": [
	  "client@iam.test"
	],
	"token_endpoint_auth_method": "client_secret_basic",
	"scope": "address phone openid email profile offline_access",
	"grant_types": [
	  "refresh_token",
	  "authorization_code"
	],
	"response_types": [
	  "code"
	]
  }`

// RCloneTemplate ..
const RCloneTemplate = `
[{{ .Instance }}]
type = s3
provider = Minio
env_auth = false
access_key_id =
secret_access_key =
session_token =
endpoint = {{ .Address }}`
