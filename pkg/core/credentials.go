package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/awnumar/memguard"
	"github.com/rs/zerolog/log"

	"github.com/gookit/color"
)

type InitClientConfig struct {
	ConfDir        string
	ClientConfig   IAMClientConfig
	Scanner        GetInputWrapper
	HTTPClient     http.Client
	IAMServer      string
	ClientTemplate string
	NoPWD          bool
}

type WellKnown struct {
	RegisterEndpoint string `json:"registration_endpoint"`
}

func GetRegisterEndpoint(endpoint string) (register_endpoint string){
	var c http.Client
	well_known := endpoint + "/.well-known/openid-configuration"
	resp, err := c.Get(well_known)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var wk WellKnown

	errUnmarshall := json.Unmarshal(body, &wk)

	if errUnmarshall != nil {
		panic(errUnmarshall)
	}

	return string(wk.RegisterEndpoint)
}

func (t *InitClientConfig) InitClient(instance string) (endpoint string, clientResponse ClientResponse, passwd *memguard.Enclave, err error) { //nolint:funlen,cyclop,gocognit,lll
	instanceConfFilename := t.ConfDir + "/" + instance + ".json"

	log.Debug().Str("filename", instanceConfFilename).Msg("credentials - init client")

	confFile, err := os.Open(instanceConfFilename)

	switch {
	case err != nil && err.Error() != "no such file or directory":
		tmpl, errParser := template.New("client").Parse(t.ClientTemplate)
		if errParser != nil {
			panic(errParser)
		}

		var b bytes.Buffer
		errExecute := tmpl.Execute(&b, t.ClientConfig)

		if errExecute != nil {
			panic(errExecute)
		}

		request := b.String()

		log.Debug().Str("URL", request).Msg("credentials")

		contentType := "application/json"

		log.Debug().Str("REFRESH_TOKEN", os.Getenv("REFRESH_TOKEN")).Msg("credentials")

		if t.IAMServer == "" {
			endpoint, err = t.Scanner.GetInputString("Insert the IAM endpoint",
				"https://iam-demo.cloud.cnaf.infn.it")
			if err != nil {
				panic(err)
			}
		} else if t.IAMServer != "" {
			log.Debug().Str("IAM endpoint used", t.IAMServer).Msg("credentials")
			color.Green.Printf("==> IAM endpoint used: %s\n", t.IAMServer)
			endpoint = t.IAMServer
		}

		register := GetRegisterEndpoint(endpoint)

		log.Debug().Str("IAM register url", register).Msg("credentials")
		color.Green.Printf("==> IAM register url: %s\n", register)

		resp, err := t.HTTPClient.Post(register, contentType, strings.NewReader(request))
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		log.Debug().Int("StatusCode", resp.StatusCode).Str("Status", resp.Status).Msg("credentials")

		var rbody bytes.Buffer

		_, err = rbody.ReadFrom(resp.Body)
		if err != nil {
			log.Err(err).Msg("credentials - read body")
			panic(err)
		}

		log.Debug().Str("body", rbody.String()).Msg("credentials")

		errUnmarshall := json.Unmarshal(rbody.Bytes(), &clientResponse)
		if errUnmarshall != nil {
			panic(errUnmarshall)
		}

		clientResponse.Endpoint = endpoint

		if !t.NoPWD { //nolint:nestif
			var errGetPasswd error

			// TODO: verify branch when REFRESH_TOKEN is passed and is not empty string
			if os.Getenv("REFRESH_TOKEN") == "" {
				passMsg := fmt.Sprintf("%s Insert a password for the secret's encryption: ", color.Yellow.Sprint("==>"))
				passwd, errGetPasswd = t.Scanner.GetPassword(passMsg, false)

				if errGetPasswd != nil {
					panic(errGetPasswd)
				}
			} else {
				passwd = memguard.NewEnclave([]byte("nopassword"))
			}

			dumpClient := Encrypt(rbody.Bytes(), passwd)

			curFile, err := os.OpenFile(instanceConfFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				log.Err(err).Msg("credentials - dump client")

				panic(err)
			}

			_, err = curFile.Write(dumpClient)
			if err != nil {
				log.Err(err).Msg("credentials - dump client")

				panic(err)
			}

			err = curFile.Close()
			if err != nil {
				log.Err(err).Msg("credentials - dump client")

				panic(err)
			}
		}
	case err == nil && !t.NoPWD:
		var errGetPasswd error

		var rbody bytes.Buffer

		defer confFile.Close()

		_, err = rbody.ReadFrom(confFile)
		if err != nil {
			log.Err(err).Msg("credentials - init client")
			panic(err)
		}

		// TODO: verify branch when REFRESH_TOKEN is passed and is not empty string
		if os.Getenv("REFRESH_TOKEN") == "" {
			passMsg := fmt.Sprintf("%s Insert a password for the secret's decryption: ", color.Yellow.Sprint("==>"))

			passwd, errGetPasswd = t.Scanner.GetPassword(passMsg, true)

			if errGetPasswd != nil {
				panic(errGetPasswd)
			}
		} else {
			passwd = memguard.NewEnclave([]byte("nopassword"))
		}

		errUnmarshal := json.Unmarshal(Decrypt(rbody.Bytes(), passwd), &clientResponse)
		if errUnmarshal != nil {
			panic(errUnmarshal)
		}

		log.Debug().Str("response endpoint", clientResponse.Endpoint).Msg("credentials")
		endpoint = strings.Split(clientResponse.Endpoint, "/register")[0]
	default:
		log.Err(err).Msg("credentials - init client")
		panic(err)
	}

	if endpoint == "" {
		panic("Something went wrong. No endpoint selected")
	}

	return endpoint, clientResponse, passwd, nil
}
