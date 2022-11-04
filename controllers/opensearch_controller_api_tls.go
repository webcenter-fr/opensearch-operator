package controllers


import (
	"context"
	"crypto/x509"
	"fmt"
	
	"github.com/disaster37/goca"
	"github.com/disaster37/goca/cert"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/pkg/errors"
	opensearchapi "github.com/webcenter-fr/opensearch-operator/api/v1alpha1"
	"github.com/webcenter-fr/opensearch-operator/pkg/pki"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OpensearchApiTlsCondition = "OpensearchApiTls"
	OpensearchApiTlsPhase     = "Generate transport TLS"
)

type OpensearchApiTlsReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	name   string
}

// Configure permit to init condition
func (r *OpensearchApiTlsReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)

	// Init condition status if not exist
	if condition.FindStatusCondition(opensearch.Status.Conditions, OpensearchApiTlsCondition) == nil {
		condition.SetStatusCondition(&opensearch.Status.Conditions, metav1.Condition{
			Type:   OpensearchApiTlsCondition,
			Status: metav1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	return nil, nil
}

// Read existing Api TLS
func (r *OpensearchApiTlsReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)
	s := &corev1.Secret{}
	var rootCA *goca.CA
	var apiCrt *x509.Certificate

	// Read existing secret
	secretName := opensearch.GetSecretNameForTlsApi()
	if err = r.Client.Get(ctx, types.NamespacedName{Namespace: opensearch.Namespace, Name: secretName}, s); err != nil {
		if !k8serrors.IsNotFound(err) {
			return res, errors.Wrapf(err, "Error when read existing secret %s", secretName)
		}
		if !opensearch.IsSelfManagedSecretForTlsApi() {
			r.log.Warnf("The secret %s not yet exist, Retry in few time", secretName)
			return ctrl.Result{RequeueAfter: requeuedDuration}, nil
		}
		s = nil
	}

	// Existing secret with self managed
	if s != nil && opensearch.IsSelfManagedSecretForTlsApi() {

		// Load root CA
		rootCA, err = pki.LoadRootCAApi(s.Data["ca.key"], s.Data["ca.pub"], s.Data["ca.crt"], s.Data["ca.crl"], r.log)
		if err != nil {
			return res, errors.Wrap(err, "Error when load PKI for Api layout")
		}

		// Load Api certificate
		apiCrt, err = cert.LoadCertFromPem(s.Data["api.crt"])
		if err != nil {
			return res, errors.Wrap(err, "Error when load Api certificate")
		}

		
		data["rootCA"] = rootCA
		data["apiCertificate"] = apiCrt
	}

	// Existing secret without self managed
	if s != nil && !opensearch.IsSelfManagedSecretForTlsApi() {
		if len(s.Data["key.tls"]) == 0 {
			r.log.Warnf("The secret %s not contend key.tls, Retry in few time", secretName)
			return ctrl.Result{RequeueAfter: requeuedDuration}, nil
		}
		if len(s.Data["tls.crt"]) == 0 {
			r.log.Warnf("The secret %s not contend key.crt, Retry in few time", secretName)
			return ctrl.Result{RequeueAfter: requeuedDuration}, nil
		}
	}

	data["currentSecret"] = s

	return res, nil
}

// Create save secret with new API certificate
func (r *OpensearchApiTlsReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedSecret")
	if err != nil {
		return res, err
	}
	expectedSecret := d.(*corev1.Secret)

	if err = r.Client.Update(ctx, expectedSecret); err != nil {
		return res, errors.Wrapf(err, "Error when create secret %s for Api", expectedSecret.Name)
	}

	return res, nil
}

// Update permit to update TLS secret
func (r *OpensearchApiTlsReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedSecret")
	if err != nil {
		return res, err
	}
	expectedSecret := d.(*corev1.Secret)

	if err = r.Client.Update(ctx, expectedSecret); err != nil {
		return res, errors.Wrapf(err, "Error when update secret %s for Api", expectedSecret.Name)
	}

	return res, nil
}

// Delete permit to delete TLS secret
// We add parent link, so k8s auto delete children
func (r *OpensearchApiTlsReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	// Update metrics
	controllerMetrics.WithLabelValues(r.name).Dec()

	return nil
}

// Diff permit to check if TLS secret is up to date
func (r *OpensearchApiTlsReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)
	var d any
	var rootCA *goca.CA
	var apiCrt *x509.Certificate

	d, err = helper.Get(data, "currentSecret")
	if err != nil {
		return diff, err
	}
	currentSecret := d.(*corev1.Secret)


	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	// Do somethink only if operator manage API certificate
	if opensearch.IsSelfManagedSecretForTlsApi() {

		// Load pki and current certificate
		d, err = helper.Get(data, "rootCA")
		if err != nil {
			return diff, err
		}
		rootCA = d.(*goca.CA)

		d, err = helper.Get(data, "apiCertificate")
		if err != nil {
			return diff, err
		}
		apiCrt = d.(*x509.Certificate)

		// Create new secret if not yet exist
		if currentSecret == nil {
			diff.NeedCreate = true
			diff.Diff = "Secret not exist"
	
			expectedSecret, err := r.generateSecret(opensearch)
			if err != nil {
				return diff, errors.Wrapf(err, "Error when generate secret %s for TLS Api", opensearch.GetSecretNameForTlsApi())
			}
			data["expectedSecret"] = expectedSecret
	
			r.log.Info("Create PKI for Api layer")
	
			return diff, nil
		}

		// Check if CA need to be renewed or Api certificate
		needRenewCA, err := pki.NeedRenewCertificate(rootCA.GoCertificate(), r.log)
		if err != nil {
			return diff, errors.Wrap(err, "Error when check is CA need to be renewed")
		}
		needRenewApiCertificate, err := pki.NeedRenewCertificate(apiCrt, r.log)
		if err != nil {
			return diff, errors.Wrap(err, "Error when check if Api certificate need to be renewed")
		}
		if needRenewCA || needRenewApiCertificate {
			expectedSecret, err := r.generateSecret(opensearch)
			if err != nil {
				return diff, errors.Wrapf(err, "Error when generate secret %s for TLS Api", opensearch.GetSecretNameForTlsApi())
			}
			data["expectedSecret"] = expectedSecret
			diff.NeedUpdate = true
			diff.Diff = "CA root will expire. Renew all certificates"

			r.log.Info("Renew PKI for Api layout")

			return diff, nil
		}
	}

	return diff, nil
}

// OnError permit to set status condition on the right state and record error
func (r *OpensearchApiTlsReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)

	r.log.Error(err)
	r.recorder.Event(resource, corev1.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&opensearch.Status.Conditions, metav1.Condition{
		Type:    OpensearchApiTlsCondition,
		Status:  metav1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})

	// Update metrics
	totalErrors.Inc()
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *OpensearchApiTlsReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {
	opensearch := resource.(*opensearchapi.Opensearch)

	if diff.NeedCreate {
		r.recorder.Eventf(resource, corev1.EventTypeNormal, "Completed", "Secret %s successfully created", opensearch.GetSecretNameForTlsApi())
	}

	if diff.NeedUpdate {
		r.recorder.Eventf(resource, corev1.EventTypeNormal, "Completed", "Secret %s successfully updated", opensearch.GetSecretNameForTlsApi())
	}

	// Update condition status if needed
	if !condition.IsStatusConditionPresentAndEqual(opensearch.Status.Conditions, OpensearchApiTlsCondition, metav1.ConditionTrue) {
		condition.SetStatusCondition(&opensearch.Status.Conditions, metav1.Condition{
			Type:    OpensearchApiTlsCondition,
			Reason:  "Success",
			Status:  metav1.ConditionTrue,
			Message: fmt.Sprintf("Secret %s up to date", opensearch.GetSecretNameForTlsApi()),
		})
	}

	return nil
}



// generateSecret generate the secret with all certificate needed by Api layout
func (r *OpensearchApiTlsReconciler) generateSecret(opensearch *opensearchapi.Opensearch) (secret *corev1.Secret, err error) {
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opensearch.GetSecretNameForTlsApi(),
			Namespace: opensearch.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Set owner
	err = ctrl.SetControllerReference(opensearch, secret, r.Scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when set as owner reference")
	}

	// Generate new PKI
	rootCA, err := pki.NewRootCAApi(r.log)
	if err != nil {
		return nil, errors.Wrap(err, "Error when create Api PKI")
	}
	secret.Data["ca.crt"] = []byte(rootCA.GetCertificate())
	secret.Data["ca.key"] = []byte(rootCA.GetPrivateKey())
	secret.Data["ca.pub"] = []byte(rootCA.GetPublicKey())
	secret.Data["ca.crl"] = []byte(rootCA.GetCRL())

	// Genereate Api cert
	var altnames []string
	var altips []string
	if opensearch.IsLoadBalancerEnabled() && opensearch.Spec.Endpoint.LoadBalancer.Tls != nil && opensearch.Spec.Endpoint.LoadBalancer.Tls.SelfSignedCertificate != nil  {
		altnames = opensearch.Spec.Endpoint.LoadBalancer.Tls.SelfSignedCertificate.AltNames
		altips = opensearch.Spec.Endpoint.LoadBalancer.Tls.SelfSignedCertificate.AltIps
	}
	apiCrt, err := pki.NewApiTls(opensearch.Name, altnames, altips, rootCA, r.log)
	if err != nil {
		return nil, errors.Wrap(err, "Error when generate Api certificate")
	}
	secret.Data["api.crt"] = []byte(apiCrt.Certificate)
	secret.Data["api.key"] = []byte(apiCrt.PrivateKey)
	secret.Data["api.csr"] = []byte(apiCrt.CSR)
	pkcs12, err := goca.GeneratePkcs12(apiCrt, "")
	if err != nil {
		return nil, errors.Wrap(err, "Error when generate Pkcs12 for Api")
	}
	secret.Data["api.pfx"] = pkcs12

	return secret, nil
}
