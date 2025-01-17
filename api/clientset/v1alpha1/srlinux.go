// Copyright 2022 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

// Package v1alpha1 is an v1alpha version of a Clientset for SR Linux customer resource.
package v1alpha1

// note to my future self: see https://www.martin-helmich.de/en/blog/kubernetes-crd-client.html for details

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	typesv1alpha1 "github.com/srl-labs/srl-controller/api/types/v1alpha1"
)

// ErrUpdateFailed occurs when update operation fails on srlinux CR.
var ErrUpdateFailed = errors.New("operation update failed")

// SrlinuxInterface provides access to the Srlinux CRD.
type SrlinuxInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*typesv1alpha1.SrlinuxList, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*typesv1alpha1.Srlinux, error)
	Create(ctx context.Context, srlinux *typesv1alpha1.Srlinux) (*typesv1alpha1.Srlinux, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Unstructured(ctx context.Context, name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*typesv1alpha1.Srlinux, error)
}

// Interface is the clientset interface for srlinux.
type Interface interface {
	Srlinux(namespace string) SrlinuxInterface
}

// Clientset is a client for the srlinux crds.
type Clientset struct {
	dInterface dynamic.NamespaceableResourceInterface
	restClient rest.Interface
}

func GVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    typesv1alpha1.GroupName,
		Version:  typesv1alpha1.GroupVersion,
		Resource: "srlinuxes",
	}
}

func GV() *schema.GroupVersion {
	return &schema.GroupVersion{
		Group:   typesv1alpha1.GroupName,
		Version: typesv1alpha1.GroupVersion,
	}
}

// NewForConfig returns a new Clientset based on c.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: GVR().Group, Version: GVR().Version}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	dClient, err := dynamic.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	dInterface := dClient.Resource(GVR())

	rClient, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Clientset{
		dInterface: dInterface,
		restClient: rClient,
	}, nil
}

// Srlinux initializes srlinuxClient struct which implements SrlinuxInterface.
func (c *Clientset) Srlinux(namespace string) SrlinuxInterface {
	return &srlinuxClient{
		dInterface: c.dInterface,
		restClient: c.restClient,
		ns:         namespace,
	}
}

type srlinuxClient struct {
	dInterface dynamic.NamespaceableResourceInterface
	restClient rest.Interface
	ns         string
}

// List gets a list of SRLinux resources.
func (s *srlinuxClient) List(
	ctx context.Context,
	opts metav1.ListOptions, // skipcq: CRT-P0003
) (*typesv1alpha1.SrlinuxList, error) {
	result := typesv1alpha1.SrlinuxList{}
	err := s.restClient.
		Get().
		Namespace(s.ns).
		Resource(GVR().Resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Get gets SRLinux resource.
func (s *srlinuxClient) Get(
	ctx context.Context,
	name string,
	opts metav1.GetOptions,
) (*typesv1alpha1.Srlinux, error) {
	result := typesv1alpha1.Srlinux{}
	err := s.restClient.
		Get().
		Namespace(s.ns).
		Resource(GVR().Resource).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// Create creates SRLinux resource.
func (s *srlinuxClient) Create(
	ctx context.Context,
	srlinux *typesv1alpha1.Srlinux,
) (*typesv1alpha1.Srlinux, error) {
	result := typesv1alpha1.Srlinux{}
	err := s.restClient.
		Post().
		Namespace(s.ns).
		Resource(GVR().Resource).
		Body(srlinux).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (s *srlinuxClient) Watch(
	ctx context.Context,
	opts metav1.ListOptions, // skipcq: CRT-P0003
) (watch.Interface, error) {
	opts.Watch = true

	return s.restClient.
		Get().
		Namespace(s.ns).
		Resource(GVR().Resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

func (s *srlinuxClient) Delete(ctx context.Context,
	name string,
	opts metav1.DeleteOptions, // skipcq: CRT-P0003
) error {
	return s.restClient.
		Delete().
		Namespace(s.ns).
		Resource(GVR().Resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Name(name).
		Do(ctx).
		Error()
}

func (s *srlinuxClient) Update(
	ctx context.Context,
	obj *unstructured.Unstructured,
	opts metav1.UpdateOptions,
) (*typesv1alpha1.Srlinux, error) {
	result := typesv1alpha1.Srlinux{}

	obj, err := s.dInterface.Namespace(s.ns).UpdateStatus(ctx, obj, opts)
	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to type assert return to srlinux: %w", ErrUpdateFailed)
	}

	return &result, nil
}

func (s *srlinuxClient) Unstructured(ctx context.Context, name string, opts metav1.GetOptions,
	subresources ...string,
) (*unstructured.Unstructured, error) {
	return s.dInterface.Namespace(s.ns).Get(ctx, name, opts, subresources...)
}

func init() {
	if err := typesv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
}
