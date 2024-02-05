package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restK8s "k8s.io/client-go/rest"
)

func KubernetesAutoConfig(replicaCount int, statefulsetName string, port int) ([]string, []string, uint64) {
	// creates the in-cluster config
	config, err := restK8s.InClusterConfig()
	if err != nil {
		log.Fatal().Msg("Not running inside Kubernetes")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Msg("Unable to create config inside Kubernetes")
	}

	// Read the entire file into a byte slice
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to get current namespace")
	}

	// Convert the byte slice to a string
	namespace := string(data)
	log.Info().Str("namespace", namespace).Msg("Running in Kubernetes")

	// Get StatefulSet
	statefulSets := clientset.AppsV1().StatefulSets(namespace)
	statefulSet, err := statefulSets.Get(context.Background(), statefulsetName, v1.GetOptions{})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to get statefulsets")
	}

	selector := v1.FormatLabelSelector(statefulSet.Spec.Selector)

	// Get pods
	addresses := make([]string, replicaCount-1)
	ids := make([]string, replicaCount-1)
	pID := uint64(0)

	for {
		pods, err := clientset.CoreV1().Pods(statefulSet.Namespace).List(context.Background(), v1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to get pods")
		}

		if len(pods.Items) != replicaCount {
			log.Info().Msg("Waiting while all pods are not ready...")
			time.Sleep(time.Second * 5)
			continue
		}

		services, err := clientset.CoreV1().Services(statefulSet.Namespace).List(context.Background(), v1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to get services")
		}

		// Print pod names

		k := 0

		podID, err := extractNumber(os.Getenv("HOSTNAME"))
		if err != nil {
			log.Fatal().Err(err).Msg("unable to get index")
		}

		pID = podID

		for i, service := range services.Items {
			if i == int(podID) {
				continue
			}

			addresses[k] = service.Name + "." + namespace + ".svc.cluster.local:" + strconv.Itoa(port)
			ids[k] = strconv.Itoa(i)
			k++
		}

		break
	}

	return addresses, ids, pID
}

func extractNumber(s string) (uint64, error) {
	// Split the string using hyphens as separators
	parts := strings.Split(s, "-")

	// Check if there are any parts in the split result
	if len(parts) > 0 {
		// Get the last part, which should be the number
		lastPart := parts[len(parts)-1]

		// Convert the last part to a uint64
		number, err := strconv.ParseUint(lastPart, 10, 64)
		if err != nil {
			return 0, err
		}

		return number, nil
	}

	// No parts found
	return 0, fmt.Errorf("no number found in the string")
}
