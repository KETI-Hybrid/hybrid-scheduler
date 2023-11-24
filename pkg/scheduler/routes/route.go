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

package routes

import (
	"hybrid-scheduler/pkg/scheduler"

	"github.com/julienschmidt/httprouter"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func checkBody(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
}

func PredicateRoute(s *scheduler.Scheduler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)

		var extenderArgs extenderv1.ExtenderArgs
		var extenderFilterResult *extenderv1.ExtenderFilterResult

		if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
			klog.Errorln("decode error", err.Error())
			extenderFilterResult = &extenderv1.ExtenderFilterResult{
				Error: err.Error(),
			}
		} else {
			extenderFilterResult, err = s.Filter(extenderArgs)
			if err != nil {
				klog.Errorf("pod %v filter error, %v", extenderArgs.Pod.Name, err)
				extenderFilterResult = &extenderv1.ExtenderFilterResult{
					Error: err.Error(),
				}
			}
		}

		if resultBody, err := json.Marshal(extenderFilterResult); err != nil {
			klog.Errorf("Failed to marshal extenderFilterResult: %+v, %+v",
				err, extenderFilterResult)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
		klog.Infoln("Out Predicate Route inner func")
	}
}

func Bind(s *scheduler.Scheduler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)
		var extenderBindingArgs extenderv1.ExtenderBindingArgs
		var extenderBindingResult *extenderv1.ExtenderBindingResult

		if err := json.NewDecoder(body).Decode(&extenderBindingArgs); err != nil {
			klog.ErrorS(err, "Decode extender binding args")
			extenderBindingResult = &extenderv1.ExtenderBindingResult{
				Error: err.Error(),
			}
		} else {
			extenderBindingResult, err = s.Bind(extenderBindingArgs)
		}

		if response, err := json.Marshal(extenderBindingResult); err != nil {
			klog.ErrorS(err, "Marshal binding result", "result", extenderBindingResult)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := fmt.Sprintf("{'error':'%s'}", err.Error())
			w.Write([]byte(errMsg))
		} else {
			klog.V(5).InfoS("Return bind response", "result", extenderBindingResult)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}
	}
}

func WebHookRoute() httprouter.Handle {
	h, _ := scheduler.NewWebHook()
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		klog.Infof("Into webhookfunc")
		h.ServeHTTP(w, r)
	}
}
