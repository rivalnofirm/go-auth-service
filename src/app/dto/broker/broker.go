package broker

type AuthBrokerDto struct {
	UserId    int64  `json:"user_id"`
	IpAddress string `json:"ip_adress"`
	Device    string `json:"device"`
	Event     string `json:"event"`
}
