package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Requeue triggers a object requeue.
func Requeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// RequeueOnError triggers requeue when error is not nil.
func RequeueOnError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// RequeueWithError triggers a object requeue because the informed error happened.
func RequeueWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, err
}

// RequeueAfterSeconds triggers a object requeue after the informed seconds.
func RequeueAfterSeconds(seconds int) (ctrl.Result, error) {
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: time.Duration(seconds) * time.Second,
	}, nil
}

// NoRequeue all done, the object does not need reconciliation anymore.
func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: false}, nil
}
