package beater

import "time"

type WatchEvent struct {
	Type   string `json:"Type"`
	Object Object `json:"Object"`
}

type Labels struct {
	App                            string `json:"app"`
	ControllerRevisionHash         string `json:"controller-revision-hash"`
	Release                        string `json:"release"`
	StatefulsetKubernetesIoPodName string `json:"statefulset.kubernetes.io/pod-name"`
}
type Annotations struct {
	ChecksumConfig     string `json:"checksum/config"`
	OpenshiftIoScc     string `json:"openshift.io/scc"`
	PrometheusIoPath   string `json:"prometheus.io/path"`
	PrometheusIoPort   string `json:"prometheus.io/port"`
	PrometheusIoScrape string `json:"prometheus.io/scrape"`
}
type OwnerReferences struct {
	APIVersion         string `json:"apiVersion"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	UID                string `json:"uid"`
	Controller         bool   `json:"controller"`
	BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
}
type Metadata struct {
	Name              string            `json:"name"`
	GenerateName      string            `json:"generateName"`
	Namespace         string            `json:"namespace"`
	SelfLink          string            `json:"selfLink"`
	UID               string            `json:"uid"`
	ResourceVersion   string            `json:"resourceVersion"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            Labels            `json:"labels"`
	Annotations       Annotations       `json:"annotations"`
	OwnerReferences   []OwnerReferences `json:"ownerReferences"`
}
type PersistentVolumeClaim struct {
	ClaimName string `json:"claimName"`
}
type ConfigMap struct {
	Name        string `json:"name"`
	DefaultMode int    `json:"defaultMode"`
}
type Secret struct {
	SecretName  string `json:"secretName"`
	DefaultMode int    `json:"defaultMode"`
}
type EmptyDir struct {
}
type Volumes struct {
	Name                  string                `json:"name"`
	PersistentVolumeClaim PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
	ConfigMap             ConfigMap             `json:"configMap,omitempty"`
	Secret                Secret                `json:"secret,omitempty"`
	EmptyDir              EmptyDir              `json:"emptyDir,omitempty"`
}
type Resources struct {
}
type VolumeMounts struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}
type Capabilities struct {
	Drop []string `json:"drop"`
}
type SecurityContext struct {
	Capabilities Capabilities `json:"capabilities"`
}
type FieldRef struct {
	APIVersion string `json:"apiVersion"`
	FieldPath  string `json:"fieldPath"`
}
type ValueFrom struct {
	FieldRef FieldRef `json:"fieldRef"`
}
type SecretKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type Env struct {
	Name      string      `json:"name"`
	Value     string      `json:"value,omitempty"`
	ValueFrom interface{} `json:"valueFrom,omitempty"`
}
type InitContainers struct {
	Name                     string          `json:"name"`
	Image                    string          `json:"image"`
	Command                  []string        `json:"command,omitempty"`
	Args                     []string        `json:"args"`
	Resources                Resources       `json:"resources"`
	VolumeMounts             []VolumeMounts  `json:"volumeMounts"`
	TerminationMessagePath   string          `json:"terminationMessagePath"`
	TerminationMessagePolicy string          `json:"terminationMessagePolicy"`
	ImagePullPolicy          string          `json:"imagePullPolicy"`
	SecurityContext          SecurityContext `json:"securityContext"`
	Env                      []Env           `json:"env,omitempty"`
}
type Ports struct {
	Name          string `json:"name"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"`
}
type Exec struct {
	Command []string `json:"command"`
}
type LivenessProbe struct {
	Exec                Exec `json:"exec"`
	InitialDelaySeconds int  `json:"initialDelaySeconds"`
	TimeoutSeconds      int  `json:"timeoutSeconds"`
	PeriodSeconds       int  `json:"periodSeconds"`
	SuccessThreshold    int  `json:"successThreshold"`
	FailureThreshold    int  `json:"failureThreshold"`
}
type ReadinessProbe struct {
	Exec                Exec `json:"exec"`
	InitialDelaySeconds int  `json:"initialDelaySeconds"`
	TimeoutSeconds      int  `json:"timeoutSeconds"`
	PeriodSeconds       int  `json:"periodSeconds"`
	SuccessThreshold    int  `json:"successThreshold"`
	FailureThreshold    int  `json:"failureThreshold"`
}

type Containers struct {
	Name                     string          `json:"name"`
	Image                    string          `json:"image"`
	Command                  []string        `json:"command"`
	Args                     []string        `json:"args,omitempty"`
	Ports                    []Ports         `json:"ports"`
	Resources                Resources       `json:"resources"`
	VolumeMounts             []VolumeMounts  `json:"volumeMounts"`
	LivenessProbe            LivenessProbe   `json:"livenessProbe"`
	ReadinessProbe           ReadinessProbe  `json:"readinessProbe,omitempty"`
	TerminationMessagePath   string          `json:"terminationMessagePath"`
	TerminationMessagePolicy string          `json:"terminationMessagePolicy"`
	ImagePullPolicy          string          `json:"imagePullPolicy"`
	SecurityContext          SecurityContext `json:"securityContext"`
	Env                      []Env           `json:"env,omitempty"`
}
type NodeSelector struct {
	NodeRoleKubernetesIoCompute string `json:"node-role.kubernetes.io/compute"`
}
type SeLinuxOptions struct {
	Level string `json:"level"`
}

type ImagePullSecrets struct {
	Name string `json:"name"`
}
type Spec struct {
	Volumes                       []Volumes          `json:"volumes"`
	InitContainers                []InitContainers   `json:"initContainers"`
	Containers                    []Containers       `json:"containers"`
	RestartPolicy                 string             `json:"restartPolicy"`
	TerminationGracePeriodSeconds int                `json:"terminationGracePeriodSeconds"`
	DNSPolicy                     string             `json:"dnsPolicy"`
	NodeSelector                  NodeSelector       `json:"nodeSelector"`
	ServiceAccountName            string             `json:"serviceAccountName"`
	ServiceAccount                string             `json:"serviceAccount"`
	NodeName                      string             `json:"nodeName"`
	SecurityContext               SecurityContext    `json:"securityContext"`
	ImagePullSecrets              []ImagePullSecrets `json:"imagePullSecrets"`
	Hostname                      string             `json:"hostname"`
	Subdomain                     string             `json:"subdomain"`
	SchedulerName                 string             `json:"schedulerName"`
	Priority                      int                `json:"priority"`
}
type Conditions struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastProbeTime      interface{} `json:"lastProbeTime"`
	LastTransitionTime time.Time   `json:"lastTransitionTime"`
}
type Terminated struct {
	ExitCode    int       `json:"exitCode"`
	Reason      string    `json:"reason"`
	StartedAt   time.Time `json:"startedAt"`
	FinishedAt  time.Time `json:"finishedAt"`
	ContainerID string    `json:"containerID"`
}
type State struct {
	Running    Running    `json:"running"`
	Terminated Terminated `json:"terminated"`
}
type InitContainerStatuses struct {
	Name         string    `json:"name"`
	State        State     `json:"state"`
	LastState    LastState `json:"lastState"`
	Ready        bool      `json:"ready"`
	RestartCount int       `json:"restartCount"`
	Image        string    `json:"image"`
	ImageID      string    `json:"imageID"`
	ContainerID  string    `json:"containerID"`
}
type Running struct {
	StartedAt time.Time `json:"startedAt"`
}

type LastState struct {
	Terminated Terminated `json:"terminated"`
}

type ContainerStatuses struct {
	Name         string    `json:"name"`
	State        State     `json:"state"`
	Ready        bool      `json:"ready"`
	RestartCount int       `json:"restartCount"`
	Image        string    `json:"image"`
	ImageID      string    `json:"imageID"`
	ContainerID  string    `json:"containerID"`
	LastState    LastState `json:"lastState,omitempty"`
}
type Status struct {
	Phase                 string                  `json:"phase"`
	Conditions            []Conditions            `json:"conditions"`
	HostIP                string                  `json:"hostIP"`
	PodIP                 string                  `json:"podIP"`
	StartTime             time.Time               `json:"startTime"`
	InitContainerStatuses []InitContainerStatuses `json:"initContainerStatuses"`
	ContainerStatuses     []ContainerStatuses     `json:"containerStatuses"`
	QosClass              string                  `json:"qosClass"`
}
type Object struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
	Status   Status   `json:"status"`
}
