package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/go-kit/kit/log/level"
	"github.com/dbsystel/kube-controller-dbsystel-go-common/controller/ingress"
	"github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes"
	k8sflag "github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes/flag"
	opslog "github.com/dbsystel/kube-controller-dbsystel-go-common/log"
	logflag "github.com/dbsystel/kube-controller-dbsystel-go-common/log/flag"
	"github.com/dbsystel/AmazonRoute53-ingress-controller/controller"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app             = kingpin.New(filepath.Base(os.Args[0]), "AmazonRoute53-ingress-controller")
	whitelistPrefix = app.Flag("whitelist-prefix", "Whitelist prefix for route53 records").String()
	whitelistSuffix = app.Flag("whitelist-suffix", "Whitelist suffix for route53 records").String()
	//Here you can define more flags for your application
)

func main() {
	//Define config for logging
	var logcfg opslog.Config
	//Definie if controller runs outside of k8s
	var runOutsideCluster bool
	//Add two additional flags to application for logging and decision if inside or outside k8s
	logflag.AddFlags(app, &logcfg)
	k8sflag.AddFlags(app, &runOutsideCluster)
	//Parse all arguments
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		//Received error while parsing arguments from function app.Parse
		fmt.Fprintln(os.Stderr, "Catched the following error while parsing arguments: ", err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	//Initialize new logger from opslog
	logger, err := opslog.New(logcfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	//First usage of initialized logger for testing
	level.Debug(logger).Log("msg", "Logging initiated...")
	//Initialize new k8s client from common k8s package
	k8sClient, err := kubernetes.NewClientSet(runOutsideCluster)
	if err != nil {
		level.Error(logger).Log("msg", err.Error())
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	level.Info(logger).Log("msg", "Starting AmazonRoute53-ingress-controller...")
	sigs := make(chan os.Signal, 1) // Create channel to receive OS signals
	stop := make(chan struct{})     // Create channel to receive stop signal

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the sigs channel to receieve SIGTERM

	wg := &sync.WaitGroup{} // Goroutines can add themselves to this to be waited on so that they finish

	//Initialize new k8s ingress-controller from common k8s package
	ingressController := &ingress.IngressController{}
	ingressController.Controller = controller.New(logger, *whitelistPrefix, *whitelistSuffix)
	ingressController.Initialize(k8sClient)
	//Run initiated ingress-controller as go routine
	go ingressController.Run(stop, wg)

	<-sigs // Wait for signals (this hangs until a signal arrives)

	level.Info(logger).Log("msg", "Shutting down...")

	close(stop) // Tell goroutines to stop themselves
	wg.Wait()   // Wait for all to be stopped
}
