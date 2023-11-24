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
	"encoding/json"
	"fmt"
	"hybrid-scheduler/pkg/scheduler/config"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type webhook struct {
	decoder *admission.Decoder
}

func NewWebHook() (*admission.Webhook, error) {
	klog.Infof("Into NewWebHook")
	schema := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(schema); err != nil {
		return nil, err
	}
	decoder, _ := admission.NewDecoder(schema)
	wh := &admission.Webhook{Handler: &webhook{decoder: decoder}}
	_ = wh.InjectLogger(klogr.New())
	return wh, nil
}

func (h *webhook) Handle(_ context.Context, req admission.Request) admission.Response {
	klog.Infoln("Into Handle")
	pod := &corev1.Pod{}
	err := h.decoder.Decode(req, pod)
	if err != nil {
		klog.Errorln(err)
		return admission.Errored(http.StatusBadRequest, err)
	}
	if len(pod.Spec.Containers) == 0 {
		klog.Errorln("pod has no containers")
		return admission.Denied("pod has no containers")
	}

	if len(config.SchedulerName) > 0 {
		pod.Spec.SchedulerName = config.SchedulerName
	}
	marshaledPod, _ := json.Marshal(pod)
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}
