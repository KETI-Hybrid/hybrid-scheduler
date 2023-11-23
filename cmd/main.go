/*
 * Copyright © 2021 peizhaoyou <peizhaoyou@4paradigm.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"net/http"

	"hybrid-scheduler/pkg/util"
	"hybrid-scheduler/pkg/version"

	"hybrid-scheduler/pkg/scheduler"
	"hybrid-scheduler/pkg/scheduler/config"
	"hybrid-scheduler/pkg/scheduler/routes"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

//var version 

var (
	sher        *scheduler.Scheduler
	tlsKeyFile  string
	tlsCertFile string
	rootCmd     = &cobra.Command{
		Use:   "scheduler",
		Short: "kubernetes scheduler",
		Run: func(cmd *cobra.Command, args []string) {
			start()
		},
	}
)

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.Flags().StringVar(&config.HttpBind, "http_bind", "127.0.0.1:8080", "http server bind address")
	rootCmd.Flags().StringVar(&tlsCertFile, "cert_file", "", "tls cert file")
	rootCmd.Flags().StringVar(&tlsKeyFile, "key_file", "", "tls key file")
	rootCmd.Flags().StringVar(&config.SchedulerName, "scheduler-name", "", "the name to be added to pod.spec.schedulerName if not empty")
	rootCmd.PersistentFlags().AddGoFlagSet(util.GlobalFlagSet())
	rootCmd.AddCommand(version.VersionCmd)
}

func start() {
	sher = scheduler.NewScheduler()
	sher.Start()
	defer sher.Stop()
	go sher.RegisterFromNodeAnnotatons()

	// start http server
	router := httprouter.New()
	router.POST("/filter", routes.PredicateRoute(sher))
	router.POST("/bind", routes.Bind(sher))
	router.POST("/webhook", routes.WebHookRoute())
	//klog.Info("listen on ", config.HttpBind)
	if len(tlsCertFile) == 0 || len(tlsKeyFile) == 0 {
		klog.Info("listen on HTTP ", config.HttpBind)
		if err := http.ListenAndServe(config.HttpBind, router); err != nil {
			klog.Fatal("Listen and Serve error, ", err)
		}
	} else {
		klog.Info("listen on HTTPS TLS ", config.HttpBind)
		if err := http.ListenAndServeTLS(config.HttpBind, tlsCertFile, tlsKeyFile, router); err != nil {
			klog.Fatal("Listen and Serve error, ", err)
		}
	}
	
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}
