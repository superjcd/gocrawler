package health

type HealthChecker interface {
	Health() (bool, map[string]any)
}


