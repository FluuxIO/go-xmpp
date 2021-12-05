package stanza

import (
	"testing"
	"time"
)

func TestDateToString(t *testing.T) {
	t1 := JabberDate{value: time.Now()}
	t2 := JabberDate{value: time.Now().Add(24 * time.Hour)}

	t1Str := t1.DateToString()
	t2Str := t2.DateToString()

	if t1Str == t2Str {
		t.Fatalf("time representations should not be identical")
	}
}

func TestDateToStringOracle(t *testing.T) {
	expected := "2009-11-10"
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t1 := JabberDate{value: time.Date(2009, time.November, 10, 23, 3, 22, 89, loc)}

	t1Str := t1.DateToString()
	if t1Str != expected {
		t.Fatalf("time is different than expected. Expected: %s, Actual: %s", expected, t1Str)
	}
}

func TestDateTimeToString(t *testing.T) {
	t1 := JabberDate{value: time.Now()}
	t2 := JabberDate{value: time.Now().Add(10 * time.Second)}

	t1Str := t1.DateTimeToString(false)
	t2Str := t2.DateTimeToString(false)

	if t1Str == t2Str {
		t.Fatalf("time representations should not be identical")
	}
}

func TestDateTimeToStringOracle(t *testing.T) {
	expected := "2009-11-10T23:03:22+08:00"
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t1 := JabberDate{value: time.Date(2009, time.November, 10, 23, 3, 22, 89, loc)}

	t1Str := t1.DateTimeToString(false)
	if t1Str != expected {
		t.Fatalf("time is different than expected. Expected: %s, Actual: %s", expected, t1Str)
	}
}

func TestDateTimeToStringNanos(t *testing.T) {
	t1 := JabberDate{value: time.Now()}
	time.After(10 * time.Millisecond)
	t2 := JabberDate{value: time.Now()}

	t1Str := t1.DateTimeToString(true)
	t2Str := t2.DateTimeToString(true)

	if t1Str == t2Str {
		t.Fatalf("time representations should not be identical")
	}
}

func TestDateTimeToStringNanosOracle(t *testing.T) {
	expected := "2009-11-10T23:03:22.000000089+08:00"
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t1 := JabberDate{value: time.Date(2009, time.November, 10, 23, 3, 22, 89, loc)}

	t1Str := t1.DateTimeToString(true)
	if t1Str != expected {
		t.Fatalf("time is different than expected. Expected: %s, Actual: %s", expected, t1Str)
	}
}

func TestTimeToString(t *testing.T) {
	t1 := JabberDate{value: time.Now()}
	t2 := JabberDate{value: time.Now().Add(10 * time.Second)}

	t1Str, err := t1.TimeToString(false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t2Str, err := t2.TimeToString(false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if t1Str == t2Str {
		t.Fatalf("time representations should not be identical")
	}
}

func TestTimeToStringOracle(t *testing.T) {
	expected := "23:03:22+08:00"
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t1 := JabberDate{value: time.Date(2009, time.November, 10, 23, 3, 22, 89, loc)}

	t1Str, err := t1.TimeToString(false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if t1Str != expected {
		t.Fatalf("time is different than expected. Expected: %s, Actual: %s", expected, t1Str)
	}
}

func TestTimeToStringNanos(t *testing.T) {
	t1 := JabberDate{value: time.Now()}
	time.After(10 * time.Millisecond)
	t2 := JabberDate{value: time.Now()}

	t1Str, err := t1.TimeToString(true)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t2Str, err := t2.TimeToString(true)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if t1Str == t2Str {
		t.Fatalf("time representations should not be identical")
	}
}
func TestTimeToStringNanosOracle(t *testing.T) {
	expected := "23:03:22.000000089+08:00"
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t1 := JabberDate{value: time.Date(2009, time.November, 10, 23, 3, 22, 89, loc)}

	t1Str, err := t1.TimeToString(true)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if t1Str != expected {
		t.Fatalf("time is different than expected. Expected: %s, Actual: %s", expected, t1Str)
	}
}

func TestJabberDateParsing(t *testing.T) {
	date := "2009-11-10"
	_, err := NewJabberDateFromString(date)
	if err != nil {
		t.Fatalf(err.Error())
	}

	dateTime := "2009-11-10T23:03:22+08:00"
	_, err = NewJabberDateFromString(dateTime)
	if err != nil {
		t.Fatalf(err.Error())
	}

	dateTimeNanos := "2009-11-10T23:03:22.000000089+08:00"
	_, err = NewJabberDateFromString(dateTimeNanos)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// TODO : fix these. Parsing a time with an offset doesn't work
	//time := "23:03:22+08:00"
	//_, err = NewJabberDateFromString(time)
	//if err != nil {
	//	t.Fatalf(err.Error())
	//}

	//timeNanos := "23:03:22.000000089+08:00"
	//_, err = NewJabberDateFromString(timeNanos)
	//if err != nil {
	//	t.Fatalf(err.Error())
	//}

}
