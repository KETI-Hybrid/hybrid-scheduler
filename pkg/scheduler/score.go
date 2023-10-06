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

package scheduler

import (
	"context"
	"time"

	"hybrid-scheduler/pkg/util/client"
	"hybrid-scheduler/pkg/util/client/score"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

type NodeScore map[string]float32

func getScore() NodeScore {
	kubeClient, err := client.NewClient()
	if err != nil {
		klog.Errorln(err)
	}
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"name": "analysis-engine"}}
	pods, err := kubeClient.CoreV1().Pods("keti-system").List(context.Background(), metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:         100,
	})
	if err != nil {
		klog.Errorln(err)
	}
	scorepod := pods.Items[0]
	podIP := scorepod.Status.PodIP
	host := podIP + ":50051"
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Errorln("did not connect: %v", err)
	}
	defer conn.Close()
	metricClient := score.NewMetricGRPCClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := metricClient.GetNodeScore(ctx, &score.Request{})
	if err != nil {
		klog.Errorf("could not request: %v \n", err)
	}
	return r.Message
}
