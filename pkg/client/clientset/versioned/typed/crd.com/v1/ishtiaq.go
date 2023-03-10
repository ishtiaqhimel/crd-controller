/*
Copyright Ishtiaq Islam.

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

package v1

import (
	"context"
	"time"

	v1 "github.com/ishtiaqhimel/crd-controller/pkg/apis/crd.com/v1"
	scheme "github.com/ishtiaqhimel/crd-controller/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// IshtiaqsGetter has a method to return a IshtiaqInterface.
// A group's client should implement this interface.
type IshtiaqsGetter interface {
	Ishtiaqs(namespace string) IshtiaqInterface
}

// IshtiaqInterface has methods to work with Ishtiaq resources.
type IshtiaqInterface interface {
	Create(ctx context.Context, ishtiaq *v1.Ishtiaq, opts metav1.CreateOptions) (*v1.Ishtiaq, error)
	Update(ctx context.Context, ishtiaq *v1.Ishtiaq, opts metav1.UpdateOptions) (*v1.Ishtiaq, error)
	UpdateStatus(ctx context.Context, ishtiaq *v1.Ishtiaq, opts metav1.UpdateOptions) (*v1.Ishtiaq, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Ishtiaq, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.IshtiaqList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Ishtiaq, err error)
	IshtiaqExpansion
}

// ishtiaqs implements IshtiaqInterface
type ishtiaqs struct {
	client rest.Interface
	ns     string
}

// newIshtiaqs returns a Ishtiaqs
func newIshtiaqs(c *CrdV1Client, namespace string) *ishtiaqs {
	return &ishtiaqs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the ishtiaq, and returns the corresponding ishtiaq object, and an error if there is any.
func (c *ishtiaqs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Ishtiaq, err error) {
	result = &v1.Ishtiaq{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ishtiaqs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Ishtiaqs that match those selectors.
func (c *ishtiaqs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.IshtiaqList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.IshtiaqList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ishtiaqs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested ishtiaqs.
func (c *ishtiaqs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("ishtiaqs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a ishtiaq and creates it.  Returns the server's representation of the ishtiaq, and an error, if there is any.
func (c *ishtiaqs) Create(ctx context.Context, ishtiaq *v1.Ishtiaq, opts metav1.CreateOptions) (result *v1.Ishtiaq, err error) {
	result = &v1.Ishtiaq{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("ishtiaqs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ishtiaq).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a ishtiaq and updates it. Returns the server's representation of the ishtiaq, and an error, if there is any.
func (c *ishtiaqs) Update(ctx context.Context, ishtiaq *v1.Ishtiaq, opts metav1.UpdateOptions) (result *v1.Ishtiaq, err error) {
	result = &v1.Ishtiaq{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ishtiaqs").
		Name(ishtiaq.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ishtiaq).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *ishtiaqs) UpdateStatus(ctx context.Context, ishtiaq *v1.Ishtiaq, opts metav1.UpdateOptions) (result *v1.Ishtiaq, err error) {
	result = &v1.Ishtiaq{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ishtiaqs").
		Name(ishtiaq.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ishtiaq).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the ishtiaq and deletes it. Returns an error if one occurs.
func (c *ishtiaqs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ishtiaqs").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *ishtiaqs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ishtiaqs").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched ishtiaq.
func (c *ishtiaqs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Ishtiaq, err error) {
	result = &v1.Ishtiaq{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("ishtiaqs").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
