/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/clientset/versioned/typed/shahpratikr.dev/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeShahpratikrV1alpha1 struct {
	*testing.Fake
}

func (c *FakeShahpratikrV1alpha1) SnapshotBackups(namespace string) v1alpha1.SnapshotBackupInterface {
	return &FakeSnapshotBackups{c, namespace}
}

func (c *FakeShahpratikrV1alpha1) SnapshotRestores(namespace string) v1alpha1.SnapshotRestoreInterface {
	return &FakeSnapshotRestores{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeShahpratikrV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
