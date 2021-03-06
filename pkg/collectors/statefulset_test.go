/*
Copyright 2017 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collectors

import (
	"testing"
	"time"

	"k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/collectors/testutils"
	"k8s.io/kube-state-metrics/pkg/options"
)

var (
	statefulSet1Replicas int32 = 3
	statefulSet2Replicas int32 = 6
	statefulSet3Replicas int32 = 9

	statefulSet1ObservedGeneration int64 = 1
	statefulSet2ObservedGeneration int64 = 2
)

type mockStatefulSetStore struct {
	f func() ([]v1beta1.StatefulSet, error)
}

func (ds mockStatefulSetStore) List() (deployments []v1beta1.StatefulSet, err error) {
	return ds.f()
}

func TestStatefuleSetCollector(t *testing.T) {
	// Fixed metadata on type and help text. We prepend this to every expected
	// output so we only have to modify a single place when doing adjustments.
	const metadata = `
		# HELP kube_statefulset_created Unix creation timestamp
		# TYPE kube_statefulset_created gauge
		# HELP kube_statefulset_status_current_revision Indicates the version of the StatefulSet used to generate Pods in the sequence [0,currentReplicas).
		# TYPE kube_statefulset_status_current_revision gauge
 		# HELP kube_statefulset_status_replicas The number of replicas per StatefulSet.
 		# TYPE kube_statefulset_status_replicas gauge
		# HELP kube_statefulset_status_replicas_current The number of current replicas per StatefulSet.
		# TYPE kube_statefulset_status_replicas_current gauge
		# HELP kube_statefulset_status_replicas_ready The number of ready replicas per StatefulSet.
		# TYPE kube_statefulset_status_replicas_ready gauge
		# HELP kube_statefulset_status_replicas_updated The number of updated replicas per StatefulSet.
		# TYPE kube_statefulset_status_replicas_updated gauge
 		# HELP kube_statefulset_status_observed_generation The generation observed by the StatefulSet controller.
 		# TYPE kube_statefulset_status_observed_generation gauge
		# HELP kube_statefulset_status_update_revision Indicates the version of the StatefulSet used to generate Pods in the sequence [replicas-updatedReplicas,replicas)
		# TYPE kube_statefulset_status_update_revision gauge
 		# HELP kube_statefulset_replicas Number of desired pods for a StatefulSet.
 		# TYPE kube_statefulset_replicas gauge
 		# HELP kube_statefulset_metadata_generation Sequence number representing a specific generation of the desired state for the StatefulSet.
 		# TYPE kube_statefulset_metadata_generation gauge
		# HELP kube_statefulset_labels Kubernetes labels converted to Prometheus labels.
		# TYPE kube_statefulset_labels gauge
 	`
	cases := []struct {
		depls []v1beta1.StatefulSet
		want  string
	}{
		{
			depls: []v1beta1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "statefulset1",
						CreationTimestamp: metav1.Time{Time: time.Unix(1500000000, 0)},
						Namespace:         "ns1",
						Labels: map[string]string{
							"app": "example1",
						},
						Generation: 3,
					},
					Spec: v1beta1.StatefulSetSpec{
						Replicas:    &statefulSet1Replicas,
						ServiceName: "statefulset1service",
					},
					Status: v1beta1.StatefulSetStatus{
						ObservedGeneration: &statefulSet1ObservedGeneration,
						Replicas:           2,
						CurrentRevision:    "cr1",
						UpdateRevision:     "ur1",
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "statefulset2",
						Namespace: "ns2",
						Labels: map[string]string{
							"app": "example2",
						},
						Generation: 21,
					},
					Spec: v1beta1.StatefulSetSpec{
						Replicas:    &statefulSet2Replicas,
						ServiceName: "statefulset2service",
					},
					Status: v1beta1.StatefulSetStatus{
						CurrentReplicas:    2,
						ObservedGeneration: &statefulSet2ObservedGeneration,
						ReadyReplicas:      5,
						Replicas:           5,
						UpdatedReplicas:    3,
						CurrentRevision:    "cr2",
						UpdateRevision:     "ur2",
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "statefulset3",
						Namespace: "ns3",
						Labels: map[string]string{
							"app": "example3",
						},
						Generation: 36,
					},
					Spec: v1beta1.StatefulSetSpec{
						Replicas:    &statefulSet3Replicas,
						ServiceName: "statefulset2service",
					},
					Status: v1beta1.StatefulSetStatus{
						ObservedGeneration: nil,
						Replicas:           7,
						CurrentRevision:    "cr3",
						UpdateRevision:     "ur3",
					},
				},
			},
			want: metadata + `
				kube_statefulset_created{namespace="ns1",statefulset="statefulset1"} 1.5e+09
				kube_statefulset_status_current_revision{namespace="ns1",revision="cr1",statefulset="statefulset1"} 1
				kube_statefulset_status_current_revision{namespace="ns2",revision="cr2",statefulset="statefulset2"} 1
				kube_statefulset_status_current_revision{namespace="ns3",revision="cr3",statefulset="statefulset3"} 1
 				kube_statefulset_status_replicas{namespace="ns1",statefulset="statefulset1"} 2
 				kube_statefulset_status_replicas{namespace="ns2",statefulset="statefulset2"} 5
 				kube_statefulset_status_replicas{namespace="ns3",statefulset="statefulset3"} 7
				kube_statefulset_status_replicas_current{namespace="ns1",statefulset="statefulset1"} 0
				kube_statefulset_status_replicas_current{namespace="ns2",statefulset="statefulset2"} 2
				kube_statefulset_status_replicas_current{namespace="ns3",statefulset="statefulset3"} 0
				kube_statefulset_status_replicas_ready{namespace="ns1",statefulset="statefulset1"} 0
				kube_statefulset_status_replicas_ready{namespace="ns2",statefulset="statefulset2"} 5
				kube_statefulset_status_replicas_ready{namespace="ns3",statefulset="statefulset3"} 0
				kube_statefulset_status_replicas_updated{namespace="ns1",statefulset="statefulset1"} 0
				kube_statefulset_status_replicas_updated{namespace="ns2",statefulset="statefulset2"} 3
				kube_statefulset_status_replicas_updated{namespace="ns3",statefulset="statefulset3"} 0
 				kube_statefulset_status_observed_generation{namespace="ns1",statefulset="statefulset1"} 1
 				kube_statefulset_status_observed_generation{namespace="ns2",statefulset="statefulset2"} 2
				kube_statefulset_status_update_revision{namespace="ns1",revision="ur1",statefulset="statefulset1"} 1
				kube_statefulset_status_update_revision{namespace="ns2",revision="ur2",statefulset="statefulset2"} 1
				kube_statefulset_status_update_revision{namespace="ns3",revision="ur3",statefulset="statefulset3"} 1
 				kube_statefulset_replicas{namespace="ns1",statefulset="statefulset1"} 3
 				kube_statefulset_replicas{namespace="ns2",statefulset="statefulset2"} 6
 				kube_statefulset_replicas{namespace="ns3",statefulset="statefulset3"} 9
 				kube_statefulset_metadata_generation{namespace="ns1",statefulset="statefulset1"} 3
 				kube_statefulset_metadata_generation{namespace="ns2",statefulset="statefulset2"} 21
 				kube_statefulset_metadata_generation{namespace="ns3",statefulset="statefulset3"} 36
				kube_statefulset_labels{label_app="example1",namespace="ns1",statefulset="statefulset1"} 1
				kube_statefulset_labels{label_app="example2",namespace="ns2",statefulset="statefulset2"} 1
				kube_statefulset_labels{label_app="example3",namespace="ns3",statefulset="statefulset3"} 1
 			`,
		},
	}
	for _, c := range cases {
		sc := &statefulSetCollector{
			store: mockStatefulSetStore{
				f: func() ([]v1beta1.StatefulSet, error) { return c.depls, nil },
			},
			opts: &options.Options{},
		}
		if err := testutils.GatherAndCompare(sc, c.want, nil); err != nil {
			t.Errorf("unexpected collecting result:\n%s", err)
		}
	}
}
