package core

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v6/pkg/credentials"
	"github.com/minio/minio/pkg/auth"
	"github.com/rs/zerolog/log"
)

type RefreshTokenStruct struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type IAMClientConfig struct {
	CallbackURL string
	Host        string
	Port        int
	ClientName  string
}

type ClientResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Endpoint     string `json:"registration_client_uri"`
}

// Returns a base64 encoded random 32 byte string.
func RandomState() string {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

// IAMProvider credential provider for oidc
type IAMProvider struct {
	StsEndpoint string
	HTTPClient  *http.Client
	Token       string
	Creds       *AssumeRoleWithWebIdentityResponse
}

// AssumeRoleWithWebIdentityResponse the struct of the STS WebIdentity call response.
type AssumeRoleWithWebIdentityResponse struct {
	XMLName          xml.Name          `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleWithWebIdentityResponse" json:"-"`
	Result           WebIdentityResult `xml:"AssumeRoleWithWebIdentityResult"`
	ResponseMetadata struct {
		RequestID string `xml:"RequestId,omitempty"`
	} `xml:"ResponseMetadata,omitempty"`
}

// AssumedRoleUser - The identifiers for the temporary security credentials that
// the operation returns. Please also see https://docs.aws.amazon.com/goto/WebAPI/sts-2011-06-15/AssumedRoleUser
type AssumedRoleUser struct {
	Arn           string
	AssumedRoleID string `xml:"AssumeRoleId"`
	// contains filtered or unexported fields
}

// WebIdentityResult - Contains the response to a successful AssumeRoleWithWebIdentity
// request, including temporary credentials that can be used to make MinIO API requests.
type WebIdentityResult struct {
	AssumedRoleUser             AssumedRoleUser  `xml:",omitempty"`
	Audience                    string           `xml:",omitempty"`
	Credentials                 auth.Credentials `xml:",omitempty"`
	PackedPolicySize            int              `xml:",omitempty"`
	Provider                    string           `xml:",omitempty"`
	SubjectFromWebIdentityToken string           `xml:",omitempty"`
}

// Retrieve credentials.
func (t *IAMProvider) Retrieve() (credentials.Value, error) {
	log.Info().Msg("IAM - Retrieve")

	body := url.Values{}
	body.Set("Action", "AssumeRoleWithWebIdentity")
	body.Set("Version", "2011-06-15")
	body.Set("WebIdentityToken", t.Token)

	// TODO: parameter for duration
	body.Set("DurationSeconds", "900")

	log.Info().Str("stsEndpoint", t.StsEndpoint).Str("body", body.Encode()).Msg("IAM")

	url, err := url.Parse(
		strings.Join(
			[]string{t.StsEndpoint, "?", body.Encode()},
			"",
		),
	)
	if err != nil {
		panic(err)
	}

	log.Info().Str("url", url.String()).Msg("IAM")
	req := http.Request{ // nolint:exhaustivestruct
		Method: "POST",
		URL:    url,
	}

	// TODO: retrieve token with https POST with t.httpClient
	r, errDo := t.HTTPClient.Do(&req)
	if errDo != nil {
		log.Err(errDo).Msg("IAM")

		return credentials.Value{}, fmt.Errorf("IAM retrieve %w", errDo)
	}
	// defer r.Body.Close()
	fmt.Println(r.StatusCode, r.Status)

	rbody, errRead := ioutil.ReadAll(r.Body)
	if errRead != nil {
		log.Err(errRead).Msg("IAM")

		return credentials.Value{}, fmt.Errorf("IAM retrieve %w", errRead)
	}

	log.Info().Str("body", string(rbody)).Msg("IAM")

	t.Creds = &AssumeRoleWithWebIdentityResponse{}

	errUnmarshall := xml.Unmarshal(rbody, t.Creds)
	if errUnmarshall != nil {
		log.Err(errUnmarshall).Msg("IAM")

		return credentials.Value{}, fmt.Errorf("IAM retrieve %w", errUnmarshall)
	}

	log.Info().Str("credential", "acquired").Msg("IAM")

	return credentials.Value{ // nolint:exhaustivestruct
		AccessKeyID:     t.Creds.Result.Credentials.AccessKey,
		SecretAccessKey: t.Creds.Result.Credentials.SecretKey,
		SessionToken:    t.Creds.Result.Credentials.SessionToken,
	}, nil

}

// IsExpired test.
func (t *IAMProvider) IsExpired() bool {
	return t.Creds.Result.Credentials.IsExpired()
}
