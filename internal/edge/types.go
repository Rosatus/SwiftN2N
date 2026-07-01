package edge

type State string

const (
	StateIdle       State = "idle"
	StateStarting   State = "starting"
	StateConnecting State = "connecting"
	StateConnected  State = "connected"
	StateFailed     State = "failed"
	StateStopping   State = "stopping"
)

type Config struct {
	EdgePath            string   `json:"edgePath"`
	AllowCustomEdgePath bool     `json:"allowCustomEdgePath"`
	Supernodes          []string `json:"supernodes"`
	Community           string   `json:"community"`
	Address             string   `json:"address"`
	Key                 string   `json:"key"`
	Cipher              string   `json:"cipher"`
	HeaderEncryption    bool     `json:"headerEncryption"`
	MAC                 string   `json:"mac"`
	DeviceName          string   `json:"deviceName"`
	MTU                 int      `json:"mtu"`
	LocalPort           string   `json:"localPort"`
	ManagementPort      int      `json:"managementPort"`
	Verbose             int      `json:"verbose"`
	AuthUsername        string   `json:"authUsername"`
	AuthPassword        string   `json:"authPassword"`
	FederationPublicKey string   `json:"federationPublicKey"`
	SupernodeOnly       string   `json:"supernodeOnly"`
	Compression         string   `json:"compression"`
	AcceptMulticast     bool     `json:"acceptMulticast"`
	EnableRouting       bool     `json:"enableRouting"`
	Routes              []string `json:"routes"`
	TrafficRules        []string `json:"trafficRules"`
	WindowsMetric       int      `json:"windowsMetric"`
}

type Status struct {
	State     State  `json:"state"`
	Message   string `json:"message"`
	PID       int    `json:"pid"`
	EdgePath  string `json:"edgePath"`
	StartedAt string `json:"startedAt"`
	UpdatedAt string `json:"updatedAt"`
	LastError string `json:"lastError"`
}

type LogEvent struct {
	Time   string `json:"time"`
	Stream string `json:"stream"`
	Line   string `json:"line"`
}

type ExitEvent struct {
	Time    string `json:"time"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type EnvironmentStatus struct {
	OS              string   `json:"os"`
	Arch            string   `json:"arch"`
	Elevated        bool     `json:"elevated"`
	HelperAvailable bool     `json:"helperAvailable"`
	EdgeFound       bool     `json:"edgeFound"`
	EdgePath        string   `json:"edgePath"`
	EdgeVersion     string   `json:"edgeVersion"`
	Missing         []string `json:"missing"`
	Message         string   `json:"message"`
}

type processExit struct {
	code    int
	message string
}
