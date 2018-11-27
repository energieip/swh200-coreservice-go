package network

const (
	//MQTT protocol
	MQTT = "mqtt"

	LogNone     = "none"
	LogDebug    = "debug"
	LogInfo     = "info"
	LogWarning  = "warning"
	LogError    = "error"
	LogCritical = "critical"
)

type networkError struct {
	s string
}

func (e *networkError) Error() string {
	return e.s
}

// NewError raise an error
func NewError(text string) error {
	return &networkError{text}
}

// NetworkInterface network abstraction layer
type NetworkInterface interface {
	Disconnect()

	SendCommand(string, string) error

	Initialize(config NetworkConfig) error
}

//NetworkConfig configuration structure
type NetworkConfig struct {
	IP               string
	Port             string
	ClientName       string
	Callbacks        map[string]func(Client, Message)
	LogLevel         string
	User             string //for authentification
	Password         string
	ServerCertificat string
	ClientCertificat string
	ClientKey        string
}

// NewNetwork instanciate the appropriate networkinterface
func NewNetwork(protocol string) (NetworkInterface, error) {
	switch protocol {
	case MQTT:
		return &MQTTNetwork{}, nil
	default:
		return nil, NewError("Unknow protocol " + protocol)
	}
}
