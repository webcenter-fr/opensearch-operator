package controllers

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

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
	"software.sslmate.com/src/go-pkcs12"
)

const (
	OpensearchTransportTlsCondition = "OpensearchTransportTls"
	OpensearchTransportTlsPhase     = "Generate transport TLS"
)

type OpensearchTransportTlsReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	name   string
}

// Configure permit to init condition
func (r *OpensearchTransportTlsReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)

	// Init condition status if not exist
	if condition.FindStatusCondition(opensearch.Status.Conditions, OpensearchTransportTlsCondition) == nil {
		condition.SetStatusCondition(&opensearch.Status.Conditions, metav1.Condition{
			Type:   OpensearchTransportTlsCondition,
			Status: metav1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	return nil, nil
}

// Read existing transport TLS
func (r *OpensearchTransportTlsReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)
	s := &corev1.Secret{}
	var rootCA *goca.CA
	var adminCrt *x509.Certificate
	nodeCertificates := map[string]*x509.Certificate{}

	// Read existing secret
	if err = r.Client.Get(ctx, types.NamespacedName{Namespace: opensearch.Namespace, Name: r.GetName(opensearch.Name)}, s); err != nil {
		if !k8serrors.IsNotFound(err) {
			return res, errors.Wrapf(err, "Error when read existing secret %s", r.GetName(opensearch.Name))
		}
		s = nil
	}

	if s != nil {

		// Load root CA
		rootCA, err = pki.LoadRootCATransport(s.Data["ca.key"], s.Data["ca.pub"], s.Data["ca.crt"], s.Data["ca.crl"], r.log)
		if err != nil {
			return res, errors.Wrap(err, "Error when load PKI for transport layout")
		}

		// Load admin certificate
		adminCrt, err = cert.LoadCertFromPem(s.Data["admin.crt"])
		if err != nil {
			return res, errors.Wrap(err, "Error when load admin certificate")
		}

		// Load node certificates
		for _, nodeGroup := range opensearch.Spec.NodeGroups {
			for i := 0; i < int(nodeGroup.Replicas); i++ {
				nodeName := GetNodeName(opensearch.Name, nodeGroup.Name, i)
				if s.Data[fmt.Sprintf("%s.crt", nodeName)] != nil {
					nodeCrt, err := cert.LoadCertFromPem(s.Data[fmt.Sprintf("%s.crt", nodeName)])
					if err != nil {
						return res, errors.Wrapf(err, "Error when load node certificate %s", nodeName)
					}
					nodeCertificates[nodeName] = nodeCrt
				} else {
					nodeCertificates[nodeName] = nil
				}
			}
		}
	}

	data["rootCA"] = rootCA
	data["adminCertificate"] = adminCrt
	data["nodeCertificates"] = nodeCertificates
	data["currentSecret"] = s

	return res, nil
}

// Create generate new TLS authorities
func (r *OpensearchTransportTlsReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedSecret")
	if err != nil {
		return res, err
	}
	expectedSecret := d.(*corev1.Secret)

	if err = r.Client.Update(ctx, expectedSecret); err != nil {
		return res, errors.Wrapf(err, "Error when create secret %s for TLS transport", expectedSecret.Name)
	}

	return res, nil
}

// Update permit to update TLS secret
func (r *OpensearchTransportTlsReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedSecret")
	if err != nil {
		return res, err
	}
	expectedSecret := d.(*corev1.Secret)

	if err = r.Client.Update(ctx, expectedSecret); err != nil {
		return res, errors.Wrapf(err, "Error when update secret %s for TLS transport", expectedSecret.Name)
	}

	return res, nil
}

// Delete permit to delete TLS secret
// We add parent link, so k8s auto delete children
func (r *OpensearchTransportTlsReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	// Update metrics
	controllerMetrics.WithLabelValues(r.name).Dec()

	return nil
}

// Diff permit to check if TLS secret is up to date
func (r *OpensearchTransportTlsReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)
	var d any
	var sb strings.Builder

	d, err = helper.Get(data, "currentSecret")
	if err != nil {
		return diff, err
	}
	currentSecret := d.(*corev1.Secret)

	d, err = helper.Get(data, "rootCA")
	if err != nil {
		return diff, err
	}
	rootCA := d.(*goca.CA)

	d, err = helper.Get(data, "adminCertificate")
	if err != nil {
		return diff, err
	}
	adminCrt := d.(*x509.Certificate)

	d, err = helper.Get(data, "nodeCertificates")
	if err != nil {
		return diff, err
	}
	nodeCertificates := d.(map[string]*x509.Certificate)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	// Create new secret
	if currentSecret == nil {
		diff.NeedCreate = true
		diff.Diff = "Secret not exist"

		expectedSecret, err := r.generateSecret(opensearch)
		if err != nil {
			return diff, errors.Wrapf(err, "Error when generate secret %s for TLS transport", r.GetName(opensearch.Name))
		}
		data["expectedSecret"] = expectedSecret

		r.log.Info("Create PKI for transport layer")

		return diff, nil
	}

	// Handle existing secret
	// Check if CA need to be renewed
	needRenew, err := pki.NeedRenewCertificate(rootCA.GoCertificate(), r.log)
	if err != nil {
		return diff, errors.Wrap(err, "Error when check is CA need to be renewed")
	}
	if needRenew {
		expectedSecret, err := r.renewSecret(opensearch, currentSecret.Data["admin.pfx"])
		if err != nil {
			return diff, errors.Wrapf(err, "Error when generate secret %s for TLS transport", r.GetName(opensearch.Name))
		}
		data["expectedSecret"] = expectedSecret
		diff.NeedUpdate = true
		diff.Diff = "CA root will expire. Renew all certificates"

		r.log.Info("Renew PKI for transport layout")

		return diff, nil
	}

	// Check if node certificates must be renewed or created
	for nodeName, nodeCrt := range nodeCertificates {
		if nodeCrt == nil {
			nc, err := pki.NewNodeTLS(nodeName, rootCA, r.log)
			if err != nil {
				return diff, errors.Wrapf(err, "Error when create node certificate %s", nodeName)
			}
			currentSecret.Data[fmt.Sprintf("%s.crt", nodeName)] = []byte(nc.Certificate)
			currentSecret.Data[fmt.Sprintf("%s.key", nodeName)] = []byte(nc.PrivateKey)
			currentSecret.Data[fmt.Sprintf("%s.csr", nodeName)] = []byte(nc.CSR)
			sb.WriteString(fmt.Sprintf("Create node certificate %s\n", nodeName))
			diff.NeedUpdate = true
			continue
		}

		needRenew, err := pki.NeedRenewCertificate(nodeCrt, r.log)
		if err != nil {
			return diff, errors.Wrapf(err, "Error when check if node certificate %s need to be renewed", nodeName)
		}
		if needRenew {
			// If one expired, we regenerate all certificates to be simple
			expectedSecret, err := r.renewSecret(opensearch, currentSecret.Data["admin.pfx"])
			if err != nil {
				return diff, errors.Wrapf(err, "Error when generate secret %s for TLS transport", r.GetName(opensearch.Name))
			}
			data["expectedSecret"] = expectedSecret
			diff.NeedUpdate = true
			diff.Diff = "One node certificate expired. Renew all certificates"

			r.log.Info("Renew PKI for transport layout")

			return diff, nil
		}
	}

	// Check if admin certificate need to be renewed
	needRenew, err = pki.NeedRenewCertificate(adminCrt, r.log)
	if err != nil {
		return diff, errors.Wrap(err, "Error when check if admin certificate need to be renewed")
	}
	if needRenew {
		ac, err := pki.NewAdminCertificate(rootCA, r.log)
		if err != nil {
			return diff, errors.Wrap(err, "Error when renew admin certificate")
		}
		currentSecret.Data["admin.crt"] = []byte(ac.Certificate)
		currentSecret.Data["admin.key"] = []byte(ac.PrivateKey)
		currentSecret.Data["admin.csr"] = []byte(ac.CSR)
		pkcs12, err := goca.GeneratePkcs12(ac, "")
		if err != nil {
			return diff, errors.Wrap(err, "Error when generate Pkcs12 for admin")
		}
		currentSecret.Data["admin.pfx"] = pkcs12
		diff.NeedUpdate = true
		sb.WriteString("Renew admin certificate\n")
	}

	if diff.NeedUpdate {
		data["expectedSecret"] = currentSecret
		diff.Diff = sb.String()
	}

	return diff, nil
}

// OnError permit to set status condition on the right state and record error
func (r *OpensearchTransportTlsReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {
	opensearch := resource.(*opensearchapi.Opensearch)

	r.log.Error(err)
	r.recorder.Event(resource, corev1.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&opensearch.Status.Conditions, metav1.Condition{
		Type:    OpensearchTransportTlsCondition,
		Status:  metav1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})

	// Update metrics
	totalErrors.Inc()
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *OpensearchTransportTlsReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {
	opensearch := resource.(*opensearchapi.Opensearch)

	if diff.NeedCreate {
		r.recorder.Eventf(resource, corev1.EventTypeNormal, "Completed", "Secret %s successfully created", r.GetName(opensearch.Name))
	}

	if diff.NeedUpdate {
		r.recorder.Eventf(resource, corev1.EventTypeNormal, "Completed", "Secret %s successfully updated", r.GetName(opensearch.Name))
	}

	// Update condition status if needed
	if !condition.IsStatusConditionPresentAndEqual(opensearch.Status.Conditions, OpensearchTransportTlsCondition, metav1.ConditionTrue) {
		condition.SetStatusCondition(&opensearch.Status.Conditions, metav1.Condition{
			Type:    OpensearchTransportTlsCondition,
			Reason:  "Success",
			Status:  metav1.ConditionTrue,
			Message: fmt.Sprintf("Secret %s update to date", r.GetName(opensearch.Name)),
		})
	}

	return nil
}

// GetName permit to retrun secret name that store all TLS for transport layer
func (r *OpensearchTransportTlsReconciler) GetName(clusterName string) string {
	return fmt.Sprintf("opensearch-%s-transport-tls", clusterName)
}

// generateSecret generate the secret with all certificate needed by transport layout
func (r *OpensearchTransportTlsReconciler) generateSecret(opensearch *opensearchapi.Opensearch) (secret *corev1.Secret, err error) {
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.GetName(opensearch.Name),
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
	rootCA, err := pki.NewRootCATransport(r.log)
	if err != nil {
		return nil, errors.Wrap(err, "Error when create transport PKI")
	}
	secret.Data["ca.crt"] = []byte(rootCA.GetCertificate())
	secret.Data["ca.key"] = []byte(rootCA.GetPrivateKey())
	secret.Data["ca.pub"] = []byte(rootCA.GetPublicKey())
	secret.Data["ca.crl"] = []byte(rootCA.GetCRL())

	// Genereate admin cert
	adminCrt, err := pki.NewAdminCertificate(rootCA, r.log)
	if err != nil {
		return nil, errors.Wrap(err, "Error when generate admin certificate")
	}
	secret.Data["admin.crt"] = []byte(adminCrt.Certificate)
	secret.Data["admin.key"] = []byte(adminCrt.PrivateKey)
	secret.Data["admin.csr"] = []byte(adminCrt.CSR)
	pkcs12, err := goca.GeneratePkcs12(adminCrt, "")
	if err != nil {
		return nil, errors.Wrap(err, "Error when generate Pkcs12 for admin")
	}
	secret.Data["admin.pfx"] = pkcs12

	// Generate nodes certificates
	for _, nodeGroup := range opensearch.Spec.NodeGroups {
		for i := 0; i < int(nodeGroup.Replicas); i++ {
			nodeName := GetNodeName(opensearch.Name, nodeGroup.Name, i)
			nodeCrt, err := pki.NewNodeTLS(nodeName, rootCA, r.log)
			if err != nil {
				return nil, errors.Wrapf(err, "Error when generate node certificate for %s", nodeName)
			}
			secret.Data[fmt.Sprintf("%s.crt", nodeName)] = []byte(nodeCrt.Certificate)
			secret.Data[fmt.Sprintf("%s.key", nodeName)] = []byte(nodeCrt.PrivateKey)
			secret.Data[fmt.Sprintf("%s.csr", nodeName)] = []byte(nodeCrt.CSR)
			pkcs12, err := goca.GeneratePkcs12(nodeCrt, "")
			if err != nil {
				return nil, errors.Wrapf(err, "Error when generate Pkcs12 for node %s", nodeName)
			}
			secret.Data[fmt.Sprintf("%s.pfx", nodeName)] = pkcs12
		}
	}

	return secret, nil
}

// renewSecret regenerate all certificate needed for transport layout
// It add all valid ca certificate previously used to rolling upgrade node
func (r *OpensearchTransportTlsReconciler) renewSecret(opensearch *opensearchapi.Opensearch, oldPfx []byte) (secret *corev1.Secret, err error) {
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.GetName(opensearch.Name),
			Namespace: opensearch.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Set owner
	err = ctrl.SetControllerReference(opensearch, secret, r.Scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when set as owner reference")
	}

	oldCaCerts := make([]*x509.Certificate, 0)
	_, _, caCerts, err := pkcs12.DecodeChain(oldPfx, "")
	if err != nil {
		return nil, errors.Wrapf(err, "Error when decode old Pkcs12")
	}
	// Keep only not expired certs
	for _, crt := range caCerts {
		if crt.NotAfter.After(time.Now()) {
			oldCaCerts = append(oldCaCerts, crt)
		}
	}

	// Generate new PKI
	rootCA, err := pki.NewRootCATransport(r.log)
	if err != nil {
		return nil, errors.Wrap(err, "Error when create transport PKI")
	}
	secret.Data["ca.crt"] = []byte(rootCA.GetCertificate())
	secret.Data["ca.key"] = []byte(rootCA.GetPrivateKey())
	secret.Data["ca.pub"] = []byte(rootCA.GetPublicKey())
	secret.Data["ca.crl"] = []byte(rootCA.GetCRL())

	// Genereate admin cert
	adminCrt, err := pki.NewAdminCertificate(rootCA, r.log)
	if err != nil {
		return nil, errors.Wrap(err, "Error when generate admin certificate")
	}
	secret.Data["admin.crt"] = []byte(adminCrt.Certificate)
	secret.Data["admin.key"] = []byte(adminCrt.PrivateKey)
	secret.Data["admin.csr"] = []byte(adminCrt.CSR)
	pkcs12, err := goca.GeneratePkcs12(adminCrt, "", oldCaCerts...)
	if err != nil {
		return nil, errors.Wrap(err, "Error when generate Pkcs12 for admin")
	}
	secret.Data["admin.pfx"] = pkcs12

	// Generate nodes certificates
	for _, nodeGroup := range opensearch.Spec.NodeGroups {
		for i := 0; i < int(nodeGroup.Replicas); i++ {
			nodeName := GetNodeName(opensearch.Name, nodeGroup.Name, i)
			nodeCrt, err := pki.NewNodeTLS(nodeName, rootCA, r.log)
			if err != nil {
				return nil, errors.Wrapf(err, "Error when generate node certificate for %s", nodeName)
			}
			secret.Data[fmt.Sprintf("%s.crt", nodeName)] = []byte(nodeCrt.Certificate)
			secret.Data[fmt.Sprintf("%s.key", nodeName)] = []byte(nodeCrt.PrivateKey)
			secret.Data[fmt.Sprintf("%s.csr", nodeName)] = []byte(nodeCrt.CSR)
			pkcs12, err := goca.GeneratePkcs12(nodeCrt, "", oldCaCerts...)
			if err != nil {
				return nil, errors.Wrapf(err, "Error when generate Pkcs12 for node %s", nodeName)
			}
			secret.Data[fmt.Sprintf("%s.pfx", nodeName)] = pkcs12
		}
	}

	return secret, nil
}
