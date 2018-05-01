// Copyright © 2018 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/kubicorn/controller/service"
	"github.com/kubicorn/controller/service/aws"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	cfg = &service.ServiceConfiguration{}
	cp  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubicorn-controller",
	Short: "The Kubicorn machine controller",
	Long:  `Run the Kubicorn controller to reconcile your infrastructure like the beautiful person you are.`,
	Run: func(cmd *cobra.Command, args []string) {
		//
		//
		// Environmental Variables
		//
		//
		kubeConfigContent := os.Getenv("KUBECONFIG_CONTENT")
		if kubeConfigContent == "" {
			logger.Critical("Missing environmental variable [KUBECONFIG_CONTENT]")
			os.Exit(95)
		}
		cfg.KubeConfigContent = kubeConfigContent
		awsKey := os.Getenv("AWS_ACCESS_KEY_ID")
		if awsKey == "" {
			logger.Critical("Missing environmental variable [AWS_ACCESS_KEY_ID]")
			os.Exit(94)
		}
		awsSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if awsSecret == "" {
			logger.Critical("Missing environmental variable [AWS_SECRET_ACCESS_KEY]")
			os.Exit(93)
		}
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			logger.Critical("Missing environmental variable [AWS_REGION]")
			os.Exit(92)
		}
		awsProfile := os.Getenv("AWS_PROFILE")
		if awsProfile == "" {
			logger.Info("Missing AWS_PROFILE using: DEFAULT")
			awsProfile = "default"
		}
		//
		//
		// Cloud Provider Switch
		//
		//
		cp = strings.ToLower(cp)
		switch cp {
		case "aws":
			cp, err := aws.New(awsRegion, awsProfile)
			if err != nil {
				logger.Critical("Error loading SDK: %v", err)
				os.Exit(91)
			}
			cfg.CloudProvider = cp
		default:
			err := fmt.Errorf("Invalid cloud provider string: %s", cp)
			logger.Critical(err.Error())
			os.Exit(97)
		}
		//
		//
		// Run Service
		//
		//
		service.RunService(cfg)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	//
	//
	// Flags
	//
	//
	rootCmd.PersistentFlags().IntVarP(&logger.Level, "verbose", "v", 4, "Log level")
	//rootCmd.Flags().StringVarP(&cfg.KubeConfigContent, "kubeconfig-content", "k", "", "The content of the kubeconfig file to authenticate with.")
	rootCmd.Flags().StringVarP(&cp, "cloud-provider", "c", "aws", "The cloud provider string to use. Available options: aws")

}
