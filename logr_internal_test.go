package logr

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMakeDestName(t *testing.T) {
	now := time.Now().Truncate(time.Hour * 24)

	// No options
	n := makeDestName("/var/log/main.log", now, &Options{})
	expected := fmt.Sprintf("/var/log/main.log.%s", now.Format(TimeFormat))
	require.Equal(t, expected, n)

	// Time as prefix
	n = makeDestName("/var/log/main.log", now, &Options{
		TimeFormatAsPrefix: true,
	})
	expected = fmt.Sprintf("/var/log/main.%s.log", now.Format(TimeFormat))
	require.Equal(t, expected, n)

	// Custom time format
	n = makeDestName("/var/log/main.log", now, &Options{
		TimeFormat: "2006__01__02",
	})
	expected = fmt.Sprintf("/var/log/main.log.%s", now.Format("2006__01__02"))
	require.Equal(t, expected, n)

	// Time as prefix & custom time format
	n = makeDestName("/var/log/main.log", now, &Options{
		TimeFormatAsPrefix: true,
		TimeFormat:         "2006__01__02",
	})
	expected = fmt.Sprintf("/var/log/main.%s.log", now.Format("2006__01__02"))
	require.Equal(t, expected, n)
}
