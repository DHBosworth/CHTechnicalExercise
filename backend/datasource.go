package backend

// GameDataSource represents any type which can provide data for the games
// service
type GameDataSource interface {
	Game(id string) (Game, error)
	Report() (Report, error)
}

// ServiceDataSource represents any type which can provide data for the entire
// service.
type ServiceDataSource interface {
	GameDataSource
}
