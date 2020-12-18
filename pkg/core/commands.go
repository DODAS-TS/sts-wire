package core

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/DODAS-TS/sts-wire/pkg/template"
	"github.com/DODAS-TS/sts-wire/pkg/validator"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	programBanner = `
:'######::'########::'######:::::::::::'##:::::'##:'####:'########::'########:
'##... ##:... ##..::'##... ##:::::::::: ##:'##: ##:. ##:: ##.... ##: ##.....::
 ##:::..::::: ##:::: ##:::..::::::::::: ##: ##: ##:: ##:: ##:::: ##: ##:::::::
. ######::::: ##::::. ######::'#######: ##: ##: ##:: ##:: ########:: ######:::
:..... ##:::: ##:::::..... ##:........: ##: ##: ##:: ##:: ##.. ##::: ##...::::
'##::: ##:::: ##::::'##::: ##:::::::::: ##: ##: ##:: ##:: ##::. ##:: ##:::::::
. ######::::: ##::::. ######:::::::::::. ###. ###::'####: ##:::. ##: ########:
:......::::::..::::::......:::::::::::::...::...:::....::..:::::..::........::`
	minNumArgs = 4
)

// Execute of the sts-wire command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	// Used for flags.
	cfgFile    string //nolint:gochecknoglobals
	logFile    string //nolint:gochecknoglobals
	errNumArgs = errors.New("requires the following arguments: <instance name> <s3 endpoint> <rclone remote path> <local mount point>")

	// rootCmd the sts-wire command
	rootCmd = &cobra.Command{ //nolint:exhaustivestruct,gochecknoglobals
		Use:   "sts-wire <instance name> <s3 endpoint> <rclone remote path> <local mount point>",
		Short: "",
		Args: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			if cfgFile == "" {
				if len(args) < minNumArgs {
					return errNumArgs
				}
				if len(args) == 4 {
					if validLocalPath, err := validator.LocalPath(os.Args[4]); !validLocalPath {
						panic(err)
					}
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if logFile != "stderr" {
				logTarget, errLog := os.Create(logFile)
				if errLog != nil {
					panic(errLog)
				}
				log.Logger = zerolog.New(logTarget).With().Timestamp().Logger()
				defer logTarget.Close()
			} else {
				log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
			}

			fmt.Println(programBanner)
			log.Info().Msg("Start sts-wire")

			inputReader := *bufio.NewReader(os.Stdin)
			scanner := GetInputWrapper{
				Scanner: inputReader,
			}

			var (
				instance       string
				confDir        string
				s3Endpoint     string
				remote         string
				localMountPath string
			)

			if cfgFile != "" {
				instance = viper.GetString("instance_name")
				s3Endpoint = viper.GetString("s3_endpoint")
				remote = viper.GetString("rclone_remote_path")
				localMountPath = viper.GetString("local_mount_point")
			} else {
				instance = os.Args[1]
				s3Endpoint = os.Args[2]
				remote = os.Args[3]
				localMountPath = os.Args[4]
			}

			confDir = "." + instance

			_, errStat := os.Stat(confDir)
			if os.IsNotExist(errStat) {
				errMkdir := os.MkdirAll(confDir, os.ModePerm)

				if errMkdir != nil {
					log.Err(errMkdir).Msg("command cannot create instance folder")
					panic(errMkdir)
				}
			}

			log.Info().Str("istance", instance).Msg("command")
			log.Info().Str("s3Endpoint", s3Endpoint).Msg("command")
			log.Info().Str("remote", remote).Msg("command")
			log.Info().Str("localMountPath", localMountPath).Msg("command")
			log.Info().Str("confDir", confDir).Msg("command")

			//fmt.Println(instance)

			// if instance == "" {
			// 	instance, err := scanner.GetInputString("Insert a name for the instance: ", "")
			// 	if err != nil {
			// 		panic(err)
			// 	} else if instance == "" {
			// 		panic(fmt.Errorf("Please insert a valid name."))
			// 	}
			// }

			clientConfig := IAMClientConfig{
				Host:       "localhost",
				Port:       3128,
				ClientName: "oidc-client",
			}

			// Create a CA certificate pool and add cert.pem to it
			//caCert, err := ioutil.ReadFile("MINIO.pem")
			//if err != nil {
			//	log.Fatal(err)
			//}
			//caCertPool := x509.NewCertPool()
			//caCertPool.AppendCertsFromPEM(caCert)

			// Create the TLS Config with the CA pool and enable Client certificate validation
			cfg := &tls.Config{
				//ClientCAs: caCertPool,
				InsecureSkipVerify: true,
			}
			//cfg.BuildNameToCertificate()

			tr := &http.Transport{
				TLSClientConfig: cfg,
			}

			httpClient := &http.Client{
				Transport: tr,
			}

			iamServer := viper.GetString("local_mount_point")
			if os.Getenv("IAM_SERVER") != "" {
				iamServer = os.Getenv("IAM_SERVER")
			}
			log.Info().Str("iamServer", iamServer).Msg("command")

			clientIAM := InitClientConfig{
				ConfDir:        confDir,
				ClientConfig:   clientConfig,
				Scanner:        scanner,
				HTTPClient:     *httpClient,
				IAMServer:      iamServer,
				ClientTemplate: template.ClientTemplate,
				NoPWD:          false,
			}

			// Client registration
			endpoint, clientResponse, _, err := clientIAM.InitClient(instance)
			if err != nil {
				panic(err)
			}

			// TODO: use refresh_token
			if os.Getenv("REFRESH_TOKEN") != "" {
				clientResponse.ClientID = os.Getenv("IAM_CLIENT_ID")
				clientResponse.ClientSecret = os.Getenv("IAM_CLIENT_SECRET")
			}

			// fmt.Println(clientResponse.Endpoint)

			server := Server{
				Client:     clientIAM,
				Instance:   instance,
				S3Endpoint: s3Endpoint,
				RemotePath: remote,
				LocalPath:  localMountPath,
				Endpoint:   endpoint,
				Response:   clientResponse,
			}

			err = server.Start()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Server started and volume mounted in %s", localMountPath)
			fmt.Printf("To unmount you can see you PID in mount.pid file and kill it.")

			return nil
		},
	}
)

// init of the cobra root command and viper configuration
func init() { //nolint: gochecknoinits
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.json", "config file")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "stderr", "where the log has to write, a file path or stderr")

	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(path.Join(home, ".sts-wire"))
	viper.AddConfigPath(".")
}

// initConfig of viper
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			if cfgFile != "./config.yml" {
				fmt.Printf("Info: no '%s' file found\n", cfgFile)
			}
		} else {
			viper.SetConfigFile(cfgFile)
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Info: no configuration file found in all folders, sts-wire will use flags and arguments...")
	} else {
		fmt.Println("Info: using config file ->", viper.ConfigFileUsed())
	}
}
