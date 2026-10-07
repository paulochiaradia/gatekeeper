package ports

// AccessUseCase define o contrato para a regra de negócio de acesso.
// Qualquer serviço que implemente ProcessAccessRequest pode ser plugado ao MQTT.
type AccessUseCase interface {
	ProcessAccessRequest(uid string) (bool, error)
}
