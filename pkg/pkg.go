package pkg

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type ClusterRole struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Rules []struct {
		APIGroups []string `yaml:"apiGroups"`
		Resources []string `yaml:"resources"`
		Verbs     []string `yaml:"verbs"`
	} `yaml:"rules"`
}

func Execute(email string, clusterServer string, clusterRole string, skipIP bool, deleteSa bool) {
	if email == "" {
		fmt.Println("Please provide an email address (example@mail.com): ")
		fmt.Scanln(&email)
		if email == "" {
			fmt.Println("Error getting email address")
			os.Exit(1)
		}
	}

	// Prompt user for IP if not provided and not skipping
	if clusterServer == "" && !skipIP && !deleteSa {
		fmt.Print("Enter the Cluster IP address (press Enter to skip): ")
		fmt.Scanln(&clusterServer)
	}

	// List available contexts
	fmt.Println("\nAvailable contexts:")
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	contextsOutput, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting available contexts: %v\n", err)
		os.Exit(1)
	}
	contexts := strings.Split(strings.TrimSpace(string(contextsOutput)), "\n")
	for i, context := range contexts {
		fmt.Printf("%d. %s\n", i+1, context)
	}

	// Ask the user to choose a context
	var selectedContextIndex int
	fmt.Print("Choose a context (enter the number): ")
	fmt.Scanln(&selectedContextIndex)
	if selectedContextIndex < 1 || selectedContextIndex > len(contexts) {
		fmt.Println("Invalid context selection.")
		os.Exit(1)
	}
	selectedContext := contexts[selectedContextIndex-1]

	// Get Kubernetes context information
	cmd = exec.Command("kubectl", "config", "current-context")
	contextOutput, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting current context: %v\n", err)
		os.Exit(1)
	}
	currentContext := strings.TrimSpace(string(contextOutput))

	// Check if the selected context matches the current context
	if currentContext != selectedContext {
		// Set the selected context
		cmd = exec.Command("kubectl", "config", "use-context", selectedContext)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error setting selected context: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Switched to context: %s\n", selectedContext)
	}

	// If the IP is empty, retrieve it from the current context
	if clusterServer == "" && !skipIP {
		cmd := exec.Command("kubectl", "config", "view", "--minify", "--output", "jsonpath={.clusters[*].cluster.server}")
		serverOutput, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error getting cluster server: %v\n", err)
			os.Exit(1)
		}
		clusterServer = strings.TrimSpace(string(serverOutput))
	} else {
		clusterServer = "https://" + clusterServer + ":6443"
	}

	// Convert email to the desired format
	serviceAccountName := strings.ReplaceAll(email, "@", "-")
	serviceAccountName = strings.ReplaceAll(serviceAccountName, ".", "-")

	// Create service account
	cmd = exec.Command("kubectl", "delete", "serviceaccount", serviceAccountName, "--namespace=kube-system")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error deleting old service account: %v\n", err)
	}
	if !deleteSa {
		cmd = exec.Command("kubectl", "create", "serviceaccount", serviceAccountName, "--namespace=kube-system")
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error creating service account: %v\n", err)
			os.Exit(1)
		}
	}

	// Create service account token
	var serviceAccountToken []byte
	if !deleteSa {
		cmd = exec.Command("kubectl", "create", "token", serviceAccountName, "--duration=999999h", "--namespace=kube-system")
		serviceAccountToken, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error creating service account token: %v\n", err)
			os.Exit(1)
		}
	}

	// Define the cluster role
	if !deleteSa && clusterRole == "" {
		clusterRoleYAML := fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: %s-cluster-role
rules:
- apiGroups: [""]
  resources:
  - configmaps
  - nodes
  - pods
  - pods/log
  - services
  - resourcequotas
  - replicationcontrollers
  - limitranges
  - persistentvolumeclaims
  - persistentvolumes
  - namespaces
  - endpoints
  - daemonsets
  - deployments
  - replicasets
  - ingresses
  - statefulsets
  - cronjobs
  - jobs
  - horizontalpodautoscalers
  - bindings
  verbs: ["get", "list", "watch"]
`, serviceAccountName)

		err = os.WriteFile("cluster-role.yaml", []byte(clusterRoleYAML), 0644)
		if err != nil {
			fmt.Printf("Error writing cluster role YAML: %v\n", err)
			os.Exit(1)
		}

		// Apply the cluster role
		cmd = exec.Command("kubectl", "apply", "-f", "cluster-role.yaml")
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error applying cluster role: %v\n", err)
			os.Exit(1)
		}
	} else if !deleteSa && clusterRole != "" {
		// Read YAML data from file
		yamlFile, err := os.ReadFile(clusterRole)
		if err != nil {
			log.Fatalf("Error reading cluster role YAML file: %v", err)
		}

		var role ClusterRole

		// Unmarshal YAML data into struct
		err = yaml.Unmarshal(yamlFile, &role)
		if err != nil {
			log.Fatalf("Error unmarshalling cluster role YAML: %v", err)
		}

		// Replace "custom" in the name with "serviceAccountName"
		role.Metadata.Name = strings.Replace(role.Metadata.Name, "custom", serviceAccountName, 1)

		// Marshal the modified struct back into YAML
		newYamlData, err := yaml.Marshal(&role)
		if err != nil {
			log.Fatalf("Error marshalling YAML: %v", err)
		}

		// Write the modified YAML back to the file
		newClusterRole := "cluster-role-" + serviceAccountName + ".yaml"
		err = os.WriteFile(newClusterRole, newYamlData, 0644)
		if err != nil {
			log.Fatalf("Error writing modified YAML file: %v", err)
		}

		fmt.Println("New cluster role YAML file successfully modified and saved.")
		// Apply the custom cluster role
		cmd = exec.Command("kubectl", "apply", "-f", newClusterRole)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error applying cluster role: %v\n", err)
			os.Exit(1)
		}
	}

	// Bind the cluster role to the service account
	cmd = exec.Command("kubectl", "delete", "clusterrolebinding", serviceAccountName+"-binding")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error deleting old clusterrolebinding: %v\n", err)
	}
	if !deleteSa {
		cmd = exec.Command("kubectl", "create", "clusterrolebinding", serviceAccountName+"-binding", "--clusterrole="+serviceAccountName+"-cluster-role", "--serviceaccount=kube-system:"+serviceAccountName)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error creating cluster role binding: %v\n", err)
			os.Exit(1)
		}
	}

	// Check if CA data is available or not
	if !deleteSa {
		var caDataFlag bool
		var caData string
		cmd := exec.Command("kubectl", "config", "view", "--raw=true", "--minify", "--output=jsonpath={.clusters[*].cluster.certificate-authority-data}")
		caDataOutput, err := cmd.Output()
		if err != nil {
			fmt.Println("Warning: No certificate authority data found, insecure-skip-tls-verify: true will be used")
			caDataFlag = false
		} else {
			caDataFlag = true
			caData = strings.TrimSpace(string(caDataOutput))
		}

		// Generate kubeconfig file content
		var kubeconfig string
		if caDataFlag {
			kubeconfig = fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
contexts:
- context:
    cluster: %s
    namespace: default
    user: %s
  name: %s
current-context: %s
kind: Config
users:
- name: %s
  user:
    token: %s`, caData, clusterServer, currentContext, currentContext, serviceAccountName, currentContext, currentContext, serviceAccountName, serviceAccountToken)
		} else {
			kubeconfig = fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: %s
contexts:
- context:
    cluster: %s
    namespace: default
    user: %s
  name: %s
current-context: %s
kind: Config
users:
- name: %s
  user:
    token: %s`, clusterServer, currentContext, currentContext, serviceAccountName, currentContext, currentContext, serviceAccountName, serviceAccountToken)
		}
		// Save kubeconfig to file
		err = os.WriteFile("kubeconfig_"+serviceAccountName, []byte(kubeconfig), 0644)
		if err != nil {
			fmt.Printf("Error saving kubeconfig file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Service account, service account token, cluster role, cluster role binding, and kubeconfig_%s file created successfully!", serviceAccountName)
	} else {
		fmt.Printf("Service account %s deleted successfully!", serviceAccountName)
	}
}
