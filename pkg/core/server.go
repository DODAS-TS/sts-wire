package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	Client     InitClientConfig
	Instance   string
	S3Endpoint string
	RemotePath string
	LocalPath  string
	Endpoint   string
	Response   ClientResponse
}

// Start ..
func (s *Server) Start() error {

	state := RandomState()
	isByHand := os.Getenv("REFRESH_TOKEN")
	credsIAM := IAMCreds{}
	endpoint := s.Endpoint
	clientResponse := s.Response

	if isByHand == "" {

		endpoint := s.Endpoint
		clientResponse := s.Response

		sigint := make(chan int, 1)

		//fmt.Println(clientResponse.ClientID)
		//fmt.Println(clientResponse.ClientSecret)

		ctx := context.Background()

		config := oauth2.Config{
			ClientID:     clientResponse.ClientID,
			ClientSecret: clientResponse.ClientSecret,
			Endpoint: oauth2.Endpoint{
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
				&IAMProvider{
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
			w.Write([]byte(htmlMountingPage))

			sigint <- 1

		})

		address := fmt.Sprintf("localhost:3128")
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

		srv := &http.Server{Addr: address}

		idleConnsClosed := make(chan struct{})
		go func() {
			<-sigint

			// We received an interrupt signal, shut down.
			if err := srv.Shutdown(context.Background()); err != nil {
				// Error from closing listeners, or context timeout:
				log.Err(err).Msg("server")
			}
			close(idleConnsClosed)
		}()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
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
			&IAMProvider{
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

	MountVolume(s.Instance, s.RemotePath, s.LocalPath, s.Client.ConfDir)

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

	for {
		v := url.Values{}

		//fmt.Println(clientResponse.ClientID, clientResponse.ClientSecret, credsIAM.RefreshToken)

		v.Set("client_id", clientResponse.ClientID)
		v.Set("client_secret", clientResponse.ClientSecret)
		v.Set("grant_type", "refresh_token")
		v.Set("refresh_token", credsIAM.RefreshToken)

		url, err := url.Parse(endpoint + "/token" + "?" + v.Encode())

		//fmt.Println(url.String())

		req := http.Request{
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

		time.Sleep(10 * time.Minute)

	}
}

const (
	htmlMountingPage = `<!DOCTYPE HTML>
<html>
	<head>
	<title>STS-WIRE</title>
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<style>
		html,
		body {
		height: 100%;
		}
		body {
		display: flex;
		flex-wrap: wrap;
		margin: 0;
		}
		.header-menu,
		footer {
		display: flex;
		align-items: center;
		width: 100%;
		}
		.header-menu {
		justify-content: center;
		height: 60px;
		background: #1c87c9;
		color: #fff;
		}
		h2 {
		margin: 0 0 8px;
		}
		ul li {
		display: inline-block;
		padding: 0 10px;
		list-style: none;
		}
		section {
		flex: 1;
		margin: auto;
		width: 50%;
		padding: 10px;
		}
		article {
		margin: auto;
		width: 50%;
		padding: 10px;
		}
		footer {
		padding: 0 10px;
		background: #ccc;
		}
	</style>
	</head>
	<body>
	<header class="header-menu">
		<h1>STS-WIRE</h1>
	</header>
	<section>
		<article>
		<header>
			<h2>Info</h2>
			<small>Mounting process</small>
		</header>
		<p>Volume will be mounted in few seconds... You can close this tab!</p>
		</article>
	</section>
	<footer>
		<small>© DODAS-TS. All rights reserved</small>
	</footer>
	</body>
</html>`
)
