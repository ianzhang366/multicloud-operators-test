package apptest

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"path/filepath"

	"reflect"

	appv1alpha1 "github.ibm.com/steve-kim-ibm/multicloud-operators-test/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_apptest")

// Add creates a new AppTest Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	return &ReconcileAppTest{client: mgr.GetClient(), scheme: mgr.GetScheme()}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("apptest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AppTest
	err = c.Watch(&source.Kind{Type: &appv1alpha1.AppTest{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner AppTest
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.AppTest{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAppTest implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAppTest{}

// ReconcileAppTest reconciles a AppTest object
type ReconcileAppTest struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// clientsCache is a cluster name to client mapping used as a cache for connecting to managed clusters
var clientsCache = map[string]client.Client{}

// Reconcile reads that state of the cluster for a AppTest object and makes changes based on the state read
// and what is in the AppTest.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppTest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppTest")

	// Fetch the AppTest instance
	appTest := &appv1alpha1.AppTest{}
	err := r.client.Get(context.TODO(), request.NamespacedName, appTest)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	failedRes := []appv1alpha1.AppTestResources{}
	testStatus := "Success"
	// Check status of the defined resources
	for _, resource := range appTest.Spec.Resources {
		clusterName := resource.Cluster
		if clusterName == "/" {
			clientsCache[clusterName] = r.client
		}

		// Check if client is in cache; if not, check volume to populate the cache
		_, exists := clientsCache[clusterName]
		if !exists {
			reqLogger.Info("Cluster client not in cache")
			configPath := filepath.Join("/etc/config", clusterName)

			reqLogger.Info("Creating new config")
			cfg, err := clientcmd.BuildConfigFromFlags("", configPath)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Creating new client")

			currClient, err := client.New(cfg, client.Options{})
			if err != nil {
				return reconcile.Result{}, err
			}

			clientsCache[clusterName] = currClient
			msg := "Clients Cache updated"
			reqLogger.Info(msg)
		}

		msg := "Current Cluster: " + clusterName
		reqLogger.Info(msg)

		gvk := resource.GroupVersionKind()
		uObj := &unstructured.Unstructured{}
		uObj.SetGroupVersionKind(gvk)

		err = clientsCache[clusterName].Get(context.Background(), client.ObjectKey{
			Namespace: resource.ObjectMeta.Namespace,
			Name:      resource.ObjectMeta.Name,
		}, uObj)

		messages := []string{}

		if err != nil {
			m := err.Error()
			messages := append(messages, m)
			testStatus = "Failed"
			update := appv1alpha1.AppTestResources{
				ObjectMeta: resource.ObjectMeta,
				Cluster:    clusterName,
				Messages:   messages,
			}
			failedRes = append(failedRes, update)
			continue
		}

		// Go through the DesiredStatus to pick out the user defined fields we want to use to compare with the
		// current status
		desired := resource.DesiredStatus
		currentStatus := make(map[string]interface{})
		objStatus, ok := uObj.Object["status"].(map[string]interface{})
		if !ok {
			reqLogger.Info("Type Assertion not okay")
		}

		for key := range desired {
			objVal, exists := objStatus[key]
			if exists {
				currentStatus[key] = objVal
			} else {
				message := key + "is not a valid field under Status of kind " + gvk.Kind
				messages = append(messages, message)
			}
		}

		reqLogger.Info("Comparing currentStatus and desiredStatus")

		success := true
		if !reflect.DeepEqual(currentStatus, desired) {
			success = false
			m := "currentStatus and desiredStatus does not match."
			messages = append(messages, m)
			testStatus = "Failed"
		}

		update := appv1alpha1.AppTestResources{
			ObjectMeta:    resource.ObjectMeta,
			Cluster:       clusterName,
			DesiredStatus: desired,
			CurrentStatus: currentStatus,
			Messages:      messages,
		}

		if success {
			reqLogger.Info("Test successful")
		} else {
			reqLogger.Info("Test failed")
			failedRes = append(failedRes, update)
		}
	}

	reqLogger.Info("Updating resource status")
	status := appv1alpha1.AppTestStatus{
		TestStatus:      testStatus,
		FailedResources: failedRes,
	}

	appTest.Status = status
	err = r.client.Status().Update(context.TODO(), appTest)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}
