package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"text/template"
	"time"

	iamTmpl "github.com/DODAS-TS/sts-wire/pkg/template"
	"github.com/gookit/color"
	"github.com/minio/minio-go/v6/pkg/credentials"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"

	"github.com/rs/zerolog/log"
)

// RCloneStruct ..
type RCloneStruct struct {
	Address  string
	Instance string
}

// IAMCreds ..
type IAMCreds struct {
	AccessToken  string
	RefreshToken string
}

// Server ..
type Server struct {
	Client            InitClientConfig
	Instance          string
	S3Endpoint        string
	RemotePath        string
	LocalPath         string
	Endpoint          string
	Response          ClientResponse
	RefreshTokenRenew int
	rcloneCmd         *exec.Cmd
}

// Start ..
func (s *Server) Start() (ClientResponse, IAMCreds, string, error) { //nolint: funlen, gocognit
	state := RandomState()
	isByHand := os.Getenv("REFRESH_TOKEN")
	credsIAM := IAMCreds{}
	endpoint := s.Endpoint
	clientResponse := s.Response

	if isByHand == "" { //nolint:nestif
		endpoint := s.Endpoint
		clientResponse := s.Response

		sigint := make(chan int, 1)

		//fmt.Println(clientResponse.ClientID)
		//fmt.Println(clientResponse.ClientSecret)

		ctx := context.Background()

		config := oauth2.Config{
			ClientID:     clientResponse.ClientID,
			ClientSecret: clientResponse.ClientSecret,
			Endpoint: oauth2.Endpoint{ // nolint:exhaustivestruct
				AuthURL:  endpoint + "/authorize",
				TokenURL: endpoint + "/token",
			},
			RedirectURL: fmt.Sprintf("http://localhost:%d/oauth2/callback", s.Client.ClientConfig.Port),
			Scopes:      []string{"address", "phone", "openid", "email", "profile", "offline_access"},
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			//log.Printf("%s %s", r.Method, r.RequestURI)
			if r.RequestURI != "/" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
		})

		http.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
			//log.Printf("%s %s", r.Method, r.RequestURI)
			if r.URL.Query().Get("state") != state {
				http.Error(w, "state did not match", http.StatusBadRequest)
				return
			}

			oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
			if err != nil {
				http.Error(w, "cannot get token", http.StatusBadRequest)
				return
			}
			if !oauth2Token.Valid() {
				http.Error(w, "token expired", http.StatusBadRequest)
				return
			}

			token := oauth2Token.Extra("access_token").(string)

			credsIAM.AccessToken = token
			credsIAM.RefreshToken = oauth2Token.Extra("refresh_token").(string)

			err = ioutil.WriteFile(".token", []byte(token), 0600)
			if err != nil {
				log.Err(fmt.Errorf("Could not save token file: %s", err)).Msg("server")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			//fmt.Println(token)

			//sts, err := credentials.NewSTSWebIdentity("https://131.154.97.121:9001/", getWebTokenExpiry)
			providers := []credentials.Provider{
				&IAMProvider{ //nolint: exhaustivestruct
					StsEndpoint: s.S3Endpoint,
					Token:       token,
					HTTPClient:  &s.Client.HTTPClient,
				},
			}

			sts := credentials.NewChainCredentials(providers)
			if err != nil {
				log.Err(fmt.Errorf("Could not set STS credentials: %s", err)).Msg("server")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			creds, err := sts.Get()
			if err != nil {
				log.Err(fmt.Errorf("Could not get STS credentials: %s", err)).Msg("server")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			//fmt.Println(creds)

			response := make(map[string]interface{})
			response["credentials"] = creds
			_, err = json.MarshalIndent(response, "", "\t")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//msg := fmt.Sprintf("CREDENTIALS %s", creds)
			//w.Write([]byte(msg))
			html, errAsset := Asset("data/html/mountingPage.html")
			if errAsset != nil {
				http.Error(w, errAsset.Error(), http.StatusInternalServerError)
				return
			}

			_, errWrite := w.Write(html)
			if errWrite != nil {
				panic(errWrite)
			}

			sigint <- 1

		})

		address := "localhost:3128"
		urlBrowse := fmt.Sprintf("http://%s/", address)

		log.Info().Str("IAM auth URL", urlBrowse).Msg("Server")

		err := browser.OpenURL(urlBrowse)
		if err != nil {
			log.Err(err).Msg("Failed to open browser, trying to copy the following on you browser")
			log.Info().Msg(config.AuthCodeURL(state))
			log.Info().Msg("After that copy the resulting address and run the following command on a separate shell")
			log.Info().Msg("curl \"<your resulting address e.g. http://localhost:3128/oauth2/callback?code=1tpAd&state=9RpeJxIf>\" ")

			color.Red.Println("!!! Failed to open browser, trying to copy the following on you browser")
			fmt.Printf("==> %s\n", config.AuthCodeURL(state))
			color.Yellow.Println("=> After that copy the resulting address and run the following command on a separate shell")
			color.Yellow.Println("-> curl \"<your resulting address e.g. http://localhost:3128/oauth2/callback?code=1tpAd&state=9RpeJxIf>\"")
		}

		srv := &http.Server{Addr: address} // nolint: exhaustivestruct

		idleConnsClosed := make(chan struct{})
		closeConn := func() {
			<-sigint

			// We received an interrupt signal, shut down.
			if err := srv.Shutdown(context.Background()); err != nil {
				// Error from closing listeners, or context timeout:
				log.Err(err).Msg("server")
			}

			close(idleConnsClosed)
		}

		go closeConn()

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// Error starting or closing listener:
			fmt.Println(err)
			fmt.Println(errors.Is(err, http.ErrServerClosed))
			log.Err(err).Msg("server")
		}

		<-idleConnsClosed
	} else {
		token := os.Getenv("ACCESS_TOKEN")

		credsIAM.AccessToken = token
		credsIAM.RefreshToken = os.Getenv("REFRESH_TOKEN")

		fmt.Printf("Writing down token: %s", token)
		err := ioutil.WriteFile(".token", []byte(token), 0600)
		if err != nil {
			log.Err(fmt.Errorf("Could not save token file: %s", err)).Msg("server")
			panic(err)
		}
		//fmt.Println(token)

		//sts, err := credentials.NewSTSWebIdentity("https://131.154.97.121:9001/", getWebTokenExpiry)
		providers := []credentials.Provider{
			&IAMProvider{ // nolint:exhaustivestruct
				StsEndpoint: s.S3Endpoint,
				Token:       token,
				HTTPClient:  &s.Client.HTTPClient,
			},
		}

		sts := credentials.NewChainCredentials(providers)
		if err != nil {
			log.Err(fmt.Errorf("Could not set STS credentials: %s", err)).Msg("server")
			panic(err)
		}

		creds, err := sts.Get()
		if err != nil {
			log.Err(fmt.Errorf("Could not get STS credentials: %s", err)).Msg("server")
			panic(err)
		}

		//fmt.Println(creds)

		response := make(map[string]interface{})
		response["credentials"] = creds
		_, err = json.MarshalIndent(response, "", "\t")
		if err != nil {
			log.Err(err).Msg("server")
			panic(err)
		}
	}

	log.Info().Str("s.S3Endpoint", s.S3Endpoint).Msg("server")
	log.Info().Str("s.Instance", s.Instance).Msg("server")

	confRClone := RCloneStruct{
		Address:  s.S3Endpoint,
		Instance: s.Instance,
	}

	tmpl, err := template.New("client").Parse(iamTmpl.RCloneTemplate)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, confRClone)
	if err != nil {
		panic(err)
	}

	rclone := b.String()
	log.Info().Str("rclone config", rclone).Msg("server")

	err = ioutil.WriteFile(s.Client.ConfDir+"/"+"rclone.conf", []byte(rclone), 0600)
	if err != nil {
		panic(err)
	}

	rcloneCmd, errMount := MountVolume(s.Instance, s.RemotePath, s.LocalPath, s.Client.ConfDir)
	if errMount != nil {
		panic(errMount)
	}
	s.rcloneCmd = rcloneCmd

	log.Info().Str("Mounted on", s.LocalPath).Msg("Server")
	color.Green.Printf("==> Volume mounted on %s", s.LocalPath)

	// // TODO: start routine to keep token valid!
	// cntxt := &daemon.Context{
	// 	PidFileName: "mount.pid",
	// 	PidFilePerm: 0644,
	// 	LogFileName: "mount.log",
	// 	LogFilePerm: 0640,
	// 	WorkDir:     "./",
	// }

	// d, err := cntxt.Reborn()
	// if err != nil {
	// 	return err
	// }
	// if d != nil {
	// 	return fmt.Errorf("Process exists")
	// }
	// defer cntxt.Release()

	// log.Print("- - - - - - - - - - - - - - -")
	// log.Print("daemon started")
	return clientResponse, credsIAM, endpoint, nil
}

func (s *Server) UpdateTokenLoop(clientResponse ClientResponse, credsIAM IAMCreds, endpoint string) { //nolint:funlen
	loop := true
	signalChan := make(chan os.Signal, 1)

	signal.Ignore(os.Interrupt)
	signal.Notify(signalChan, os.Interrupt)

	defer close(signalChan)

	startT := time.Now()

	for loop {
		if time.Since(startT) >= time.Duration(s.RefreshTokenRenew)*time.Minute { //nolint:nestif
			startT = time.Now()
			v := url.Values{}

			//fmt.Println(clientResponse.ClientID, clientResponse.ClientSecret, credsIAM.RefreshToken)

			v.Set("client_id", clientResponse.ClientID)
			v.Set("client_secret", clientResponse.ClientSecret)
			v.Set("grant_type", "refresh_token")
			v.Set("refresh_token", credsIAM.RefreshToken)

			url, err := url.Parse(endpoint + "/token" + "?" + v.Encode())
			if err != nil {
				panic(err)
			}

			//fmt.Println(url.String())

			req := http.Request{ // nolint:exhaustivestruct
				Method: "POST",
				URL:    url,
			}

			// TODO: retrieve token with https POST with t.httpClient
			r, err := s.Client.HTTPClient.Do(&req)
			if err != nil {
				panic(err)
			}
			//fmt.Println(r.StatusCode, r.Status)

			var bodyJSON RefreshTokenStruct

			rbody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			//fmt.Println(string(rbody))

			//fmt.Println(string(rbody))
			err = json.Unmarshal(rbody, &bodyJSON)
			if err != nil {
				panic(err)
			}

			// TODO:
			//encrToken := core.Encrypt([]byte(bodyJSON.AccessToken, passwd)

			//fmt.Println(bodyJSON.AccessToken)

			err = ioutil.WriteFile(".token", []byte(bodyJSON.AccessToken), 0600)
			if err != nil {
				panic(err)
			}

			r.Body.Close()
		}

		select {
		case <-signalChan:
			log.Info().Msg("UpdateTokenLoop interrupt signa!")
			loop = false

			log.Info().Msg("Interrupt rclone process")
			_ = s.rcloneCmd.Process.Signal(os.Interrupt)
		default:
		}

		time.Sleep(1 * time.Second)
	}

	signal.Stop(signalChan)

	time.Sleep(1 * time.Second)
	log.Info().Msg("UpdateTokenLoop exit")
}
