/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	// "testing"
	"time"
)

const (
	api_test_key        = "QU0G8RSs5aSsoGVir2"
	api_test_secret_key = "IHmT3wcDaI7TBo0WlJaPlJj8JTMtdb5KQrZR"
)

type TimeMock struct {
	mockedTime time.Time
}

func (t *TimeMock) Now() time.Time {
	return t.mockedTime
}
