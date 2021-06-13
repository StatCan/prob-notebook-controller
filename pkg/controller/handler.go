package controller

import (
	"context"
	"fmt"
	"reflect"

	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	"github.com/prometheus/common/log"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
const (
	probBlock = "prob-block"
)

func notebookAuthPolicyName(notebook *v1.Notebook) string {
	return fmt.Sprintf("%s-%s", notebook.Name, probBlock)
}

func (c *Controller) handleAuthorizationPolicy(notebook *v1.Notebook) error {
	ctx := context.Background()
	
	// //Find any authorization policy with the same name
	ap, err := c.authorizationPoliciesListers.AuthorizationPolicies(notebook.Namespace).Get(notebookAuthPolicyName(notebook))
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	//Check that we own this authorization policy
	if ap != nil {
		if !metav1.IsControlledBy(ap, notebook) {
			msg := fmt.Sprintf("Authorization Policy \"%s/%s\" already exists and is not managed by the notebook", ap.Namespace, ap.Name)
			c.recorder.Event(notebook, corev1.EventTypeWarning, "ErrResourceExists", msg)
			return fmt.Errorf("%s", msg)
		}
	}

	//New Authorization Policy
	nap, err := c.generateAuthorizationPolicy(notebook)
	if err != nil {
		return err
	}

	// If we don't have authorization policy, then let's make one
	if ap == nil {
		_, err = c.istioClientset.SecurityV1beta1().AuthorizationPolicies(notebook.Namespace).Create(ctx, nap, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !reflect.DeepEqual(ap.Spec, nap.Spec) { //We have an authorization policy, but is it the same
		log.Infof("updated authorization Policy \"%s/%s\"", ap.Namespace, ap.Name)

		// Copy the new spec
		ap.Spec = nap.Spec

		_, err = c.istioClientset.SecurityV1beta1().AuthorizationPolicies(notebook.Namespace).Update(ctx, ap, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}


func (c *Controller) generateAuthorizationPolicy(notebook *v1.Notebook)(*istiosecurityv1beta1.AuthorizationPolicy, error){
	ap := &istiosecurityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: notebookAuthPolicyName(notebook),
			Namespace: notebook.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(notebook, v1.SchemeGroupVersion.WithKind("Notebook")),
			},
		},
		Spec: v1beta1.AuthorizationPolicy{
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: map[string]string{
					"notebook-name": notebook.Name,
					"data.statcan.gc.ca/classification": "protected-b",
				},
			},
			Action: v1beta1.AuthorizationPolicy_DENY,
			Rules: []*v1beta1.Rule{
				{
					To: []*v1beta1.Rule_To{
						{
							Operation: &v1beta1.Operation{
								Methods: []string{"POST"},
								Paths: []string{fmt.Sprintf("/notebook/%s/%s/rstudio/upload", notebook.Namespace, notebook.Name)},
							},
						},
						{
							Operation: &v1beta1.Operation{
								Methods: []string{"GET"},
								Paths: []string{fmt.Sprintf("/notebook/%s/%s/rstudio/export*", notebook.Namespace, notebook.Name)},
							},
						},
						{
							Operation: &v1beta1.Operation{
								Methods: []string{"GET"},
								Paths: []string{fmt.Sprintf("/notebook/%s/%s/files*", notebook.Namespace, notebook.Name)},
							},
						},
					},
				},
			},
		},
	}

	return ap, nil
}
