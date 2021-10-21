package core

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/minio/minio-go/v6/pkg/credentials"
	"github.com/minio/minio/pkg/auth"
	"github.com/rs/zerolog/log"
)

type RefreshTokenStruct struct {
	RefreshToken     string `json:"refresh_token"`
	AccessToken      string `json:"access_token"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
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

// IAMProvider credential provider for oidc.
type IAMProvider struct {
	StsEndpoint       string
	HTTPClient        *http.Client
	Token             string
	Creds             *AssumeRoleWithWebIdentityResponse
	RefreshTokenRenew int
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
func (t *IAMProvider) Retrieve() (credentials.Value, error) { // nolint:funlen
	log.Debug().Int("RefreshTokenRenew",
		t.RefreshTokenRenew).Str("RefreshTokenRenew string",
		strconv.Itoa(t.RefreshTokenRenew*60)).Msg("IAM - Retrieve")

	body := url.Values{}
	body.Set("Action", "AssumeRoleWithWebIdentity")
	body.Set("Version", "2011-06-15")
	body.Set("WebIdentityToken", t.Token)
	body.Set("DurationSeconds", strconv.Itoa(t.RefreshTokenRenew*60))

	log.Debug().Str("stsEndpoint", t.StsEndpoint).Str("body", body.Encode()).Msg("IAM")

	url, errParse := url.Parse(
		strings.Join(
			[]string{t.StsEndpoint, "?", body.Encode()},
			"",
		),
	)
	if errParse != nil {
		panic(errParse)
	}

	log.Debug().Str("url", url.String()).Msg("IAM")

	req := http.Request{ // nolint:exhaustivestruct
		Method: "POST",
		URL:    url,
	}

	resp, errDo := t.HTTPClient.Do(&req)
	if errDo != nil {
		log.Err(errDo).Msg("IAM connect client")

		if strings.Contains(errDo.Error(), "connection refused") {
			color.Red.Println("IAM client connection")
			color.Red.Println(fmt.Sprintf("==> Cannot connect to '%s'", url))
			color.Red.Println("==> Verify your IAM client")

			panic("IAM client connection")
		}

		return credentials.Value{}, fmt.Errorf("IAM retrieve %w", errDo)
	}
	defer resp.Body.Close()

	log.Debug().Int("statusCode", resp.StatusCode).Str("status", resp.Status).Msg("IAM")

	var rbody bytes.Buffer

	_, errRead := rbody.ReadFrom(resp.Body)
	if errRead != nil {
		log.Err(errRead).Msg("IAM read body")

		return credentials.Value{}, fmt.Errorf("IAM retrieve %w", errRead)
	}

	log.Debug().Str("body", rbody.String()).Msg("IAM")

	t.Creds = &AssumeRoleWithWebIdentityResponse{}

	errUnmarshall := xml.Unmarshal(rbody.Bytes(), t.Creds)
	if errUnmarshall != nil {
		log.Err(errUnmarshall).Msg("IAM xml unmarshal")

		return credentials.Value{}, fmt.Errorf("IAM retrieve %w", errUnmarshall)
	}

	log.Debug().Str("credentials", "acquired").Msg("IAM")

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
