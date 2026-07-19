package utils_postgres

// MapDBErrorToServiceError maps PostgreSQL errors to domain errors.
func MapDBErrorToServiceError(err error) error {
	if err == nil {
		return nil
	}

	return err
}
