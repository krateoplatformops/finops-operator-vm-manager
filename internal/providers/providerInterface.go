package providers

type ProviderInterface interface {
	Connect() error
	SetResourceStatus() error
}
