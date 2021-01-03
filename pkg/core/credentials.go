package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func (t *InitClientConfig) InitClient(instance string) (endpoint string, clientResponse ClientResponse, passwd *memguard.Enclave, err error) {
	rbody, err := ioutil.ReadFile(t.ConfDir + "/" + instance + ".json")

	switch {

	case err != nil:
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
			log.Debug().Str("IAM endpoint used", t.IAMServer).Msg("Credentials")
			color.Green.Printf("==> IAM endpoint used: %s\n", t.IAMServer)
			endpoint = t.IAMServer
		}

		register := endpoint + "/register"

		log.Debug().Str("IAM register url", register).Msg("Credentials")
		color.Green.Printf("==> IAM register url: %s\n", register)

		resp, err := t.HTTPClient.Post(register, contentType, strings.NewReader(request))
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		log.Debug().Int("StatusCode", resp.StatusCode).Str("Status", resp.Status).Msg("credentials")

		rbody, errReadBody := ioutil.ReadAll(resp.Body)
		if errReadBody != nil {
			panic(errReadBody)
		}

		log.Debug().Str("body", string(rbody)).Msg("credentials")

		errUnmarshall := json.Unmarshal(rbody, &clientResponse)
		if errUnmarshall != nil {
			panic(errUnmarshall)
		}

		clientResponse.Endpoint = endpoint

		if !t.NoPWD { //nolint:nestif
			var errGetPasswd error

			// TODO: verify branch when REFRESH_TOKEN is passed and is not empty string
			if os.Getenv("REFRESH_TOKEN") == "" {
				passMsg := fmt.Sprintf("%s Insert a pasword for the secret's encryption: ", color.Yellow.Sprint("==>"))
				passwd, errGetPasswd = t.Scanner.GetPassword(passMsg, false)

				if errGetPasswd != nil {
					panic(errGetPasswd)
				}
			} else {
				passwd = memguard.NewEnclave([]byte("nopassword"))
			}

			dumpClient := Encrypt(rbody, passwd)

			err = ioutil.WriteFile(t.ConfDir+"/"+instance+".json", dumpClient, 0600)
			if err != nil {
				panic(err)
			}
		}
	case err == nil && !t.NoPWD:
		var errGetPasswd error

		// TODO: verify branch when REFRESH_TOKEN is passed and is not empty string
		if os.Getenv("REFRESH_TOKEN") == "" {
			passMsg := fmt.Sprintf("%s Insert a pasword for the secret's decryption: ", color.Yellow.Sprint("==>"))
			passwd, errGetPasswd = t.Scanner.GetPassword(passMsg, true)

			if errGetPasswd != nil {
				panic(errGetPasswd)
			}
		} else {
			passwd = memguard.NewEnclave([]byte("nopassword"))
		}

		errUnmarshal := json.Unmarshal(Decrypt(rbody, passwd), &clientResponse)
		if errUnmarshal != nil {
			panic(errUnmarshal)
		}

		log.Debug().Str("response endpoint", clientResponse.Endpoint).Msg("credentials")
		endpoint = strings.Split(clientResponse.Endpoint, "/register")[0]
	}

	if endpoint == "" {
		panic("Something went wrong. No endpoint selected")
	}

	return endpoint, clientResponse, passwd, nil
}
