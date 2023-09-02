package mail_test

import (
	"camrec/mail"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseEmpty(t *testing.T) {
	require.Empty(t, mail.ParseTimestamp(""))
}

func TestParseNormal(t *testing.T) {
	require.Equal(t, "2023-08-30 22:41:06", mail.ParseTimestamp("C3WN(K49112334) Сигнал обнаружения движения C3WN(K49112334) 2023-08-30 22:41:06 You can view more via EZVIZ APP. Getting too many email alerts? You can disable the alerts by going to Menu&gt;General"))
}

func TestBuildEmpty(t *testing.T) {
	ts, err := mail.BuildTimestamp("")
	require.Error(t, err)
	require.Nil(t, ts)
}

func TestBuildNormal(t *testing.T) {
	ts, err := mail.BuildTimestamp("2023-08-30 22:41:06")

	require.NoError(t, err)
	require.Equal(t, time.Date(2023, 8, 30, 22, 41, 06, 0, time.Local), *ts)
}
