package core

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/DODAS-TS/sts-wire/pkg/template"
	"github.com/DODAS-TS/sts-wire/pkg/validator"
	"github.com/gookit/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/c-bata/go-prompt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	programBanner = `:'######::'########::'######:::::::::::'##:::::'##:'####:'########::'########:
'##... ##:... ##..::'##... ##:::::::::: ##:'##: ##:. ##:: ##.... ##: ##.....::
 ##:::..::::: ##:::: ##:::..::::::::::: ##: ##: ##:: ##:: ##:::: ##: ##:::::::
. ######::::: ##::::. ######::'#######: ##: ##: ##:: ##:: ########:: ######:::
:..... ##:::: ##:::::..... ##:........: ##: ##: ##:: ##:: ##.. ##::: ##...::::
'##::: ##:::: ##::::'##::: ##:::::::::: ##: ##: ##:: ##:: ##::. ##:: ##:::::::
. ######::::: ##::::. ######:::::::::::. ###. ###::'####: ##:::. ##: ########:
:......::::::..::::::......:::::::::::::...::...:::....::..:::::..::........::`
	minNumArgs = 4
)

// Execute of the sts-wire command.
func Execute() {
	fmt.Println(programBanner)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

const (
	errNumArgsS          = "requires the following arguments: <instance name> <s3 endpoint> <rclone remote path> <local mount point>"
	numAcceptedArguments = 4
)

var (
	// Used for flags.
	cfgFile           string //nolint:gochecknoglobals
	logFile           string //nolint:gochecknoglobals
	defaultLogFile    string //nolint:gochecknoglobals
	rcloneMountFlags  string //nolint:gochecknoglobals
	insecureConn      bool   //nolint:gochecknoglobals
	refreshTokenRenew int    //nolint:gochecknoglobals
	noPWD             bool   //nolint:gochecknoglobals
	debug             bool   //nolint:gochecknoglobals
	noModtime         bool   //nolint:gochecknoglobals
	readOnly          bool   //nolint:gochecknoglobals
	tryRemount        bool   //nolint:gochecknoglobals
	errNumArgs        = errors.New(errNumArgsS)

	// rootCmd the sts-wire command.
	rootCmd = &cobra.Command{ //nolint:exhaustivestruct,gochecknoglobals
		Use:   "sts-wire <IAM server> <instance name> <s3 endpoint> <rclone remote path> <local mount point>",
		Short: "",
		Args: func(cmd *cobra.Command, args []string) error {
			if cfgFile == "" {
				if len(args) < minNumArgs {
					return errNumArgs
				}
				if len(args) == numAcceptedArguments {
					if validIAMServer, err := validator.WebURL(os.Args[1]); !validIAMServer {
						panic(err)
					}
					if validInstanceName, err := validator.InstanceName(os.Args[2]); !validInstanceName {
						panic(err)
					}
					if validEndpoint, err := validator.S3Endpoint(os.Args[3]); !validEndpoint {
						panic(err)
					}
					if validRemotePath, err := validator.RemotePath(os.Args[4]); !validRemotePath {
						panic(err)
					}
					if validLocalPath, err := validator.LocalPath(os.Args[5]); !validLocalPath {
						panic(err)
					}
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if confLog := viper.GetString("log"); logFile == defaultLogFile && confLog != "" {
				logFile = confLog
			}

			if debug {
				logFile = "stderr"
			}

			var firstLogWriter *os.File

			if logFile != "stderr" {
				if valid, err := validator.LogFile(logFile); !valid {
					panic(err)
				}
				_, errBaseDir := os.Stat(filepath.Dir(logFile))
				if errBaseDir != nil && os.IsNotExist(errBaseDir) {
					errMkdirs := os.MkdirAll(filepath.Dir(logFile), os.ModePerm)
					if errMkdirs != nil {
						panic(errMkdirs)
					}
				}

				logTarget, errOpenLog := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, fileMode)
				if errOpenLog != nil {
					panic(errOpenLog)
				}

				firstLogWriter = logTarget
				log.Logger = zerolog.New(logTarget).With().Timestamp().Logger()
				defer logTarget.Close()
			} else {
				log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}) // nolint:exhaustivestruct
			}

			log.Debug().Str("log file", logFile).Msg("logging")
			log.Debug().Msg("Start sts-wire")

			inputReader := *bufio.NewReader(os.Stdin)
			scanner := GetInputWrapper{
				Scanner: inputReader,
			}

			var (
				iamServer      string
				instance       string
				confDir        string
				s3Endpoint     string
				remote         string
				localMountPath string
			)

			if cfgFile != "" {
				iamServer = viper.GetString("IAM_Server")
				instance = viper.GetString("instance_name")
				s3Endpoint = viper.GetString("s3_endpoint")
				remote = viper.GetString("rclone_remote_path")
				localMountPath = viper.GetString("local_mount_point")
			} else {
				iamServer = os.Args[1]
				instance = os.Args[2]
				s3Endpoint = os.Args[3]
				remote = os.Args[4]
				localMountPath = os.Args[5]
			}

			// TODO: check if it is useful or not to have env variable overwrite mechanism
			// ENV VARIABLE OVERWRITE
			if os.Getenv("IAM_SERVER") != "" {
				iamServer = os.Getenv("IAM_SERVER")
			}

			noPWD = noPWD || viper.GetBool("noPassword")
			noModtime = noModtime || viper.GetBool("noModtime")
			readOnly = readOnly || viper.GetBool("readOnly")
			tryRemount = tryRemount || viper.GetBool("tryRemount")

			log.Debug().Str("iamServer", iamServer).Msg("command")
			log.Debug().Str("istance", instance).Msg("command")
			log.Debug().Str("s3Endpoint", s3Endpoint).Msg("command")
			log.Debug().Str("remote", remote).Msg("command")
			log.Debug().Str("localMountPath", localMountPath).Msg("command")
			log.Debug().Bool("noPassword", noPWD).Msg("command")
			log.Debug().Bool("noModtime", noModtime).Msg("command")
			log.Debug().Bool("readOnly", readOnly).Msg("command")
			log.Debug().Bool("tryRemount", tryRemount).Msg("command")

			if cfgFile != "" {
				if validIAMServer, err := validator.WebURL(iamServer); !validIAMServer {
					panic(fmt.Errorf("not a valid IAM server %w", err))
				}
				if validInstanceName, err := validator.InstanceName(instance); !validInstanceName {
					panic(err)
				}
				if validEndpoint, err := validator.S3Endpoint(s3Endpoint); !validEndpoint {
					panic(err)
				}
				if validRemotePath, err := validator.RemotePath(remote); !validRemotePath {
					panic(err)
				}
				if validLocalPath, err := validator.LocalPath(localMountPath); !validLocalPath {
					panic(err)
				}
			}

			// -------------------- CONFIG IAM URL AND PORT --------------------
			iamcURL := "localhost"
			var iamcPort int

			if newIamcURL := viper.GetString("IAMAuthURL"); newIamcURL != "" {
				if valid, err := validator.WebURL(newIamcURL); !valid || err != nil {
					panic(err)
				}
				iamcURL = newIamcURL
			}

			if newIamcPort := viper.GetInt("IAMAuthURLPort"); newIamcPort > 0 {
				iamcPort = newIamcPort
			} else {
				// select a random port available from the OS
				randomPort, errRandPort := availableRandomPort()
				if errRandPort != nil {
					log.Err(errRandPort).Msg("server")

					panic(errRandPort)
				}

				randomPortInt, errConv := strconv.ParseInt(randomPort, 10, 64)
				if errConv != nil {
					log.Err(errConv).Msg("server")

					panic(errConv)
				}

				iamcPort = int(randomPortInt)
			}

			log.Debug().Str("iamcURL", iamcURL).Msg("command")
			log.Debug().Int("iamcPort", iamcPort).Msg("command")

			clientConfig := IAMClientConfig{ // nolint:exhaustivestruct
				Host:       iamcURL,
				Port:       iamcPort,
				ClientName: "oidc-client",
			}

			// ------------------------ CONFIG INSTANCE ------------------------
			confDir = "." + instance

			_, errStat := os.Stat(confDir)
			if os.IsNotExist(errStat) {
				errMkdir := os.MkdirAll(confDir, os.ModePerm)

				if errMkdir != nil {
					log.Err(errMkdir).Msg("command cannot create instance folder")
					panic(errMkdir)
				}

			}

			log.Debug().Str("confDir", confDir).Msg("command")

			// ---------------------- CONFIG INSTANCE LOG ----------------------
			instanceLogFilename = filepath.Join(confDir, "instance.log")

			instanceLogFile, errOpenLog := os.OpenFile(instanceLogFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, fileMode)
			if errOpenLog != nil {
				panic(errOpenLog)
			}

			defer instanceLogFile.Close()

			var multi zerolog.LevelWriter
			if firstLogWriter != nil {
				multi = zerolog.MultiLevelWriter(firstLogWriter, instanceLogFile)
			} else {
				multi = zerolog.MultiLevelWriter(
					zerolog.ConsoleWriter{Out: os.Stderr}, //nolint: exhaustivestruct
					instanceLogFile,
				)
			}
			log.Logger = zerolog.New(multi).With().Timestamp().Logger()

			// ---------------------- CONFIG INSTANCE INFO ---------------------
			instanceFilename := path.Join(confDir, "instance.info")

			_, errStats := os.Stat(instanceFilename)
			if os.IsNotExist(errStats) {
				instanceRef, errInstanceFile := os.OpenFile(instanceFilename, os.O_WRONLY|os.O_CREATE, fileMode)
				if errInstanceFile != nil {
					panic(errInstanceFile)
				}

				infoBytes, errMarshall := json.MarshalIndent(InstanceInfo{
					Name:     instance,
					LogFile:  "./instance.log",
					Port:     iamcPort,
					Password: !noPWD,
				}, "", "  ")

				if errMarshall != nil {
					panic(errMarshall)
				}

				if _, errWrite := instanceRef.Write(infoBytes); errWrite != nil {
					panic(errWrite)
				}

				instanceRef.Close()
			} else {
				instanceFile, errOpenLog := os.Open(instanceFilename)
				if errOpenLog != nil {
					panic(errOpenLog)
				}

				defer instanceFile.Close()

				var curInstanceInfo InstanceInfo
				var buffer bytes.Buffer

				_, errReadInstance := buffer.ReadFrom(instanceFile)
				if errReadInstance != nil {
					panic(errOpenLog)
				}

				if errUnmarshal := json.Unmarshal(buffer.Bytes(), &curInstanceInfo); errUnmarshal != nil {
					panic(errUnmarshal)
				}

				if curInstanceInfo.Password {
					clientConfig.Port = curInstanceInfo.Port
				}
			}

			// ------------------- CONFIG REFRESH TOKEN INFO -------------------
			if newRefreshTokenRenew := viper.GetInt("refreshTokenRenew"); newRefreshTokenRenew != 0 && refreshTokenRenew == 15 {
				refreshTokenRenew = newRefreshTokenRenew
			}

			if valid, errRefreshToken := validator.RefreshTokenRenew(refreshTokenRenew); errRefreshToken != nil || !valid {
				panic(errRefreshToken)
			}

			log.Debug().Int("refreshTokenRenew", refreshTokenRenew).Msg("command")
			log.Debug().Str("rcloneMountFlags", rcloneMountFlags).Msg("command")

			if rcloneMountFlags != "" {
				if validMountFlags, err := validator.RcloneMountFlags(rcloneMountFlags); !validMountFlags {
					panic(fmt.Errorf("not valid rclone mount flags %w", err))
				}
			}

			// Create a CA certificate pool and add cert.pem to it
			// TODO: convert ioutil.ReadFile to os
			// 		 ref: https://www.srcbeat.com/2021/01/golang-ioutil-deprecated/
			// caCert, err := ioutil.ReadFile("MINIO.pem")
			// if err != nil {
			//	log.Fatal(err)
			// }
			// caCertPool := x509.NewCertPool()
			// caCertPool.AppendCertsFromPEM(caCert)

			// Create the TLS Config with the CA pool and enable Client certificate validation

			if insecureConnVip := viper.GetBool("insecureConn"); insecureConnVip != insecureConn {
				insecureConn = insecureConnVip
			}
			log.Debug().Bool("insecureConn", insecureConn).Msg("command")

			cfg := &tls.Config{ // nolint: exhaustivestruct
				// ClientCAs: caCertPool,
				InsecureSkipVerify: insecureConn, // nolint:gosec
			}
			// cfg.BuildNameToCertificate()

			tr := &http.Transport{ // nolint:exhaustivestruct
				TLSClientConfig: cfg,
			}

			httpClient := &http.Client{ // nolint:exhaustivestruct
				Transport: tr,
			}

			clientIAM := InitClientConfig{
				ConfDir:        confDir,
				ClientConfig:   clientConfig,
				Scanner:        scanner,
				HTTPClient:     *httpClient,
				IAMServer:      iamServer,
				ClientTemplate: template.ClientTemplate,
				NoPWD:          noPWD,
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

			server := Server{
				Client:            clientIAM,
				Instance:          instance,
				S3Endpoint:        s3Endpoint,
				RemotePath:        remote,
				LocalPath:         localMountPath,
				Endpoint:          endpoint,
				CurClientResponse: clientResponse,
				RefreshTokenRenew: refreshTokenRenew,
				ReadOnly:          readOnly,
				NoModtime:         noModtime,
				MountNewFlags:     rcloneMountFlags,
				TryRemount:        tryRemount,
			}

			credsIAM, endpoint, errStart := server.Start()
			if errStart != nil {
				panic(errStart)
			}

			if refreshToken := os.Getenv("REFRESH_TOKEN"); refreshToken != "" {
				log.Debug().Str("refreshToken", refreshToken).Msg("Force refresh token call")
				server.RefreshToken(credsIAM, endpoint)
			}

			color.Green.Printf("==> Server started successfully and volume mounted at %s\n", localMountPath)

			server.UpdateTokenLoop(credsIAM, endpoint)

			return nil
		},
	}

	versionCmd = &cobra.Command{ // nolint:exhaustivestruct,gochecknoglobals
		Use:   "version",
		Short: "Print the version number of sts-wire",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(buildCmdVersion())
		},
	}

	cleanCmd = &cobra.Command{ // nolint:exhaustivestruct,gochecknoglobals
		Use:   "clean",
		Short: "Clean sts-wire stuff",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("=> Get rclone path\n")
			rcloneExePath, err := ExePath()
			if err != nil {
				panic(err)
			}

			fmt.Printf("==> Remove rclone: %s\n", rcloneExePath)
			err = os.Remove(rcloneExePath)
			if err != nil && !os.IsNotExist(err) {
				panic(err)
			}

			pattern := strings.Builder{}

			pattern.WriteString(".")
			pattern.WriteRune(os.PathSeparator)
			pattern.WriteString(".**")
			pattern.WriteRune(os.PathSeparator)
			pattern.WriteString("instance.info")

			matches, _ := filepath.Glob(pattern.String())
			for _, match := range matches {
				curDir := filepath.Dir(match)

				fmt.Printf("=> Remove instance folder: %s\n", curDir)
				os.RemoveAll(curDir)
			}

			logFiles, _ := filepath.Glob(filepath.Join(getBaseLogDir(), "log", "*.log"))
			for _, curLog := range logFiles {
				fmt.Printf("=> Remove log: %s\n", curLog)
				os.RemoveAll(curLog)
			}

			fmt.Printf("==> sts-wire env cleaned!\n")
		},
	}

	reportCmd = &cobra.Command{ // nolint:exhaustivestruct,gochecknoglobals
		Use:   "report",
		Short: "search and open sts-wire reports",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("==> Please select a report:\n")
			report := prompt.Input("> ", reportCompleter)

			if report == "quit" {
				os.Exit(0)
			}

			reportFile, err := os.Open(report)
			if err != nil {
				panic(err)
			}

			defer reportFile.Close()

			var buffer bytes.Buffer

			_, err = buffer.ReadFrom(reportFile)
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s\n%s\n", divider, buffer.String())
		},
	}
)

func reportCompleter(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{}

	suggestions = append(suggestions, prompt.Suggest{
		Text: "quit", Description: "exit from report command",
	})

	matches, _ := filepath.Glob("./.**/report_*.out")
	for _, match := range matches {
		suggestions = append(suggestions, prompt.Suggest{
			Text: match, Description: fmt.Sprintf("folder -> %s", path.Dir(match)),
		})
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func buildCmdVersion() string {
	versionString := strings.Builder{}
	versionString.WriteString(divider)
	versionString.WriteRune('\n')
	versionString.WriteString(fmt.Sprintf(" Version:\t\t%s\n", StsVersion))
	versionString.WriteString(fmt.Sprintf(" Git Commit:\t\t%s\n", GitCommit))
	versionString.WriteString(fmt.Sprintf(" Go Version:\t\t%s\n", runtime.Version()))
	versionString.WriteString(fmt.Sprintf(" Built Time:\t\t%s\n", BuiltTime))
	versionString.WriteString(fmt.Sprintf(" OS/Arch:\t\t%s\n", OsArch))
	versionString.WriteString(fmt.Sprintf(" Rclone Version:\t%s\n", RcloneVersion))
	versionString.WriteString(divider)

	return versionString.String()
}

func getBaseLogDir() (baseLogDir string) {
	baseLogDir, errConfDir := os.UserConfigDir()
	if errConfDir != nil {
		curDir, errAbsCurDir := filepath.Abs(filepath.Dir(os.Args[0]))
		if errAbsCurDir != nil {
			panic(errAbsCurDir)
		}

		baseLogDir = curDir
	}

	return baseLogDir
}

// init of the cobra root command and viper configuration.
func init() { //nolint: gochecknoinits
	cobra.OnInitialize(initConfig)

	defaultLogFile = filepath.Join(getBaseLogDir(), "log", "sts-wire.log")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.json", "config file")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", defaultLogFile,
		"where the log has to write, a file path or stderr")
	rootCmd.PersistentFlags().StringVar(&rcloneMountFlags, "rcloneMountFlags", rcloneMountFlags,
		"overwrite the rclone mount flags")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "start the program in debug mode")
	rootCmd.PersistentFlags().BoolVar(&insecureConn, "insecureConn", false, "check the http connection certificate")
	rootCmd.PersistentFlags().IntVar(&refreshTokenRenew, "refreshTokenRenew", 15,
		"time span to renew the refresh token in minutes")
	rootCmd.PersistentFlags().BoolVar(&noPWD, "noPassword", false, "to not encrypt the data with a password")
	rootCmd.PersistentFlags().BoolVar(&noModtime, "noModtime", false, "mount with noModtime option")
	rootCmd.PersistentFlags().BoolVar(&readOnly, "readOnly", false, "mount with read-only option")
	rootCmd.PersistentFlags().BoolVar(&tryRemount, "tryRemount", true,
		"try to remount if there are any rclone errors (up to 10 times)")

	errFlag := viper.BindPFlag("insecureConn", rootCmd.PersistentFlags().Lookup("insecureConn"))
	if errFlag != nil {
		panic(errFlag)
	}

	errFlag = viper.BindPFlag("refreshTokenRenew", rootCmd.PersistentFlags().Lookup("refreshTokenRenew"))
	if errFlag != nil {
		panic(errFlag)
	}

	errFlag = viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
	if errFlag != nil {
		panic(errFlag)
	}

	errFlag = viper.BindPFlag("rcloneMountFlags", rootCmd.PersistentFlags().Lookup("rcloneMountFlags"))
	if errFlag != nil {
		panic(errFlag)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(path.Join(home, ".sts-wire"))
	viper.AddConfigPath(".")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(reportCmd)
}

// initConfig of viper.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		if _, err := os.Stat(cfgFile); err == nil || os.IsExist(err) {
			viper.SetConfigFile(cfgFile)
		} else {
			cfgFile = ""
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("==> sts-wire is using config file: %s\n", viper.ConfigFileUsed())
	}
}
