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

func (t *InitClientConfig) InitClient(instance string) (endpoint string, clientResponse ClientResponse, passwd string, err error) {

	rbody, err := ioutil.ReadFile(t.ConfDir + "/" + instance + ".json")
	if err != nil {

		tmpl, err := template.New("client").Parse(t.ClientTemplate)
		if err != nil {
			panic(err)
		}

		var b bytes.Buffer
		err = tmpl.Execute(&b, t.ClientConfig)
		if err != nil {
			panic(err)
		}

		request := b.String()

		fmt.Println(request)

		contentType := "application/json"
		//body := url.Values{}
		//body. Set(request)

		fmt.Println(os.Getenv("REFRESH_TOKEN"))

		if t.IAMServer == "" {

			endpoint, err = t.Scanner.GetInputString("Insert the IAM endpoint for the instance: ", "https://iam-demo.cloud.cnaf.infn.it")
			if err != nil {
				panic(err)
			}
		} else if t.IAMServer != "" {
			log.Info().Str("IAM endpoint used", t.IAMServer).Msg("Credentials")
			color.Green.Printf("==> IAM endpoint used: %s", t.IAMServer)
			endpoint = t.IAMServer
		}

		register := endpoint + "/register"

		log.Info().Str("IAM register url", register).Msg("Credentials")
		color.Green.Printf("==> IAM register url: %s", register)

		r, err := t.HTTPClient.Post(register, contentType, strings.NewReader(request))
		if err != nil {
			panic(err)
		}

		//fmt.Println(r.StatusCode, r.Status)

		rbody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		//fmt.Println(string(rbody))

		err = json.Unmarshal(rbody, &clientResponse)
		if err != nil {
			panic(err)
		}

		clientResponse.Endpoint = endpoint

		if !t.NoPWD {
			passwd := ""

			if os.Getenv("REFRESH_TOKEN") == "" {

				passMsg := fmt.Sprintf("\n%s Insert a pasword for the secret's encryption: ", color.Yellow.Sprint("==>"))
				passwd, err = t.Scanner.GetPassword(passMsg, false)

				if err != nil {
					panic(err)
				}
			} else {
				passwd = "asdasdasd"
			}

			dumpClient := Encrypt(rbody, passwd)

			err = ioutil.WriteFile(t.ConfDir+"/"+instance+".json", dumpClient, 0600)
			if err != nil {
				panic(err)
			}
		}
	} else {
		if !t.NoPWD {
			passwd := ""

			if os.Getenv("REFRESH_TOKEN") == "" {
				passMsg := fmt.Sprintf("\n%s Insert a pasword for the secret's decryption: ", color.Yellow.Sprint("==>"))
				passwd, err = t.Scanner.GetPassword(passMsg, true)

				if err != nil {
					panic(err)
				}
			} else {
				passwd = "asdasdasd"
			}

			err = json.Unmarshal(Decrypt(rbody, passwd), &clientResponse)
			if err != nil {
				panic(err)
			}
			fmt.Println(clientResponse.Endpoint)
			endpoint = strings.Split(clientResponse.Endpoint, "/register")[0]
		}
	}

	if endpoint == "" {
		panic("Something went wrong. No endpoint selected")
	}

	return endpoint, clientResponse, passwd, nil
}
