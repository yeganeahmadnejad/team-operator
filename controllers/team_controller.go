/*
Copyright 2022.

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

package controllers

import (
	"context"

	//"text/template"

	teamv1 "github.com/yeganeahmadnejad/team-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	//"k8s.io/client-go/tools/clientcmd"
	"encoding/json"
	//"github.com/argoproj/argo-cd/pkg/apiclient"
	// sessionpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	//"github.com/argoproj/argo-cd/v2/pkg/apiclient/account"
	b64 "encoding/base64"

	"bytes"
	"io/ioutil"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	//"time"
)

//clientOpts.ServerAddr = fmt.Sprintf("%s:%d", *address, *port)
//clientOpts.PlainText = true

//apiClient, err := argoclient.NewClient(clientOpts)

var logf = log.Log.WithName("controller_team")

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}
type account struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

//+kubebuilder:rbac:groups=team.snappcloud.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=team.snappcloud.io,resources=teams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=team.snappcloud.io,resources=teams/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete


// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the clus k8s.io/api closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Team object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *TeamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//ArgoCDClientset, _ := apiclient.NewClient(&apiclient.ClientOptions{Insecure: true, ServerAddr: "argocd.apps.private.okd4.ts-2.staging-snappcloud.io", PlainText: true})
	//ArgoCDClientset.UpdatePassword
	//ArgoCDClientset.UpdatePassword(ctx, &account.UpdatePasswordRequest{CurrentPassword: "oldpassword", NewPassword: "newpassword"})

	log := log.FromContext(ctx)
	reqLogger := logf.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling team")
	team := &teamv1.Team{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, team)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	} else {
		log.Info("team is found and teamAdmin is : " + team.Spec.TeamAdmin)

	}
	//	cm := &corev1.ConfigMap{}

	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-cm",
			Namespace: "argocd",
		},
	}
	log.Info("cm is found and cm is : " + cm.Name)
	//r.Client.create(ctx, cm)
	// staticUser := map[string]string{
	// 	"accounts." + team.Spec.TeamAdmin: "apiKey,login",
	// }
	// // bytes, _ := json.Marshal(staticUser)
	// type patch struct {
	// 	data map[string]string `json:"data"`
	// }
	// test := patch{
	// 	data: staticUser,
	// }
	// bytes, _ := json.Marshal(test)
	// log.Info("staticUser is  : " + staticUser[])

	// log.Info("byte is  : " + string(bytes))

	// staticUser2 := map[string]interface{}{
	// 	"data": map[string]string{
	// 		"accounts."+user : "apiKey,login",
	// 		"eagle":  "squak",
	// 	},
	// 	"total birds": 2,
	// }
	// data2, _ := json.Marshal(staticUser2)
	// fmt.Println(string(data2))
	log.Info("testtt")

	staticUser := map[string]map[string]string{
		"data": {
			"accounts." + team.Spec.Argo.Tokens.ArgocdUser: "apiKey,login",
		},
	}

	staticUserByte, _ := json.Marshal(staticUser)
	log.Info(string(staticUserByte))

	//patch := []byte(`{"data":{"accounts.$team.Spec.TeamAdmin" : "apiKey,login"}}`)
	//patchbyte := []byte(bytes)
	// type myStruct struct {
	//     Name        string    `json:"name"`
	//     SaveConfig  string    `json:"saveconfig"`
	// }

	err = r.Client.Patch(context.Background(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "argocd",
			Name:      "argocd-cm",
		},
	}, client.RawPatch(types.StrategicMergePatchType, staticUserByte))
	if err != nil {
		log.Error(err, "Failed to patch   cm")
		return ctrl.Result{}, err
	}
	cm1 := &corev1.ConfigMap{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: "argocd-cm", Namespace: "argocd"}, cm1)
	if err != nil {
		log.Error(err, "Failed to get  cm")
		return ctrl.Result{}, err
	}
	/**/
	hash, _ := HashPassword(team.Spec.Argo.Tokens.ArgocdPass) // ignore error for the sake of simplicity

	log.Info("Password:" + team.Spec.TeamAdmin)
	log.Info("Hash:    " + hash)

	// match := CheckPasswordHash(team.Spec.TeamAdmin, hash)
	//log.Info(match)

	sEnc := b64.StdEncoding.EncodeToString([]byte(hash))
	log.Info(sEnc)
	staticPassword := map[string]map[string]string{
		"data": {
			"accounts." + team.Spec.Argo.Tokens.ArgocdUser + ".password":      sEnc,
			"accounts." + team.Spec.Argo.Tokens.ArgocdUser + ".passwordMtime": "MjAyMi0wNC0zMFQwOTowODo0Mlo=",
			//"accounts." + team.Spec.Argo.Tokens.ArgocdUser + ".tokens":        "bnVsbA==",
		},
	}
	staticPassByte, _ := json.Marshal(staticPassword)
	log.Info(string(staticPassByte))

	err = r.Client.Patch(context.Background(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "argocd",
			Name:      "argocd-secret",
		},
	}, client.RawPatch(types.StrategicMergePatchType, staticPassByte))
	if err != nil {
		log.Error(err, "Failed to patch  secret")
		return ctrl.Result{}, err
	}
	account1 := account{Password: team.Spec.Argo.Tokens.ArgocdPass, Username: team.Spec.Argo.Tokens.ArgocdUser}
	log.Info("account is")
	log.Info(account1.Password + account1.Username)
	log.Info("token is")
//	time.Sleep(8 * time.Second)

	//log.Info(getToken(account1))

	type response struct {
		Token string `json:"token"`
	}

	url := "https://argocd.apps.private.okd4.ts-2.staging-snappcloud.io/api/v1/session"
	// curl https://argocd.apps.private.okd4.ts-2.staging-snappcloud.io/api/v1/session -X POST -d '{"password": "newpassword","username": "admin"}'
	data2, err := json.Marshal(account1)
	//reqLogger := logf.WithValues("Request.Namespace",  "Request.Name")
    log.Info("data2")
	log.Info(string(data2))
	if err != nil {
		log.Error(err, "test")
	}
	// Create a new request using http
	req1, err2 := http.NewRequest("POST", url, bytes.NewBuffer(data2))
	req1.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req1)
	if err2 != nil {
		log.Error(err,"Error on response.\n[ERROR] -")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err,"Error while reading the response bytes:")
	}
	log.Info(string([]byte(body)))
	response1 := response{}
	jsonErr := json.Unmarshal(body, &response1)
	if jsonErr != nil {
		log.Error(jsonErr,"json err")
	}
	log.Info(response1.Token)
	/*
		// cm1 := &corev1.ConfigMap{}
		// err = r.Client.Get(ctx, types.NamespacedName{Name: "my-cm", Namespace: "argocd"}, cm1)
		// if err != nil {
		// 	log.Error(err, "Failed to get   cm")
		// 	return ctrl.Result{}, err
		// }

		/*
			found := &corev1.ConfigMap{}
			err = r.Client.Get(ctx, typ
				es.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				// Define a new deployment
				cm1 := r.configmapfortest(cm)
				log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				err = r.Client.Create(ctx, cm1)
				r.Client.P
				if err != nil {
					log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
					return ctrl.Result{}, err
				}
				// Deployment created successfully - return and requeue
				return ctrl.Result{Requeue: true}, nil
			} else if err != nil {
				log.Error(err, "Failed to get Deployment")
				return ctrl.Result{}, err
			}

	*/

	//log.Info(cm.Data["key"])
	//if _, err := clientset.CoreV1().ConfigMaps("game").Get("game-data", metav1.GetOptions{}); errors.IsNotFound(err) {}
	//coreV1().ConfigMaps("my-namespace")
	//clientset.CoreV1().ConfigMaps("my-namespace").Create(&cm)
	// team := &teamv1.Team{}
	// err := r.Get(ctx , team)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		// Request object not found, could have been deleted after reconcile request.
	// 		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
	// 		// Return and don't requeue
	// 		log.Info("Resource not found. Ignoring since object must be deleted")
	// 		return ctrl.Result{}, nil
	// 	}
	// 	// Error reading the object - requeue the request.
	// 	log.Error(err, "Failed to get team")
	// 	return ctrl.Result{}, err
	// }
	// your logic here
	//log := r.Log.WithValues("Team" )
	//log.Info("team resource not found. Ignoring since object must be deleted")

	// Fetch the Team instance
	//team := &teamv1.Team{}

	// err := r.Get(ctx, req.NamespacedName, team)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		// Request object not found, could have been deleted after reconcile request.
	// 		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
	// 		// Return and don't requeue
	// 		log.Info("team resource not found. Ignoring since object must be deleted")
	// 		return ctrl.Result{}, nil
	// 	}
	// 	// Error reading the object - requeue the request.
	// 	log.Error(err, "Failed to get team")
	// 	return ctrl.Result{}, err
	// }
	return ctrl.Result{}, nil
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&teamv1.Team{}).
		Complete(r)
}

func getToken(account1 account) string {
	type response struct {
		Token string `json:"token"`
	}

	url := "https://argocd.apps.private.okd4.ts-2.staging-snappcloud.io/api/v1/session"
	// curl https://argocd.apps.private.okd4.ts-2.staging-snappcloud.io/api/v1/session -X POST -d '{"password": "newpassword","username": "admin"}'
	data2, err := json.Marshal(account1)
	reqLogger := logf.WithValues("Request.Namespace", "Request.Name")

	reqLogger.Info(string(data2))
	if err != nil {
		reqLogger.Error(err, "test")
	}
	// Create a new request using http
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data2))
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		reqLogger.Info("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Error while reading the response bytes:", err)
	}
	reqLogger.Info(string([]byte(body)))
	response1 := response{}
	jsonErr := json.Unmarshal(body, &response1)
	if jsonErr != nil {
		reqLogger.Error(jsonErr, "test")
	}
	reqLogger.Info(response1.Token)
	return response1.Token

}
