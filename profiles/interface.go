package profiles

type Repository interface {
	persist(profiles map[string]Profile) error
}
