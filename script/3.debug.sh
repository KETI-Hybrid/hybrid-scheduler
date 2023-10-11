#/bin/bash
NS=keti-system
NAME=$(kubectl get pod -n $NS | grep -E 'hybrid-scheduler' | awk '{print $1}')

kubectl logs -f -n $NS $NAME -c hybrid-scheduler

