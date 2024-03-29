/*
Copyright 2022 The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
)

// BucketAccessClassLister helps list BucketAccessClasses.
// All objects returned here must be treated as read-only.
type BucketAccessClassLister interface {
	// List lists all BucketAccessClasses in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.BucketAccessClass, err error)
	// Get retrieves the BucketAccessClass from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.BucketAccessClass, error)
	BucketAccessClassListerExpansion
}

// bucketAccessClassLister implements the BucketAccessClassLister interface.
type bucketAccessClassLister struct {
	indexer cache.Indexer
}

// NewBucketAccessClassLister returns a new BucketAccessClassLister.
func NewBucketAccessClassLister(indexer cache.Indexer) BucketAccessClassLister {
	return &bucketAccessClassLister{indexer: indexer}
}

// List lists all BucketAccessClasses in the indexer.
func (s *bucketAccessClassLister) List(selector labels.Selector) (ret []*v1alpha1.BucketAccessClass, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BucketAccessClass))
	})
	return ret, err
}

// Get retrieves the BucketAccessClass from the index for a given name.
func (s *bucketAccessClassLister) Get(name string) (*v1alpha1.BucketAccessClass, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("bucketaccessclass"), name)
	}
	return obj.(*v1alpha1.BucketAccessClass), nil
}
